package utils

import (
	"context"
	"math"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []ErrorCode
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []ErrorCode{
			ErrorCodeAPITimeout,
			ErrorCodeAPIRateLimit,
			ErrorCodeAPIServerError,
			ErrorCodeNetworkError,
			ErrorCodeTimeoutError,
		},
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// Retry executes a function with retry logic
func Retry(ctx context.Context, config *RetryConfig, fn RetryableFunc, logger Logger) error {
	if config == nil {
		config = DefaultRetryConfig()
	}
	
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			if attempt > 0 {
				logger.Info("Function succeeded after retry",
					NewField("attempt", attempt),
					NewField("max_retries", config.MaxRetries),
				)
			}
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !isRetryableError(err, config.RetryableErrors) {
			logger.Debug("Error is not retryable, giving up",
				NewField("error", err.Error()),
				NewField("attempt", attempt),
			)
			return err
		}
		
		// Don't wait after the last attempt
		if attempt == config.MaxRetries {
			logger.Error("Max retries exceeded", err,
				NewField("max_retries", config.MaxRetries),
			)
			break
		}
		
		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, config)
		
		logger.Warn("Function failed, retrying",
			NewField("error", err.Error()),
			NewField("attempt", attempt),
			NewField("max_retries", config.MaxRetries),
			NewField("delay_ms", delay.Milliseconds()),
		)
		
		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return lastErr
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error, retryableErrors []ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		for _, retryableCode := range retryableErrors {
			if appErr.Code == retryableCode {
				return true
			}
		}
	}
	return false
}

// calculateDelay calculates the delay for exponential backoff
func calculateDelay(attempt int, config *RetryConfig) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt))
	
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	return time.Duration(delay)
}

// RateLimiter implements rate limiting functionality
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	requests    []time.Time
	logger      Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, window time.Duration, logger Logger) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		requests:    make([]time.Time, 0),
		logger:      logger,
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	now := time.Now()
	
	// Remove old requests outside the window
	rl.cleanupOldRequests(now)
	
	// Check if we're under the limit
	if len(rl.requests) < rl.maxRequests {
		rl.requests = append(rl.requests, now)
		return true
	}
	
	rl.logger.Debug("Rate limit exceeded",
		NewField("max_requests", rl.maxRequests),
		NewField("window", rl.window.String()),
		NewField("current_requests", len(rl.requests)),
	)
	
	return false
}

// WaitForSlot waits until a slot becomes available
func (rl *RateLimiter) WaitForSlot(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}
		
		// Calculate wait time until next slot
		now := time.Now()
		rl.cleanupOldRequests(now)
		
		if len(rl.requests) == 0 {
			// No requests, should be able to proceed
			continue
		}
		
		// Wait until the oldest request expires
		oldestRequest := rl.requests[0]
		waitTime := rl.window - now.Sub(oldestRequest)
		
		if waitTime <= 0 {
			continue
		}
		
		rl.logger.Debug("Waiting for rate limit slot",
			NewField("wait_time_ms", waitTime.Milliseconds()),
		)
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to next iteration
		}
	}
}

// cleanupOldRequests removes requests outside the current window
func (rl *RateLimiter) cleanupOldRequests(now time.Time) {
	cutoff := now.Add(-rl.window)
	
	// Find the first request within the window
	validIndex := 0
	for i, req := range rl.requests {
		if req.After(cutoff) {
			validIndex = i
			break
		}
		validIndex = i + 1
	}
	
	// Remove old requests
	if validIndex > 0 {
		rl.requests = rl.requests[validIndex:]
	}
}

// GetCurrentRequestCount returns the current number of requests in the window
func (rl *RateLimiter) GetCurrentRequestCount() int {
	rl.cleanupOldRequests(time.Now())
	return len(rl.requests)
}

// GetTimeToNextSlot returns the time until the next slot becomes available
func (rl *RateLimiter) GetTimeToNextSlot() time.Duration {
	now := time.Now()
	rl.cleanupOldRequests(now)
	
	if len(rl.requests) < rl.maxRequests {
		return 0
	}
	
	oldestRequest := rl.requests[0]
	return rl.window - now.Sub(oldestRequest)
}

// RetryWithRateLimit combines retry logic with rate limiting
func RetryWithRateLimit(ctx context.Context, retryConfig *RetryConfig, rateLimiter *RateLimiter, fn RetryableFunc, logger Logger) error {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}
	
	return Retry(ctx, retryConfig, func() error {
		// Wait for rate limit slot
		if err := rateLimiter.WaitForSlot(ctx); err != nil {
			return err
		}
		
		// Execute the function
		return fn()
	}, logger)
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitState
	logger          Logger
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitStateClosed:
		return "closed"
	case CircuitStateOpen:
		return "open"
	case CircuitStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitStateClosed,
		logger:       logger,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == CircuitStateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitStateHalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open state")
		} else {
			return NewAppError(ErrorCodeNetworkError, "Circuit breaker is open", nil)
		}
	}
	
	err := fn()
	
	if err != nil {
		cb.onFailure(err)
		return err
	}
	
	cb.onSuccess()
	return nil
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	
	if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateClosed
		cb.logger.Info("Circuit breaker transitioning to closed state")
	}
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure(err error) {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitStateOpen
		cb.logger.Error("Circuit breaker transitioning to open state", err,
			NewField("failure_count", cb.failureCount),
			NewField("max_failures", cb.maxFailures),
		)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	return cb.failureCount
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.failureCount = 0
	cb.state = CircuitStateClosed
	cb.logger.Info("Circuit breaker manually reset to closed state")
}

// RetryWithCircuitBreaker combines retry logic with circuit breaker
func RetryWithCircuitBreaker(ctx context.Context, retryConfig *RetryConfig, circuitBreaker *CircuitBreaker, fn RetryableFunc, logger Logger) error {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}
	
	return Retry(ctx, retryConfig, func() error {
		return circuitBreaker.Execute(fn)
	}, logger)
}