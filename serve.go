// Copyright 2016 Andrew Stuart

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
