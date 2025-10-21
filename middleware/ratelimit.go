package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/casapps/casreg/config"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RateLimiter interface for different rate limiting backends
type RateLimiter interface {
	Allow(key string, limit int, window time.Duration) (bool, int, time.Time, error)
}

// InMemoryRateLimiter implements rate limiting using in-memory storage
type InMemoryRateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	limit      int
	window     time.Duration
}

func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		buckets: make(map[string]*bucket),
	}

	// Start cleanup goroutine to remove old buckets
	go limiter.cleanup()

	return limiter
}

func (rl *InMemoryRateLimiter) Allow(key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create bucket
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     limit,
			lastRefill: now,
			limit:      limit,
			window:     window,
		}
		rl.buckets[key] = b
	}

	// Refill tokens if window has passed
	if now.Sub(b.lastRefill) >= b.window {
		b.tokens = b.limit
		b.lastRefill = now
	}

	// Check if request is allowed
	if b.tokens > 0 {
		b.tokens--
		resetAt := b.lastRefill.Add(b.window)
		return true, b.tokens, resetAt, nil
	}

	resetAt := b.lastRefill.Add(b.window)
	return false, 0, resetAt, nil
}

func (rl *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			// Remove buckets that haven't been used in 2x their window
			if now.Sub(b.lastRefill) > b.window*2 {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(cfg *config.Config) (*RedisRateLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisRateLimiter{client: client}, nil
}

func (rl *RedisRateLimiter) Allow(key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	ctx := context.Background()
	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

	// Use Redis transaction to atomically check and decrement
	pipe := rl.client.Pipeline()

	// Get current count
	countCmd := pipe.Get(ctx, rateLimitKey)

	// Set expiry
	pipe.Expire(ctx, rateLimitKey, window)

	_, err := pipe.Exec(ctx)

	var count int
	if err == redis.Nil {
		// Key doesn't exist, initialize it
		count = 0
	} else if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("redis error: %w", err)
	} else {
		count, err = strconv.Atoi(countCmd.Val())
		if err != nil {
			count = 0
		}
	}

	// Check if limit exceeded
	if count >= limit {
		ttl := rl.client.TTL(ctx, rateLimitKey).Val()
		resetAt := time.Now().Add(ttl)
		return false, 0, resetAt, nil
	}

	// Increment counter
	newCount := rl.client.Incr(ctx, rateLimitKey).Val()

	// Set expiry if this is the first request
	if newCount == 1 {
		rl.client.Expire(ctx, rateLimitKey, window)
	}

	ttl := rl.client.TTL(ctx, rateLimitKey).Val()
	resetAt := time.Now().Add(ttl)
	remaining := limit - int(newCount)

	return true, remaining, resetAt, nil
}

// Global rate limiter instance
var globalLimiter RateLimiter

// RateLimit middleware applies rate limiting to requests
func RateLimit(cfg *config.Config) func(http.Handler) http.Handler {
	// Initialize rate limiter on first use
	if globalLimiter == nil {
		if cfg.Redis.Enabled {
			limiter, err := NewRedisRateLimiter(cfg)
			if err != nil {
				logrus.WithError(err).Warn("Failed to initialize Redis rate limiter, falling back to in-memory")
				globalLimiter = NewInMemoryRateLimiter()
			} else {
				globalLimiter = limiter
				logrus.Info("Redis rate limiter initialized")
			}
		} else {
			globalLimiter = NewInMemoryRateLimiter()
			logrus.Info("In-memory rate limiter initialized")
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Determine rate limit based on authentication
			var key string
			var limit int

			user := GetUserFromContext(r.Context())
			if user != nil {
				// Authenticated user
				key = fmt.Sprintf("user:%d", user.ID)
				limit = cfg.RateLimit.Requests
			} else {
				// Anonymous user - use IP address
				ip := r.RemoteAddr
				if realIP := r.Context().Value("real_ip"); realIP != nil {
					if ipStr, ok := realIP.(string); ok {
						ip = ipStr
					}
				}
				key = fmt.Sprintf("ip:%s", ip)
				limit = cfg.RateLimit.AnonymousLimit
			}

			// Check rate limit
			allowed, remaining, resetAt, err := globalLimiter.Allow(key, limit, cfg.RateLimit.Window)
			if err != nil {
				logrus.WithError(err).Error("Rate limiter error")
				// On error, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if !allowed {
				// Rate limit exceeded
				retryAfter := int(time.Until(resetAt).Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				logrus.WithFields(logrus.Fields{
					"key":       key,
					"limit":     limit,
					"reset_at":  resetAt,
				}).Warn("Rate limit exceeded")

				http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded. Please try again later."}}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
