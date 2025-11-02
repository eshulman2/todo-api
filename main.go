package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/mux"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

var todos = []Todo{
	{ID: 1, Task: "Learn Go basics", Done: true},
	{ID: 2, Task: "Build a simple API", Done: false},
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(todos)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func ReadHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid ID! ID must be an integer", http.StatusBadRequest)
		return
	}

	var found *Todo
	for i := range todos {
		if todos[i].ID == id {
			found = &todos[i]
			break
		}
	}
	if found == nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(found)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
		return
	}
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	var data Todo
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Task == ""{
		http.Error(w, "Task is empty", http.StatusBadRequest)
		return
	}

	var maxID int
	for _, todo := range todos {
		if maxID < todo.ID {
			maxID = todo.ID
		}
	}

	newTask := Todo{
		ID:   maxID + 1,
		Task: data.Task,
		Done: data.Done,
	}

	log.Printf("Added new task %d: %s", newTask.ID, newTask.Task)

	todos = append(todos, newTask)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)

}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid ID! ID must be an integer", http.StatusBadRequest)
		return
	}

	var data Todo
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if id != data.ID {
		http.Error(w, "Id in url doesn't match the id in the body", http.StatusConflict)
		return
	}

	idx := slices.IndexFunc(todos, func(t Todo) bool {
		return t.ID == id
	})

	if idx == -1 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	todos[idx] = data
	log.Printf("Updated todo id %d to: %v", data.ID, data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid ID! ID must be an integer", http.StatusBadRequest)
		return
	}

	idx := slices.IndexFunc(todos, func(t Todo) bool {
		return t.ID == id
	})

	if idx == -1 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	todos = slices.Delete(todos, idx, idx+1)

	log.Printf("Deleted item %d from todos, new todos: %v", idx, todos)

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	fmt.Println("starting server")
	router := mux.NewRouter()

	router.HandleFunc("/todos", ListHandler).Methods("GET")
	router.HandleFunc("/todos/{id}", ReadHandler).Methods("GET")
	router.HandleFunc("/todos", CreateHandler).Methods("POST")
	router.HandleFunc("/todos/{id}", UpdateHandler).Methods("PUT")
	router.HandleFunc("/todos/{id}", DeleteHandler).Methods("DELETE")

	err := http.ListenAndServe(":5555", router)
	if err != nil {
		log.Fatal(err)
	}
}
