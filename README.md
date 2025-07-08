# test-playground

## 개요
Go + Gin으로 구현한 Todo API와 Testcontainers + LocalStack을 활용한 통합테스트 예제 프로젝트입니다.

## 시작 배경
블라디미르 코리코프의 "단위 테스트" 책을 읽으면서 깨달은 것들을 정리하고, 직접 구현해보는 프로젝트입니다.
기존에 작성하던 테스트의 문제점들을 파악하고, 더 나은 테스트 작성법을 실험해보고 싶었습니다.
Mock 대신 실제 의존성(TestContainers, LocalStack..)을 사용하는 통합테스트에 집중했습니다.

## 테스트 스위트 설계
### 결과 중심 테스트
구현 세부사항이 아닌 최종 결과를 검증하는 것을 목표로 했습니다. 코드가 "어떻게" 동작하는지가 아니라 "무엇을" 달성하는지에 집중했습니다.
이를 위해 Testcontainers를 사용해 실제 의존성(MySQL, Redis, SQS)과 통합된 결과를 검증합니다. 
이렇게 하면 리팩토링 시에도 테스트가 깨지지 않고, 실제 사용자 관점에서의 동작을 보장할 수 있습니다.

### 고품질 테스트만 유지
의미없는 테스트나 불안정한 테스트는 오히려 개발 속도를 늦추고 신뢰도를 떨어뜨립니다. 따라서 각 테스트가 실제 비즈니스 가치를 검증하는지, 안정적으로 동작하는지를 중요하게 고려해야 합니다.

### 개발 주기 통합
테스트는 개발 주기에 자연스럽게 통합되어야 합니다. 이를 위해 GitHub-Actions에 테스트를 포함시켜 모든 Pull Request에서 자동으로 실행되도록 구성했습니다.

### 테스트 구현 예시
```go
func TestTodoIntegration(t *testing.T) {
    // Docker Compose로 실제 환경 구성
    composeStack, err := compose.NewDockerCompose("../../docker-compose.yml")
    err = composeStack.Up(ctx, compose.Wait(true))
    
    // 실제 데이터베이스 연결 (Repository 검증용)
    db, err := database.Connect(getTestDatabaseConfig())
    repo := repository.NewTodoRepository(db)

    // AAA패턴으로 API 테스트 + Repository 검증
    t.Run("Create Todo", func (t *testing.T) {
        // Arrange: 테스트 데이터 준비
        todoReq := map[string]interface{}{
            "title":       "dummy title",
            "description": "desc",
        }
        reqBody, err := json.Marshal(todoReq)
        assert.NoError(t, err)
        
		// Act: 실제 동작 수행
        resp, err := http.Post(baseURL+"/api/v1/todos",
        "application/json", bytes.NewBuffer(reqBody))
        
        // Assert: HTTP 응답 검증
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
        
        // Assert: Repository로 실제 저장 확인
        todoModel, err := repo.GetTodo(int(todoResp.Todo.ID))
        assert.NoError(t, err)
        assert.Equal(t, "dummy title", todoModel.Title)
    })
....
}

```

## 기술 스택
### Backend
- Go + Gin - 간단한 HTTP API 구성
- MySQL - 실제 데이터베이스 테스트
- Redis - Rate Limiting 기능 검증

### Testing
- Testcontainers - Mock 없는 실제 의존성 테스트
- LocalStack - AWS 서비스 통합 테스트
- GitHub Actions - 테스트 자동화

### 사전 요구사항
- Go 1.24+
- Docker & Docker Compose

## 참고 자료
- **"단위 테스트(생산성과 품질을 위한 단위 테스트 원칙과 패턴)"** by 블라디미르 코리코프
- **[Testcontainers](https://testcontainers.org/)** - 실제 의존성을 사용한 테스트 도구
- **[LocalStack](https://localstack.cloud/)** - AWS 서비스 로컬 모킹
