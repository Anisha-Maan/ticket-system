package main

import "sync"

type Store struct {
	mu sync.RWMutex

	users   map[int]*User
	tickets map[int]*Ticket

	nextUserID   int
	nextTicketID int
}

func NewStore() *Store {
	return &Store{
		users:        make(map[int]*User),
		tickets:      make(map[int]*Ticket),
		nextUserID:   1,
		nextTicketID: 1,
	}
}

var store = NewStore()