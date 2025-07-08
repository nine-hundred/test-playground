package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"integration-test-example/internal/model"
	"integration-test-example/internal/repository"
	"integration-test-example/pkg/config"
	"integration-test-example/pkg/database"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestTodoIntegration(t *testing.T) {
	ctx := context.Background()

	composeStack, err := compose.NewDockerCompose("../../docker-compose.yml")
	require.NoError(t, err, "Failed to create docker-compose stack")

	err = composeStack.Up(ctx,
		compose.Wait(true),
	)
	require.NoError(t, err, "Failed to start docker-compose stack")

	defer func() {
		err := composeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
		if err != nil {
			t.Logf("Cleanup failed: %v", err)
			t.Fail()
		}
	}()

	baseURL := "http://localhost:8080"

	db, err := database.Connect(getTestDatabaseConfig())
	repo := repository.NewTodoRepository(db)

	t.Log("Waiting for application to be ready...")
	waitForApplication(t, baseURL)

	t.Run("Create Todo", func(t *testing.T) {
		// Arrange
		todoReq := map[string]interface{}{
			"title":       "title",
			"description": "desc",
		}
		reqBody, err := json.Marshal(todoReq)
		require.NoError(t, err)

		// Act
		resp, err := http.Post(baseURL+"/api/v1/todos", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)

		// Assert
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.Equal(t, "Todo created successfully", respBody["message"])
	})

	t.Run("Get All Todos", func(t *testing.T) {
		// Arrange
		dummyTodos := []*model.Todo{
			&model.Todo{
				Title:       "test1 title",
				Description: "test1 desc",
			},
			&model.Todo{
				Title:       "test2 title",
				Description: "test2 title",
			},
		}
		for _, todo := range dummyTodos {
			_, err = repo.Create(todo)
			assert.NoError(t, err)
		}

		// Act
		resp, err := http.Get(baseURL + "/api/v1/todos")
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)

		todoList := respBody["todos"].([]interface{})
		assert.GreaterOrEqual(t, len(todoList), 2)
	})

	t.Run("Get todo", func(t *testing.T) {
		// Arrange
		dummyTodo := &model.Todo{
			Title:       "dummy title",
			Description: "dummy desc",
		}
		dummyTodo, err = repo.Create(dummyTodo)
		assert.NoError(t, err)

		// Act
		resp, err := http.Get(baseURL + "/api/v1/todos/" + strconv.Itoa(int(dummyTodo.ID)))
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		type TodoResponse struct {
			Todo model.Todo `json:"todo"`
		}
		var todoResp TodoResponse
		err = json.Unmarshal(body, &todoResp)
		assert.NoError(t, err)
		assert.Equal(t, dummyTodo.ID, todoResp.Todo.ID)
	})

	t.Run("Update todo", func(t *testing.T) {
		// Arrange
		dummyTodo := &model.Todo{
			Title:       "dummy title",
			Description: "dummy desc",
		}
		dummyTodo, err = repo.Create(dummyTodo)
		assert.NoError(t, err)

		updatedTitle := "updated title"
		updateReq := map[string]interface{}{
			"title": updatedTitle,
		}
		reqBody, err := json.Marshal(updateReq)
		assert.NoError(t, err)

		// Act
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPut, baseURL+"/api/v1/todos/"+strconv.Itoa(int(dummyTodo.ID)), bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		defer resp.Body.Close()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		type UpdateResponse struct {
			Message string     `json:"message"`
			Todo    model.Todo `json:"todo"`
		}
		var todoResp UpdateResponse
		err = json.Unmarshal(body, &todoResp)
		assert.NoError(t, err)
		assert.Equal(t, dummyTodo.ID, todoResp.Todo.ID)
		assert.Equal(t, todoResp.Todo.Title, updatedTitle)
	})

	t.Run("Delete Todo", func(t *testing.T) {
		// Arrange
		dummyTodo := &model.Todo{
			Title:       "dummy title",
			Description: "dummy desc",
		}
		dummyTodo, err = repo.Create(dummyTodo)
		assert.NoError(t, err)

		// Act
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodDelete, baseURL+"/api/v1/todos/"+strconv.Itoa(int(dummyTodo.ID)), nil)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		defer resp.Body.Close()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_, err = repo.GetTodo(int(dummyTodo.ID))
		assert.Equal(t, err, repository.ErrTodoNotFound)
	})

	t.Run("Get Not Exist Todo Should Return 404", func(t *testing.T) {
		// Act
		resp, err := http.Get(baseURL + "/api/v1/todos/" + strconv.Itoa(int(time.Now().Unix())))
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("Invalid Request", func(t *testing.T) {
		t.Run("Create Todo with empty title", func(t *testing.T) {
			// Arrange: 잘못된 요청 데이터 (제목 없음)
			invalidReq := map[string]interface{}{
				"title":       "", // 빈 제목
				"description": "설명만 있음",
			}
			reqBody, _ := json.Marshal(invalidReq)

			// Act
			resp, err := http.Post(baseURL+"/api/v1/todos", "application/json", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Assert
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
		// 다양한 에러 케이스 검증 추가...
	})
}

func waitForApplication(t *testing.T, baseURL string) {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Log("✅ Application is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		t.Logf("Waiting for application... attempt %d/%d", i+1, maxAttempts)
		time.Sleep(2 * time.Second)
	}

	t.Fatal("Application failed to start within timeout")
}

// getTestDatabaseConfig: Test용 DB 설정 테스트 픽스쳐
func getTestDatabaseConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "todouser",
		Password: "password",
		Name:     "todoapp",
	}
}
