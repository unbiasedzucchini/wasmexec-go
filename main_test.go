package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupServer(t *testing.T) http.Handler {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewBlobStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	mux := http.NewServeMux()
	RegisterRoutes(mux, store)
	return mux
}

func TestBlobPutGet(t *testing.T) {
	srv := httptest.NewServer(setupServer(t))
	defer srv.Close()

	body := []byte("hello world")

	// PUT blob
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/blobs", bytes.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	hash := result["hash"]
	if hash == "" {
		t.Fatal("expected hash in response")
	}

	// GET blob
	resp2, err := http.Get(srv.URL + "/blobs/" + hash)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}
	got, _ := io.ReadAll(resp2.Body)
	if !bytes.Equal(got, body) {
		t.Fatalf("expected %q, got %q", body, got)
	}
}

func TestBlobNotFound(t *testing.T) {
	srv := httptest.NewServer(setupServer(t))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/blobs/deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestExecuteEcho(t *testing.T) {
	srv := httptest.NewServer(setupServer(t))
	defer srv.Close()

	// Load echo.wasm
	wasmBytes, err := os.ReadFile("testdata/echo.wasm")
	if err != nil {
		t.Fatal(err)
	}

	// Upload it
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/blobs", bytes.NewReader(wasmBytes))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	hash := result["hash"]

	// Execute with input
	input := []byte("ping")
	resp2, err := http.Post(srv.URL+"/execute/"+hash, "application/octet-stream", bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 200, got %d: %s", resp2.StatusCode, body)
	}
	got, _ := io.ReadAll(resp2.Body)
	if !bytes.Equal(got, input) {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestExecuteEmptyInput(t *testing.T) {
	srv := httptest.NewServer(setupServer(t))
	defer srv.Close()

	wasmBytes, err := os.ReadFile("testdata/echo.wasm")
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/blobs", bytes.NewReader(wasmBytes))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	hash := result["hash"]

	resp2, err := http.Post(srv.URL+"/execute/"+hash, "application/octet-stream", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 200, got %d: %s", resp2.StatusCode, body)
	}
	got, _ := io.ReadAll(resp2.Body)
	if len(got) != 0 {
		t.Fatalf("expected empty output, got %q", got)
	}
}
