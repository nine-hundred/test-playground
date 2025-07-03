package repository

import (
	"database/sql"
	"errors"
	"integration-test-example/internal/model"
	"time"
)

type TodoRepository struct {
	db *sql.DB
}

var (
	ErrTodoNotFound = errors.New("todo not found")
)

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{
		db: db,
	}
}

func (r *TodoRepository) Create(todo *model.Todo) (*model.Todo, error) {
	query := `INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now

	result, err := r.db.Exec(
		query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	todo.ID = id
	return todo, nil
}

func (r TodoRepository) GetAll() ([]*model.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*model.Todo
	for rows.Next() {
		todo := &model.Todo{}
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r TodoRepository) GetTodo(id int) (*model.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = ?
	`

	todo := &model.Todo{}
	err := r.db.QueryRow(query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}

	return todo, nil
}

func (r TodoRepository) Update(todo *model.Todo) (*model.Todo, error) {
	query := `
		UPDATE todos
		SET title = ?, description = ?, completed = ?, updated_at = ?
		WHERE id = ?
	`

	todo.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.UpdatedAt,
		todo.ID,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

func (r TodoRepository) Delete(id int) error {
	query := `DELETE FROM todos WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTodoNotFound
	}

	return nil
}
