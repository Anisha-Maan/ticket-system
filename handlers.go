package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for _, user := range store.users {
		if user.Email == req.Email {
			http.Error(w, "email already registered", http.StatusBadRequest)
			return
		}
	}

	salt := generateSalt()

	user := &User{
		ID:           store.nextUserID,
		Email:        req.Email,
		Salt:         salt,
		PasswordHash: hashPassword(req.Password, salt),
	}

	store.users[user.ID] = user
	store.nextUserID++

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "user registered successfully",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	var foundUser *User

	for _, user := range store.users {
		if user.Email == req.Email {
			foundUser = user
			break
		}
	}

	if foundUser == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !verifyPassword(
		req.Password,
		foundUser.Salt,
		foundUser.PasswordHash,
	) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := GenerateJWT(foundUser.ID)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func TicketsHandler(w http.ResponseWriter, r *http.Request) {

	userID := GetUserID(r)

	switch r.Method {

	case http.MethodPost:

		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}

		store.mu.Lock()
		defer store.mu.Unlock()

		ticket := &Ticket{
			ID:          store.nextTicketID,
			Title:       req.Title,
			Description: req.Description,
			Status:      "open",
			UserID:      userID,
		}

		store.tickets[ticket.ID] = ticket
		store.nextTicketID++

		writeJSON(w, http.StatusCreated, ticket)

	case http.MethodGet:

		store.mu.RLock()
		defer store.mu.RUnlock()

		userTickets := make([]*Ticket, 0)

		for _, ticket := range store.tickets {
			if ticket.UserID == userID {
				userTickets = append(userTickets, ticket)
			}
		}

		writeJSON(w, http.StatusOK, userTickets)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func TicketHandler(w http.ResponseWriter, r *http.Request) {

	userID := GetUserID(r)

	path := strings.TrimPrefix(r.URL.Path, "/tickets/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	ticketID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid ticket id", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	ticket, exists := store.tickets[ticketID]

	if !exists {
		http.Error(w, "ticket not found", http.StatusNotFound)
		return
	}

	if ticket.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// GET /tickets/{id}
	if len(parts) == 1 && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, ticket)
		return
	}

	// PATCH /tickets/{id}/status
	if len(parts) == 2 &&
		parts[1] == "status" &&
		r.Method == http.MethodPatch {

		var req struct {
			Status string `json:"status"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if !IsValidStatus(req.Status) {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}

		if !CanTransition(ticket.Status, req.Status) {
			http.Error(w, "invalid status transition", http.StatusBadRequest)
			return
		}

		ticket.Status = req.Status

		writeJSON(w, http.StatusOK, ticket)
		return
	}

	http.Error(w, "route not found", http.StatusNotFound)
}