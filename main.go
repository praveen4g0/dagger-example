package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Welcome to Red Hat \n"))
	})

	http.ListenAndServe(":9090", nil)
}
