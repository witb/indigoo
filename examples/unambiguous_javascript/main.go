package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/witb/indigoo"
	"log"
	"net/http"
)

func main() {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux = indigoo.RenderApplicationWithMux(mux)

	err := http.ListenAndServe(":4200", mux)
	if err != nil {
		log.Println(err)
	}

	log.Println("Server started on port 4200")
}
