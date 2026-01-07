-- name: GetTodo :one
SELECT id, title, description, completed, created_at, updated_at
FROM todos
WHERE id = ? LIMIT 1;

-- name: ListTodos :many
SELECT id, title, description, completed, created_at, updated_at
FROM todos
ORDER BY created_at DESC;

-- name: ListTodosByStatus :many
SELECT id, title, description, completed, created_at, updated_at
FROM todos
WHERE completed = ?
ORDER BY created_at DESC;

-- name: CreateTodo :one
INSERT INTO todos (title, description, completed)
VALUES (?, ?, ?)
RETURNING id, title, description, completed, created_at, updated_at;

-- name: UpdateTodo :one
UPDATE todos
SET title = ?, description = ?, completed = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, title, description, completed, created_at, updated_at;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = ?;

-- name: ToggleTodoCompleted :one
UPDATE todos
SET completed = CASE WHEN completed = 0 THEN 1 ELSE 0 END, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, title, description, completed, created_at, updated_at;
