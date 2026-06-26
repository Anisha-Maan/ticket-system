# Ticket System (Golang)

## Features
- JWT authentication (no external libraries)
- Ownership-based ticket system
- Status workflow: open → in_progress → closed
- Dockerized deployment
- In-memory store

## Run locally

go run .

## Run with Docker

docker build -t ticket-system .
docker run -p 8080:8080 ticket-system

## API Endpoints

POST   /auth/register
POST   /auth/login
POST   /tickets
GET    /tickets
GET    /tickets/{id}
PATCH  /tickets/{id}/status
GET    /health

## Deployment URLs

**Deployment URL:** `[https://ticket-system-hbbp.onrender.com]`  
**Public Health Check URL:** `https://ticket-system-hbbp.onrender.com/health`

## Assumptions & Notes
- Uses in-memory storage (data resets on restart, as permitted by scope).
- JWT implemented using HMAC SHA256 (no external libraries).
- Passwords are hashed with SHA256 and salted.
- Supported Ticket Status transitions strictly follow: `open -> in_progress -> closed`. A `closed` ticket cannot be reopened.
