package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"backend-go/pkg/logger"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type loginFailureRecord struct {
	timestamps  []time.Time
	bannedUntil time.Time
}

// LoginBruteForceBlocker tracks failed login attempts per IP and per account (email)
// to prevent brute-force attacks and credential stuffing.
type LoginBruteForceBlocker struct {
	mu           sync.RWMutex
	ipRecords    map[string]*loginFailureRecord
	userRecords  map[string]*loginFailureRecord
	maxAttempts  int
	window       time.Duration
	lockDuration time.Duration
	whitelist    map[string]bool
	stopChan     chan struct{}
}

// NewLoginBruteForceBlocker creates a new brute-force limiter.
// Default: 5 failed attempts within 5 minutes results in a 15-minute lockout.
func NewLoginBruteForceBlocker(maxAttempts int, window time.Duration, lockDuration time.Duration) *LoginBruteForceBlocker {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if window <= 0 {
		window = 5 * time.Minute
	}
	if lockDuration <= 0 {
		lockDuration = 15 * time.Minute
	}

	blocker := &LoginBruteForceBlocker{
		ipRecords:    make(map[string]*loginFailureRecord),
		userRecords:  make(map[string]*loginFailureRecord),
		maxAttempts:  maxAttempts,
		window:       window,
		lockDuration: lockDuration,
		whitelist:    make(map[string]bool),
		stopChan:     make(chan struct{}),
	}

	// Periodic cleanup worker
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				blocker.cleanup()
			case <-blocker.stopChan:
				return
			}
		}
	}()

	return blocker
}

func (b *LoginBruteForceBlocker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	select {
	case <-b.stopChan:
	default:
		close(b.stopChan)
	}
}

func (b *LoginBruteForceBlocker) SetWhitelisted(ip string, whitelisted bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if whitelisted {
		b.whitelist[ip] = true
	} else {
		delete(b.whitelist, ip)
	}
}

// IsLocked checks whether either the client IP or the account identifier (email) is locked out.
func (b *LoginBruteForceBlocker) IsLocked(ip, identifier string) (bool, time.Duration, string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	now := time.Now()

	// 1. Check IP lock
	if !b.whitelist[ip] {
		if rec, exists := b.ipRecords[ip]; exists && rec.bannedUntil.After(now) {
			return true, rec.bannedUntil.Sub(now), "ip"
		}
	}

	// 2. Check identifier lock
	cleanIdent := strings.ToLower(strings.TrimSpace(identifier))
	if cleanIdent != "" {
		if rec, exists := b.userRecords[cleanIdent]; exists && rec.bannedUntil.After(now) {
			return true, rec.bannedUntil.Sub(now), "account"
		}
	}

	return false, 0, ""
}

// RecordFailure records a failed authentication attempt for both the IP and the identifier.
// Returns true if either the IP or the account was newly locked out as a result.
func (b *LoginBruteForceBlocker) RecordFailure(ip, identifier string) (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-b.window)
	newlyLocked := false
	maxRem := time.Duration(0)

	// Record for IP
	if !b.whitelist[ip] {
		rec, exists := b.ipRecords[ip]
		if !exists {
			rec = &loginFailureRecord{timestamps: []time.Time{now}}
			b.ipRecords[ip] = rec
		} else if !rec.bannedUntil.After(now) {
			valid := rec.timestamps[:0]
			for _, t := range rec.timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			valid = append(valid, now)
			rec.timestamps = valid

			if len(rec.timestamps) >= b.maxAttempts {
				rec.bannedUntil = now.Add(b.lockDuration)
				rec.timestamps = nil
				newlyLocked = true
				if b.lockDuration > maxRem {
					maxRem = b.lockDuration
				}
			}
		} else {
			rem := rec.bannedUntil.Sub(now)
			if rem > maxRem {
				maxRem = rem
			}
		}
	}

	// Record for Account identifier (email)
	cleanIdent := strings.ToLower(strings.TrimSpace(identifier))
	if cleanIdent != "" {
		rec, exists := b.userRecords[cleanIdent]
		if !exists {
			rec = &loginFailureRecord{timestamps: []time.Time{now}}
			b.userRecords[cleanIdent] = rec
		} else if !rec.bannedUntil.After(now) {
			valid := rec.timestamps[:0]
			for _, t := range rec.timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			valid = append(valid, now)
			rec.timestamps = valid

			if len(rec.timestamps) >= b.maxAttempts {
				rec.bannedUntil = now.Add(b.lockDuration)
				rec.timestamps = nil
				newlyLocked = true
				if b.lockDuration > maxRem {
					maxRem = b.lockDuration
				}
			}
		} else {
			rem := rec.bannedUntil.Sub(now)
			if rem > maxRem {
				maxRem = rem
			}
		}
	}

	if newlyLocked {
		return true, maxRem
	}

	// If already locked
	if maxRem > 0 {
		return true, maxRem
	}

	return false, 0
}

