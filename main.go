package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthHandler)

	mux.HandleFunc("/auth/register", RegisterHandler)

	mux.HandleFunc("/auth/login", LoginHandler)

	mux.Handle(
		"/tickets",
		AuthMiddleware(
			http.HandlerFunc(TicketsHandler),
		),
	)

	mux.Handle(
		"/tickets/",
		AuthMiddleware(
			http.HandlerFunc(TicketHandler),
		),
	)

	log.Println("Server running on :8080")

	log.Fatal(
		http.ListenAndServe(":8080", mux),
	)
}