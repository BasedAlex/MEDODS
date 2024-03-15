package main

import (
	"net/http"
)

func (app *Config) routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.Ping)
	
	mux.HandleFunc("/auth", app.Auth) 
	mux.HandleFunc("/refresh", app.Refresh)


	return mux
}
