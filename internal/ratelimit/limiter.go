package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-crypto/internal/config"

	"github.com/sirupsen/logrus"
)

// ClientBucket represents a client's rate limit bucket
type ClientBucket struct {
	IP              string
	Tier            string
	TokensPerMinute int
	TokensPerHour   int
	CurrentMinute   int
	CurrentHour     int
	LastMinuteReset time.Time
	LastHourReset   time.Time
	IsBlocked       bool
	BlockedUntil    time.Time
	RequestCount    int64
	LastRequest     time.Time
	BurstTokens     int
	MaxBurstTokens  int
}

// RateLimiter implements a token bucket rate limiter with multiple tiers
type RateLimiter struct {
	config        *config.RateLimitConfig
	clients       map[string]*ClientBucket
	mutex         sync.RWMutex
	logger        *logrus.Logger
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(cfg *config.RateLimitConfig, logger *logrus.Logger) *RateLimiter {
	rl := &RateLimiter{
		config:      cfg,
		clients:     make(map[string]*ClientBucket),
		logger:      logger,
		stopCleanup: make(chan bool),
	}

	// Start cleanup routine if enabled
	if cfg.Enabled && cfg.CleanupInterval > 0 {
		rl.startCleanupRoutine()
	}

	return rl
}

// IsAllowed checks if a request is allowed for the given IP
func (rl *RateLimiter) IsAllowed(ip string) (bool, *RateLimitStatus) {
	if !rl.config.Enabled {
		return true, &RateLimitStatus{
			Allowed:         true,
			RemainingMinute: -1, // unlimited
			RemainingHour:   -1, // unlimited
			ResetTimeMinute: time.Time{},
			ResetTimeHour:   time.Time{},
			Tier:            "unlimited",
		}
	}

	// Check IP whitelist
	if rl.isWhitelisted(ip) {
		return true, &RateLimitStatus{
			Allowed:         true,
			RemainingMinute: -1, // unlimited
			RemainingHour:   -1, // unlimited
			ResetTimeMinute: time.Time{},
			ResetTimeHour:   time.Time{},
			Tier:            "whitelisted",
		}
	}

	// Check IP blacklist
	if rl.isBlacklisted(ip) {
		return false, &RateLimitStatus{
			Allowed:         false,
			RemainingMinute: 0,
			RemainingHour:   0,
			ResetTimeMinute: time.Now().Add(24 * time.Hour), // blocked for 24h
			ResetTimeHour:   time.Now().Add(24 * time.Hour),
			Tier:            "blacklisted",
			BlockedUntil:    time.Now().Add(24 * time.Hour),
		}
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket := rl.getOrCreateBucket(ip)
	now := time.Now()

	// Check if client is currently blocked
	if bucket.IsBlocked && now.Before(bucket.BlockedUntil) {
		return false, &RateLimitStatus{
			Allowed:         false,
			RemainingMinute: 0,
			RemainingHour:   0,
			ResetTimeMinute: bucket.BlockedUntil,
			ResetTimeHour:   bucket.BlockedUntil,
			Tier:            bucket.Tier,
			BlockedUntil:    bucket.BlockedUntil,
		}
	}

	// Reset block status if time has passed
	if bucket.IsBlocked && now.After(bucket.BlockedUntil) {
		bucket.IsBlocked = false
		bucket.BlockedUntil = time.Time{}
		rl.logger.WithField("ip", ip).Info("Client unblocked")
	}

	// Refill tokens based on time passed
	rl.refillTokens(bucket, now)

	tier := rl.config.Tiers[bucket.Tier]

	// Check if unlimited tier
	if tier.RequestsPerMinute == 0 && tier.RequestsPerHour == 0 {
		return true, &RateLimitStatus{
			Allowed:         true,
			RemainingMinute: -1,
			RemainingHour:   -1,
			ResetTimeMinute: time.Time{},
			ResetTimeHour:   time.Time{},
			Tier:            bucket.Tier,
		}
	}

	// Check minute limit
	if tier.RequestsPerMinute > 0 && bucket.CurrentMinute >= tier.RequestsPerMinute {
		// Check if we can use burst tokens
		if bucket.BurstTokens <= 0 {
			rl.handleRateLimitExceeded(bucket, tier, "minute")
			return false, rl.createRateLimitStatus(bucket, tier, false)
		}
		bucket.BurstTokens--
	}

	// Check hour limit
	if tier.RequestsPerHour > 0 && bucket.CurrentHour >= tier.RequestsPerHour {
		rl.handleRateLimitExceeded(bucket, tier, "hour")
		return false, rl.createRateLimitStatus(bucket, tier, false)
	}

	// Allow request and consume tokens
	bucket.CurrentMinute++
	bucket.CurrentHour++
	bucket.RequestCount++
	bucket.LastRequest = now

	return true, rl.createRateLimitStatus(bucket, tier, true)
}

// refillTokens refills the client's token buckets based on time elapsed
func (rl *RateLimiter) refillTokens(bucket *ClientBucket, now time.Time) {
	// Reset minute bucket if a minute has passed
	if now.Sub(bucket.LastMinuteReset) >= time.Minute {
		bucket.CurrentMinute = 0
		bucket.LastMinuteReset = now

		// Refill burst tokens (partial refill based on time passed)
		tier := rl.config.Tiers[bucket.Tier]
		if tier.BurstAllowance > 0 {
			minutesPassed := int(now.Sub(bucket.LastMinuteReset) / time.Minute)
			if minutesPassed < 1 {
				minutesPassed = 1
			}

			refillAmount := minutesPassed * tier.BurstAllowance / 10 // Gradual refill
			bucket.BurstTokens += refillAmount
			if bucket.BurstTokens > tier.BurstAllowance {
				bucket.BurstTokens = tier.BurstAllowance
			}
		}
	}

	// Reset hour bucket if an hour has passed
	if now.Sub(bucket.LastHourReset) >= time.Hour {
		bucket.CurrentHour = 0
		bucket.LastHourReset = now
	}
}

// getOrCreateBucket gets an existing bucket or creates a new one
func (rl *RateLimiter) getOrCreateBucket(ip string) *ClientBucket {
	if bucket, exists := rl.clients[ip]; exists {
		return bucket
	}

	// Create new bucket
	tier := rl.getTierForIP(ip)
	tierConfig := rl.config.Tiers[tier]
	now := time.Now()

	bucket := &ClientBucket{
		IP:              ip,
		Tier:            tier,
		TokensPerMinute: tierConfig.RequestsPerMinute,
		TokensPerHour:   tierConfig.RequestsPerHour,
		CurrentMinute:   0,
		CurrentHour:     0,
		LastMinuteReset: now,
		LastHourReset:   now,
		IsBlocked:       false,
		RequestCount:    0,
		LastRequest:     now,
		BurstTokens:     tierConfig.BurstAllowance,
		MaxBurstTokens:  tierConfig.BurstAllowance,
	}

	rl.clients[ip] = bucket
	rl.logger.WithFields(logrus.Fields{
		"ip":   ip,
		"tier": tier,
	}).Debug("Created new rate limit bucket")

	return bucket
}

// getTierForIP determines the appropriate tier for an IP
func (rl *RateLimiter) getTierForIP(ip string) string {
	// This is where you'd implement tier assignment logic
	// For now, use default tier
	// In a production system, you might:
	// - Check user authentication/API keys
	// - Use IP-based tier assignment
	// - Integrate with user subscription levels
	return rl.config.DefaultTier
}

// handleRateLimitExceeded handles when a client exceeds rate limits
func (rl *RateLimiter) handleRateLimitExceeded(bucket *ClientBucket, tier config.RateLimitTier, limitType string) {
	if tier.BlockDurationMinutes > 0 {
		bucket.IsBlocked = true
		bucket.BlockedUntil = time.Now().Add(time.Duration(tier.BlockDurationMinutes) * time.Minute)

		rl.logger.WithFields(logrus.Fields{
			"ip":            bucket.IP,
			"tier":          bucket.Tier,
			"limit_type":    limitType,
			"blocked_until": bucket.BlockedUntil,
		}).Warn("Client rate limit exceeded and blocked")
	}
}

// createRateLimitStatus creates a rate limit status response
func (rl *RateLimiter) createRateLimitStatus(bucket *ClientBucket, tier config.RateLimitTier, allowed bool) *RateLimitStatus {
	remainingMinute := tier.RequestsPerMinute - bucket.CurrentMinute
	if remainingMinute < 0 {
		remainingMinute = 0
	}

	remainingHour := tier.RequestsPerHour - bucket.CurrentHour
	if remainingHour < 0 {
		remainingHour = 0
	}

	return &RateLimitStatus{
		Allowed:         allowed,
		RemainingMinute: remainingMinute,
		RemainingHour:   remainingHour,
		ResetTimeMinute: bucket.LastMinuteReset.Add(time.Minute),
		ResetTimeHour:   bucket.LastHourReset.Add(time.Hour),
		Tier:            bucket.Tier,
		RequestCount:    bucket.RequestCount,
		BurstTokens:     bucket.BurstTokens,
		MaxBurstTokens:  bucket.MaxBurstTokens,
		BlockedUntil:    bucket.BlockedUntil,
	}
}

// isWhitelisted checks if an IP is whitelisted
func (rl *RateLimiter) isWhitelisted(ip string) bool {
	for _, whiteIP := range rl.config.IPWhitelist {
		if rl.matchIP(ip, whiteIP) {
			return true
		}
	}
	return false
}

// isBlacklisted checks if an IP is blacklisted
func (rl *RateLimiter) isBlacklisted(ip string) bool {
	for _, blackIP := range rl.config.IPBlacklist {
		if rl.matchIP(ip, blackIP) {
			return true
		}
	}
	return false
}

// matchIP checks if an IP matches a pattern (supports CIDR)
func (rl *RateLimiter) matchIP(ip, pattern string) bool {
	// Exact match
	if ip == pattern {
		return true
	}

	// CIDR match
	if strings.Contains(pattern, "/") {
		_, network, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ipAddr := net.ParseIP(ip)
		return network.Contains(ipAddr)
	}

	return false
}

// startCleanupRoutine starts the background cleanup routine
func (rl *RateLimiter) startCleanupRoutine() {
	rl.cleanupTicker = time.NewTicker(time.Duration(rl.config.CleanupInterval) * time.Minute)

	go func() {
		for {
			select {
			case <-rl.cleanupTicker.C:
				rl.cleanup()
			case <-rl.stopCleanup:
				rl.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup removes old client buckets
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cleanupThreshold := 24 * time.Hour // Remove clients not seen for 24 hours
	removed := 0

	for ip, bucket := range rl.clients {
		if now.Sub(bucket.LastRequest) > cleanupThreshold {
			delete(rl.clients, ip)
			removed++
		}
	}

	if removed > 0 {
		rl.logger.WithField("removed_clients", removed).Info("Cleaned up old rate limit buckets")
	}
}

// Stop stops the rate limiter and cleanup routines
func (rl *RateLimiter) Stop() {
	if rl.cleanupTicker != nil {
		close(rl.stopCleanup)
	}
}

// GetClientIP extracts the real client IP from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (nginx proxy)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimitStatus represents the current rate limit status for a client
type RateLimitStatus struct {
	Allowed         bool      `json:"allowed"`
	RemainingMinute int       `json:"remaining_minute"`
	RemainingHour   int       `json:"remaining_hour"`
	ResetTimeMinute time.Time `json:"reset_time_minute"`
	ResetTimeHour   time.Time `json:"reset_time_hour"`
	Tier            string    `json:"tier"`
	RequestCount    int64     `json:"request_count"`
	BurstTokens     int       `json:"burst_tokens"`
	MaxBurstTokens  int       `json:"max_burst_tokens"`
	BlockedUntil    time.Time `json:"blocked_until,omitempty"`
}

// ToHeaders converts rate limit status to HTTP headers
func (rls *RateLimitStatus) ToHeaders() map[string]string {
	headers := map[string]string{
		"X-RateLimit-Limit-Minute":     fmt.Sprintf("%d", rls.RemainingMinute+1), // +1 for current request
		"X-RateLimit-Remaining-Minute": fmt.Sprintf("%d", rls.RemainingMinute),
		"X-RateLimit-Reset-Minute":     fmt.Sprintf("%d", rls.ResetTimeMinute.Unix()),
		"X-RateLimit-Limit-Hour":       fmt.Sprintf("%d", rls.RemainingHour+1),
		"X-RateLimit-Remaining-Hour":   fmt.Sprintf("%d", rls.RemainingHour),
		"X-RateLimit-Reset-Hour":       fmt.Sprintf("%d", rls.ResetTimeHour.Unix()),
		"X-RateLimit-Tier":             rls.Tier,
	}

	if !rls.BlockedUntil.IsZero() {
		headers["X-RateLimit-Blocked-Until"] = fmt.Sprintf("%d", rls.BlockedUntil.Unix())
	}

	return headers
}
