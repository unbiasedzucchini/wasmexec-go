package main

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const inputOffset = 0x10000

func Execute(ctx context.Context, wasmBytes []byte, input []byte) ([]byte, error) {
	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	mod, err := rt.Instantiate(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("instantiate: %w", err)
	}

	run := mod.ExportedFunction("run")
	if run == nil {
		return nil, fmt.Errorf("module does not export 'run'")
	}

	mem := mod.Memory()
	if mem == nil {
		return nil, fmt.Errorf("module does not export memory")
	}

	// Grow memory if needed to fit input at inputOffset
	needed := uint32(inputOffset + len(input))
	if needed > mem.Size() {
		pages := (needed - mem.Size() + 65535) / 65536
		if _, ok := mem.Grow(pages); !ok {
			return nil, fmt.Errorf("failed to grow memory")
		}
	}

	if !mem.Write(inputOffset, input) {
		return nil, fmt.Errorf("failed to write input to memory")
	}

	results, err := run.Call(ctx, api.EncodeI32(inputOffset), api.EncodeI32(int32(len(input))))
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}

	outPtr := uint32(results[0])

	// Read output length (4 bytes LE) at outPtr
	lenBytes, ok := mem.Read(outPtr, 4)
	if !ok {
		return nil, fmt.Errorf("failed to read output length at %d", outPtr)
	}
	outLen := binary.LittleEndian.Uint32(lenBytes)

	output, ok := mem.Read(outPtr+4, outLen)
	if !ok {
		return nil, fmt.Errorf("failed to read output (%d bytes at %d)", outLen, outPtr+4)
	}

	return output, nil
}
