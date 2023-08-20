package main

import (
	"github.com/witb/indigoo"
	"net/http"
)

func main() {
	mux := indigoo.RenderApplication()

	http.ListenAndServe(":4200", mux)
}
