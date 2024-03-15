package main

import (
	"net/http"
)

func (app *Config) routes() http.Handler {

	mux := http.NewServeMux()

	// http.HandleFunc("/", app.Ping)

	mux.HandleFunc("/", app.Ping)
	
	mux.HandleFunc("/auth", app.Auth) 
	mux.HandleFunc("/refresh", app.Refresh)


	return mux
}



// {"auth_token":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ0ZXN0X3VzZXJfaWRkIn0.i5Via0QLMdY-w66I9hVW8HTI-gXp429Bh1xukq9FOAxg8NohTitsfNNYpGFnAVb-7rUZbp46xSKgDx9iF9faTg","refresh_token":"RlB0dmdOU3NUTEEyTzFGR0JaWGI="}