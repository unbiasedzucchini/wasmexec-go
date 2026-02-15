# wasmexec-go

HTTP server for content-addressable blob storage with WebAssembly execution.

## API

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/blobs` | Upload a blob. Returns its SHA-256 hash. |
| `GET` | `/blobs/:hash` | Retrieve a blob by hash. |
| `POST` | `/execute/:hash` | Execute a wasm blob. Request body = input, response body = output. |

Blobs are content-addressable and immutable.

## Wasm Contract

Modules must export:
- `memory` — the module's linear memory
- `run(input_ptr: i32, input_len: i32) -> i32` — entry point

The host writes input bytes into the module's memory at offset `0x10000`.
The host then calls `run(0x10000, input_len)`.

`run` returns a pointer to the output, formatted as:
```
[output_len: u32 LE][output_bytes...]
```

No WASI. No imported functions. Pure computation.

## Quick Start

```bash
go run .
curl -X PUT --data-binary @testdata/echo.wasm http://localhost:8000/blobs
curl -X POST --data 'hello' http://localhost:8000/execute/<hash>
```

## Test

```bash
go test -v ./...
```
