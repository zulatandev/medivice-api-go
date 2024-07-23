package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type Todo struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	CreateDate time.Time `json:"createDate"`
	Completed  bool      `json:"completed"`
}

var todos []Todo

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(todos)
	case "POST":
		var todo Todo
		err := json.NewDecoder(r.Body).Decode(&todo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		todo.ID = len(todos) + 1
		todo.CreateDate = time.Now()
		todos = append(todos, todo)
		json.NewEncoder(w).Encode(todo)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
