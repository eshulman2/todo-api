# Todo API

A simple REST API built with Go as a first learning project. Implements basic CRUD operations (Create, Read, Update, Delete, List) for managing todo items using the Gorilla Mux router.

## Tech Stack

- Go
- [gorilla/mux](https://github.com/gorilla/mux) for routing
- In-memory storage

## Running the Server

```bash
go run main.go
```

Server starts on `http://localhost:5555`

## API Endpoints

- `GET /todos` - List all todos
- `GET /todos/{id}` - Get a specific todo
- `POST /todos` - Create a new todo
- `PUT /todos/{id}` - Update a todo
- `DELETE /todos/{id}` - Delete a todo

## Example Usage

```bash
# List all todos
curl http://localhost:5555/todos

# Create a new todo
curl -X POST http://localhost:5555/todos \
  -H "Content-Type: application/json" \
  -d '{"task": "Learn Go", "done": false}'

# Get a specific todo
curl http://localhost:5555/todos/1

# Update a todo
curl -X PUT http://localhost:5555/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"id": 1, "task": "Learn Go", "done": true}'

# Delete a todo
curl -X DELETE http://localhost:5555/todos/1
```
