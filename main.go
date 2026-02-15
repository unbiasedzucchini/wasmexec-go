package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, store *BlobStore) {
	mux.HandleFunc("/blobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hash, err := store.Put(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"hash": hash})
	})

	mux.HandleFunc("/blobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hash := strings.TrimPrefix(r.URL.Path, "/blobs/")
		if hash == "" {
			http.Error(w, "hash required", http.StatusBadRequest)
			return
		}
		data, err := store.Get(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if data == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(data)
	})

	mux.HandleFunc("/execute/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hash := strings.TrimPrefix(r.URL.Path, "/execute/")
		if hash == "" {
			http.Error(w, "hash required", http.StatusBadRequest)
			return
		}
		wasmBytes, err := store.Get(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if wasmBytes == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		input, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		output, err := Execute(r.Context(), wasmBytes, input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(output)
	})
}

func main() {
	store, err := NewBlobStore("blobs.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	mux := http.NewServeMux()
	RegisterRoutes(mux, store)

	log.Println("listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
