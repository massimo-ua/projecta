package web

import (
	"github.com/gorilla/mux"
	"net/http"
)

func createSwaggerHandler(r *mux.Router) {
	r.PathPrefix("/swagger").Handler(http.StripPrefix("/swagger", http.FileServer(http.Dir("swagger"))))
}
