package handler

import (
	"github.com/gin-gonic/gin"
	"integration-test-example/internal/model"
	"integration-test-example/internal/service"
	"net/http"
	"strconv"
)

type TodoHandler struct {
	todoService *service.TodoService
}

func NewTodoHandler(todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

func (h TodoHandler) CreateTodo(c *gin.Context) {
	var req model.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	todo, err := h.todoService.CreateTodo(req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Todo created successfully",
		"todo":    todo,
	})
}

func (h TodoHandler) GetTodos(c *gin.Context) {
	todos, err := h.todoService.GetAllTodos()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Fail to get todos",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"todos": todos,
	})
}

func (h TodoHandler) GetTodo(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid todo Id",
		})
	}

	todo, err := h.todoService.GetTodoById(id)
	if err != nil {
		if err == service.ErrTodoNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Todo not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Fail to get todo",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"todo": todo,
	})
}

func (h TodoHandler) UpdateTodo(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid todo Id",
		})
	}

	var req model.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	todo, err := h.todoService.UpdateTodo(id, req.Title, req.Description, req.Completed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Fail to update todo",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo updated successfully",
		"todo":    todo,
	})
}

func (h TodoHandler) DeleteTodo(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid todo Id",
		})
	}

	err = h.todoService.DeleteTodo(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Fail to delete todo",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Todo deleted successfully",
	})
}
