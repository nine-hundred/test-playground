package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"time"
)

type RateLimiter struct {
	redisClient *redis.Client
	limit       int
	window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := c.Request.Context()

		val, err := rl.redisClient.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			// Redis 에러 시 요청 허용 (graceful degradation)
			c.Next()
			return
		}

		// 첫 요청인 경우
		if err == redis.Nil {
			// 카운터 초기화 (1로 설정하고 TTL 적용)
			pipe := rl.redisClient.Pipeline()
			pipe.Set(ctx, key, 1, rl.window)
			_, err := pipe.Exec(ctx)
			if err != nil {
				c.Next()
				return
			}

			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.limit-1))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.window).Unix(), 10))

			c.Next()
			return
		}

		currentCount, err := strconv.Atoi(val)
		if err != nil {
			c.Next()
			return
		}

		if currentCount >= rl.limit {
			ttl, _ := rl.redisClient.TTL(ctx, key).Result()
			resetTime := time.Now().Add(ttl).Unix()

			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", rl.limit, rl.window),
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		newCount, err := rl.redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		remaining := rl.limit - int(newCount)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if rl.redisClient.TTL(ctx, key).Val() == -1 {
			rl.redisClient.Expire(ctx, key, rl.window)
		}

		c.Next()
	}
}
