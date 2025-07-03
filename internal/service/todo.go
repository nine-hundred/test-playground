package service

import (
	"errors"
	"integration-test-example/internal/model"
	"integration-test-example/internal/repository"
)

var (
	ErrTodoNotFound = errors.New("todo not found")
)

type TodoService struct {
	todoRepository *repository.TodoRepository
}

func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{todoRepository: repo}
}

func (s TodoService) CreateTodo(title, description string) (*model.Todo, error) {
	todo := &model.Todo{
		Title:       title,
		Description: description,
		Completed:   false,
	}
	return s.todoRepository.Create(todo)
}

func (s TodoService) GetAllTodos() ([]*model.Todo, error) {
	todos, err := s.todoRepository.GetAll()
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (s TodoService) GetTodoById(id int) (*model.Todo, error) {
	todo, err := s.todoRepository.GetTodo(id)
	if err != nil {
		if err == repository.ErrTodoNotFound {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}
	return todo, nil
}

func (s *TodoService) UpdateTodo(id int, title, description *string, completed *bool) (*model.Todo, error) {
	existingTodo, err := s.todoRepository.GetTodo(id)
	if err != nil {
		return nil, err
	}

	if title != nil {
		existingTodo.Title = *title
	}
	if description != nil {
		existingTodo.Description = *description
	}
	if completed != nil {
		existingTodo.Completed = *completed
	}

	updatedTodo, err := s.todoRepository.Update(existingTodo)
	if err != nil {
		return nil, err
	}

	return updatedTodo, nil
}

func (s *TodoService) DeleteTodo(id int) error {
	err := s.todoRepository.Delete(id)
	if err != nil {
		return err
	}

	return nil
}
