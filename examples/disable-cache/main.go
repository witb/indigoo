package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/witb/indigoo"
	"net/http"
)

func main() {
	indigoo.Cache = false

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux = indigoo.RenderApplicationWithMux(mux)

	http.ListenAndServe(":4200", mux)
}
