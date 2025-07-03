package integration

import (
	"context"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"net/http"
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
		require.NoError(t, err, "Failed to cleanup docker-compose stack")
	}()

	baseURL := "http://localhost:8080"

	t.Log("Waiting for application to be ready...")
	waitForApplication(t, baseURL)

	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Log("Health check passed")
	})

	t.Run("Create Todo", func(t *testing.T) {

	})

	t.Run("Get All Todos", func(t *testing.T) {

	})

	t.Run("Get todo", func(t *testing.T) {

	})

	t.Run("Update todo", func(t *testing.T) {

	})

	t.Run("Delete Todo", func(t *testing.T) {

	})

	t.Run("Get Deleted Todo Should Return 404", func(t *testing.T) {

	})

	t.Run("Invalid Request", func(t *testing.T) {

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
