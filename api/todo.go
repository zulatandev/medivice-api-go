package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Todo struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	CreateDate time.Time `json:"createDate"`
	Completed  bool      `json:"completed"`
}

var rdb *redis.Client

const todosKey = "todos"

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "first-caribou-42781.upstash.io:6379",
		Password: "AacdAAIjcDFjYWQ5Y2RhNzJlOWQ0MGQ4YTllYTUzZjY5NmZjODJjZXAxMA",
		DB:       0,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetTodos(w, r)
	case "POST":
		handleCreateTodo(w, r)
	case "PUT":
		handleUpdateTodo(w, r)
	case "DELETE":
		handleDeleteTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	todos, err := rdb.LRange(ctx, todosKey, 0, -1).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var todoList []Todo
	for _, todoStr := range todos {
		var todo Todo
		err := json.Unmarshal([]byte(todoStr), &todo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todoList = append(todoList, todo)
	}

	json.NewEncoder(w).Encode(todoList)
}

func handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	nextID, err := rdb.Incr(ctx, "todo_id").Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	todo.ID = int(nextID)

	todo.CreateDate = time.Now()

	todoBytes, err := json.Marshal(todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rdb.RPush(ctx, todosKey, todoBytes).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedTodo Todo
	err = json.NewDecoder(r.Body).Decode(&updatedTodo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	todos, err := rdb.LRange(ctx, todosKey, 0, -1).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, todoStr := range todos {
		var todo Todo
		err := json.Unmarshal([]byte(todoStr), &todo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if todo.ID == id {
			todo.Title = updatedTodo.Title
			todo.Completed = updatedTodo.Completed

			updatedTodoBytes, err := json.Marshal(todo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = rdb.LSet(ctx, todosKey, int64(i), updatedTodoBytes).Err()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(todo)
			return
		}
	}

	http.Error(w, "Todo not found", http.StatusNotFound)
}

func handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	_, err = rdb.LRem(ctx, todosKey, 0, int64(id)).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
