package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenBucket struct {
	mu          sync.Mutex
	tokens      int
	lastUpdated time.Time
}

const (
	bucketSize       = 10          // Maximum number of tokens
	bucketRefillRate = time.Second // 1 tokens per second
)

var buckets sync.Map

func getTokenBucket(clientIP string) *TokenBucket {
	if bucket, ok := buckets.Load(clientIP); ok {
		return bucket.(*TokenBucket)
	}

	bucket := &TokenBucket{
		tokens:      bucketSize,
		lastUpdated: time.Now(),
	}
	buckets.Store(clientIP, bucket)
	return bucket
}

func (tb *TokenBucket) allowRequest() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdated)
	tb.lastUpdated = now

	tb.tokens += int(elapsed / bucketRefillRate)
	if tb.tokens > bucketSize {
		tb.tokens = bucketSize
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/ping-rl", func(c *gin.Context) {
		clientIP := c.ClientIP()
		log.Printf("Client IP: %s", clientIP)
		bucket := getTokenBucket(clientIP)
		if !bucket.allowRequest() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.Run(":8090")

}
