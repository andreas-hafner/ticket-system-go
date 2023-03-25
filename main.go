package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Comment struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Text   string `json:"text"`
}

type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Creator     string    `json:"creator"`
	Approver    string    `json:"approver,omitempty"`
	Comments    []Comment `json:"comments,omitempty"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var users = []User{
	{Username: "Andy", Password: "123456"},
	{Username: "Harald", Password: "123456"},
}

var tickets []Ticket
var ticketMutex sync.Mutex
var ticketIDCounter int

func main() {
	loadTickets()

	// Initialisiere den ticketIDCounter mit der aktuellen Länge der Tickets
	ticketIDCounter = len(tickets)

	r := mux.NewRouter()
	r.HandleFunc("/api/login", login).Methods("POST")
	r.HandleFunc("/api/tickets", createTicket).Methods("POST")
	r.HandleFunc("/api/tickets", getTickets).Methods("GET")
	r.HandleFunc("/api/next-ticket-id", getNextTicketID).Methods("GET")
	r.HandleFunc("/api/tickets/{id}", updateTicket).Methods("PUT")
	r.HandleFunc("/api/tickets/{id}/comments", addComment).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./client/")))

	http.ListenAndServe(":8080", r)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	for _, u := range users {
		if user.Username == u.Username && user.Password == u.Password {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			return
		}
	}

	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password"})
}

func createTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newTicket Ticket
	err := json.NewDecoder(r.Body).Decode(&newTicket)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	ticketMutex.Lock()
	defer ticketMutex.Unlock()

	// Inkrementiere den ticketIDCounter und setze die ID des neuen Tickets
	ticketIDCounter++
	newTicket.ID = strconv.Itoa(ticketIDCounter)
	tickets = append(tickets, newTicket)
	saveTickets()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTicket)
}

func getTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ticketMutex.Lock()
	defer ticketMutex.Unlock()

	json.NewEncoder(w).Encode(tickets)
}

func updateTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	ticketID := vars["id"]

	var updatedTicket Ticket
	err := json.NewDecoder(r.Body).Decode(&updatedTicket)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	ticketMutex.Lock()
	defer ticketMutex.Unlock()

	found := false
	for i, ticket := range tickets {
		if ticket.ID == ticketID {
			tickets[i].Approver = updatedTicket.Approver
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ticket not found"})
		return
	}

	saveTickets()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTicket)
}

func addComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	ticketID := vars["id"]

	var newComment Comment
	err := json.NewDecoder(r.Body).Decode(&newComment)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	ticketMutex.Lock()
	defer ticketMutex.Unlock()

	found := false
	for i, ticket := range tickets {
		if ticket.ID ==
			ticketID {
			newComment.ID = strconv.Itoa(len(tickets[i].Comments) + 1)
			tickets[i].Comments = append(tickets[i].Comments, newComment)
			found = true
			break
		}
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ticket not found"})
		return
	}

	saveTickets()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newComment)
}

func loadTickets() {
	ticketFile, err := os.Open("tickets.json")
	if err != nil {
		return
	}
	defer ticketFile.Close()
	bytes, err := io.ReadAll(ticketFile)
	if err != nil {
		return
	}

	json.Unmarshal(bytes, &tickets)
}

func saveTickets() {
	bytes, err := json.Marshal(tickets)
	if err != nil {
		return
	}
	err = os.WriteFile("tickets.json", bytes, 0644)
	if err != nil {
		return
	}
}

func getNextTicketID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ticketMutex.Lock()
	defer ticketMutex.Unlock()

	nextID := ticketIDCounter + 1
	json.NewEncoder(w).Encode(map[string]int{"nextID": nextID})
}
