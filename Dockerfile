# Build stage
FROM golang:alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app

# Go 모듈 파일 복사 및 의존성 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# config.json 생성 (Build stage에서) - 디버깅 추가
RUN echo '{"server":{"port":8080},"database":{"host":"mysql","port":3306,"user":"todouser","password":"password","name":"todoapp"}}' > config.json

# Runtime stage
FROM alpine:latest

# 필요한 패키지 설치
RUN apk --no-cache add ca-certificates tzdata

# 작업 디렉토리 설정
WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/main .
RUN echo '{"server":{"port":8080},"database":{"host":"mysql","port":3306,"user":"todouser","password":"password","name":"todoapp"}, "redis": {"Host": "localhost","Port": 6379,"Password": "","DB": 0}}' > config.json
# 포트 노출
EXPOSE 8080

# 애플리케이션 실행
CMD ["./main"]