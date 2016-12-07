package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func maybeServeMetadata() {
	if *file != "" {
		http.HandleFunc("/saml/metadata", func(w http.ResponseWriter, r *http.Request) {
			meta, err := os.Open(*file)
			if err != nil {
				log.Println("Error opening file metadata for http request")
			}
			defer meta.Close()
			io.Copy(w, meta)
		})
	}
}
