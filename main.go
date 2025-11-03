package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

var db *sql.DB

func ListHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, task, done from todos")
	if err != nil {
		slog.Error("Error querying todos", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.Task, &todo.Done)
		if err != nil {
			slog.Error("Error scanning rows", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		slog.Error("Error iterating rows", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(todos)
	if err != nil {
		slog.Error("Error encoding JSON", "error", err)
		return
	}
}

func ReadHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid ID! ID must be an integer", http.StatusBadRequest)
		return
	}
	var todo Todo
	row := db.QueryRow("SELECT id, task, done FROM todos WHERE id = ?", id)

	err = row.Scan(&todo.ID, &todo.Task, &todo.Done)

	if err == sql.ErrNoRows {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	if err != nil {
		slog.Error("Error querying todo", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(todo)
	if err != nil {
		slog.Error("Error encoding JSON", "error", err)
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

	if data.Task == "" {
		http.Error(w, "Task is empty", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO todos (task, done) VALUES (?, ?)", data.Task, data.Done)
	if err != nil {
		slog.Error("Error inserting todo", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("Error getting last insert ID", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newTask := Todo{
		ID:   int(id),
		Task: data.Task,
		Done: data.Done,
	}

	slog.Info("Added new task", "ID", newTask.ID, "Task", newTask.Task, "Done", newTask.Done)

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

	result, err := db.Exec("UPDATE todos SET task = ?, done = ? WHERE id = ?", data.Task, data.Done, id)
	if err != nil {
		slog.Error("Error updating todo", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Error getting rows affected", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	slog.Info("Updated todo", "ID", data.ID, "Data", data)

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

	result, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		slog.Error("Error deleting todo", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Error getting rows affected", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	slog.Info("Deleted item from todos", "ID", id)

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	var err error
	db, err = sql.Open("mysql", connectionStr)
	if err != nil {
		slog.Error("Failed to connect to DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("Failed to ping DB", "error", err)
		os.Exit(1)
	}
	slog.Info("DB connected")

	createTable := `
CREATE TABLE IF NOT EXISTS todos (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task VARCHAR(255) NOT NULL,
    done BOOLEAN DEFAULT FALSE
)
`
	_, err = db.Exec(createTable)
	if err != nil {
		slog.Error("Failed creating table", "error", err)
		os.Exit(1)
	}
	slog.Info("Table created or already exists")

	fmt.Println("starting server")
	router := mux.NewRouter()

	router.HandleFunc("/todos", ListHandler).Methods("GET")
	router.HandleFunc("/todos/{id}", ReadHandler).Methods("GET")
	router.HandleFunc("/todos", CreateHandler).Methods("POST")
	router.HandleFunc("/todos/{id}", UpdateHandler).Methods("PUT")
	router.HandleFunc("/todos/{id}", DeleteHandler).Methods("DELETE")

	if err = http.ListenAndServe(":5555", router); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