// RecordSuccess clears failure records on a successful login.
func (b *LoginBruteForceBlocker) RecordSuccess(ip, identifier string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.ipRecords, ip)
	cleanIdent := strings.ToLower(strings.TrimSpace(identifier))
	if cleanIdent != "" {
		delete(b.userRecords, cleanIdent)
	}
}

// Reset clears all records (useful for testing or admin unban).
func (b *LoginBruteForceBlocker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ipRecords = make(map[string]*loginFailureRecord)
	b.userRecords = make(map[string]*loginFailureRecord)
}

func (b *LoginBruteForceBlocker) cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-b.window)

	for k, rec := range b.ipRecords {
		if rec.bannedUntil.After(now) {
			continue
		}
		hasRecent := false
		for _, t := range rec.timestamps {
			if t.After(cutoff) {
				hasRecent = true
				break
			}
		}
		if !hasRecent {
			delete(b.ipRecords, k)
		}
	}

	for k, rec := range b.userRecords {
		if rec.bannedUntil.After(now) {
			continue
		}
		hasRecent := false
		for _, t := range rec.timestamps {
			if t.After(cutoff) {
				hasRecent = true
				break
			}
		}
		if !hasRecent {
			delete(b.userRecords, k)
		}
	}
}

var (
	defaultLoginBlocker *LoginBruteForceBlocker
	loginBlockerOnce    sync.Once
)

// GetDefaultLoginBlocker returns the global login brute-force blocker singleton.
// Default configuration: 5 failed attempts within 5 minutes triggers a 15-minute lockout.
func GetDefaultLoginBlocker() *LoginBruteForceBlocker {
	loginBlockerOnce.Do(func() {
		defaultLoginBlocker = NewLoginBruteForceBlocker(5, 5*time.Minute, 15*time.Minute)
	})
	return defaultLoginBlocker
}

// RespondBruteForceLocked sends a standardized HTTP 429 response when locked out.
func RespondBruteForceLocked(c *gin.Context, remaining time.Duration, reason string) {
	seconds := int(remaining.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	minutes := (seconds + 59) / 60

	c.Header("Retry-After", fmt.Sprintf("%d", seconds))
	response.JSON(c, http.StatusTooManyRequests, gin.H{
		"code":                "TOO_MANY_FAILED_ATTEMPTS",
		"error":               "Too Many Requests",
		"message":             fmt.Sprintf("Too many failed login attempts. To protect system security, access is temporarily locked for %d minutes. Please try again later.", minutes),
		"retry_after_seconds": seconds,
		"lock_reason":         reason,
	})
	c.Abort()
}

// LogSecurityBruteForce logs security alerts to both local logger and Prolead client.
func LogSecurityBruteForce(c *gin.Context, clientIP, email, details string) {
	msg := fmt.Sprintf("[SECURITY BRUTE FORCE] IP %s targeted %s: %s", clientIP, logger.MaskPII(email), details)
	logger.Warn(msg)

	if client := logger.GetProleadClient(); client != nil {
		traceID, _ := c.Get("trace_id")
		tID, _ := traceID.(string)
		client.Log(&logger.LogPayload{
			Level:   "WARN",
			Message: msg,
			TraceID: tID,
			HTTP: &logger.HTTPMetadata{
				Method:     c.Request.Method,
				Path:       c.Request.URL.Path,
				StatusCode: http.StatusTooManyRequests,
				ClientIP:   clientIP,
				UserAgent:  c.Request.UserAgent(),
			},
			Context: map[string]interface{}{
				"security_alert": "login_brute_force_lockout",
				"client_ip":      clientIP,
				"targeted_email": logger.MaskPII(email),
				"details":        details,
			},
		})
	}
}
