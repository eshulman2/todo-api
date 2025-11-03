package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func TestMain(m *testing.M) {
	// Setup: Connect to test database
	setupTestDB()

	// Run all tests
	code := m.Run()

	// Teardown: Close database
	db.Close()

	// Exit with test result code
	os.Exit(code)
}

func setupTestDB() {
	var err error
	// Connect to TEST database
	db, err = sql.Open("mysql", "root:mypassword@tcp(127.0.0.1:3306)/todo_db_test")
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create table
	createTable := `
    CREATE TABLE IF NOT EXISTS todos (
        id INT AUTO_INCREMENT PRIMARY KEY,
        task VARCHAR(255) NOT NULL,
        done BOOLEAN DEFAULT FALSE
    )
    `
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTodos(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM todos")
	if err != nil {
		t.Fatalf("Failed to clear todos: %v", err)
	}
}

func seedTodo(t *testing.T, task string, done bool) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO todos (task, done) VALUES (?, ?)", task, done)
	if err != nil {
		t.Fatalf("Failed to seed todo: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/todos", ListHandler).Methods("GET")
	router.HandleFunc("/todos/{id}", ReadHandler).Methods("GET")
	router.HandleFunc("/todos", CreateHandler).Methods("POST")
	router.HandleFunc("/todos/{id}", UpdateHandler).Methods("PUT")
	router.HandleFunc("/todos/{id}", DeleteHandler).Methods("DELETE")
	return router
}

func TestListHandler(t *testing.T) {
	clearTodos(t)

	seedTodo(t, "some task", false)
	seedTodo(t, "another task", true)

	router := setupRouter()

	req := httptest.NewRequest("GET", "/todos", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var todos []Todo
	err := json.Unmarshal(rr.Body.Bytes(), &todos)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(todos))
	}

	if todos[0].Task != "some task" {
		t.Errorf("Expected task 'some task', got '%s'", todos[0].Task)
	}
	if todos[0].Done != false {
		t.Errorf("Expected done=false, got %v", todos[0].Done)
	}

}

func TestReadHandler(t *testing.T) {
	clearTodos(t)
	id := seedTodo(t, "some task", false)

	router := setupRouter()

	req := httptest.NewRequest("GET", "/todos/"+strconv.Itoa(id), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var todo Todo
	err := json.Unmarshal(rr.Body.Bytes(), &todo)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if todo.Task != "some task" {
		t.Errorf("Expected task 'some task', got '%s'", todo.Task)
	}
	if todo.Done != false {
		t.Errorf("Expected done=false, got %v", todo.Done)
	}
}

func TestCreateHandler(t *testing.T) {
	clearTodos(t)

	router := setupRouter()

	body := strings.NewReader(`{"task":"New task","done":false}`)
	req := httptest.NewRequest("POST", "/todos", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", status)
	}

	var created Todo
	err := json.Unmarshal(rr.Body.Bytes(), &created)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if created.Task != "New task" {
		t.Errorf("Expected 'New task', got '%s'", created.Task)
	}

	if created.Done != false {
		t.Errorf("Expected false, got '%v'", created.Done)

	}
}

func TestUpdateHandler(t *testing.T) {
	clearTodos(t)
	id := seedTodo(t, "some task", false)

	router := setupRouter()

	body := strings.NewReader(fmt.Sprintf(`{"id": %d,"task":"New task","done":false}`, id))
	req := httptest.NewRequest("PUT", "/todos/"+strconv.Itoa(id), body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}

	var updated Todo
	err := json.Unmarshal(rr.Body.Bytes(), &updated)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if updated.Task != "New task" {
		t.Errorf("Expected 'New task', got '%s'", updated.Task)
	}

	if updated.Done != false {
		t.Errorf("Expected false, got '%v'", updated.Done)

	}
}

func TestDeleteHandler(t *testing.T) {
	clearTodos(t)
	id := seedTodo(t, "some task", false)

	router := setupRouter()

	req := httptest.NewRequest("DELETE", "/todos/"+strconv.Itoa(id), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", status)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM todos WHERE id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected todo to be deleted, but it still exists")
	}
}
