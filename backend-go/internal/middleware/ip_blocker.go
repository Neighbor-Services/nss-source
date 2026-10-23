package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"backend-go/pkg/logger"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type ipScanRecord struct {
	timestamps  []time.Time
	bannedUntil time.Time
}

// IPBlocker tracks 404 error frequency per IP and temporarily bans aggressive scanners.
type IPBlocker struct {
	mu          sync.RWMutex
	records     map[string]*ipScanRecord
	max404s     int
	window      time.Duration
	banDuration time.Duration
	whitelist   map[string]bool
	stopChan    chan struct{}
}

// NewIPBlocker creates a new tracker to detect and ban 404 scanners.
// Default: 10 404 errors within 5 minutes results in a 7-day IP ban.
func NewIPBlocker(max404s int, window time.Duration, banDuration time.Duration) *IPBlocker {
	if max404s <= 0 {
		max404s = 10
	}
	if window <= 0 {
		window = 5 * time.Minute
	}
	if banDuration <= 0 {
		banDuration = 7 * 24 * time.Hour // 7 days
	}

	blocker := &IPBlocker{
		records:     make(map[string]*ipScanRecord),
		max404s:     max404s,
		window:      window,
		banDuration: banDuration,
		whitelist:   make(map[string]bool),
		stopChan:    make(chan struct{}),
	}

	// Start periodic cleanup worker
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
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

// Stop cleanly terminates the background cleanup goroutine.
func (b *IPBlocker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	select {
	case <-b.stopChan:
		// already closed
	default:
		close(b.stopChan)
	}
}

// SetWhitelisted adds or removes an IP from the whitelist (whitelisted IPs are never blocked).
func (b *IPBlocker) SetWhitelisted(ip string, whitelisted bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if whitelisted {
		b.whitelist[ip] = true
	} else {
		delete(b.whitelist, ip)
	}
}

// IsBanned checks if an IP is currently banned, and returns the remaining ban duration.
func (b *IPBlocker) IsBanned(ip string) (bool, time.Duration) {
	b.mu.RLock()
	rec, exists := b.records[ip]
	whitelisted := b.whitelist[ip]
	b.mu.RUnlock()

	if whitelisted || !exists {
		return false, 0
	}

	now := time.Now()
	if rec.bannedUntil.After(now) {
		return true, rec.bannedUntil.Sub(now)
	}

	return false, 0
}

// Record404 records a 404 status occurrence for an IP. Returns true if the IP was newly banned.
func (b *IPBlocker) Record404(ip string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.whitelist[ip] {
		return false
	}

	now := time.Now()
	cutoff := now.Add(-b.window)

	rec, exists := b.records[ip]
	if !exists {
		rec = &ipScanRecord{
			timestamps: []time.Time{now},
		}
		b.records[ip] = rec
		return false
	}

	// If already banned, don't re-ban
	if rec.bannedUntil.After(now) {
		return false
	}

	// Filter out expired timestamps outside the sliding window
	validTimestamps := rec.timestamps[:0]
	for _, t := range rec.timestamps {
		if t.After(cutoff) {
			validTimestamps = append(validTimestamps, t)
		}
	}
	validTimestamps = append(validTimestamps, now)
	rec.timestamps = validTimestamps

	// Check if threshold reached
	if len(rec.timestamps) >= b.max404s {
		rec.bannedUntil = now.Add(b.banDuration)
		rec.timestamps = nil // clear history once banned
		return true
	}

	return false
}

// Unban manually removes a ban on an IP address.
func (b *IPBlocker) Unban(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.records, ip)
}

// GetBannedIPs returns a snapshot of currently active bans.
func (b *IPBlocker) GetBannedIPs() map[string]time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()

	now := time.Now()
	res := make(map[string]time.Time)
	for ip, rec := range b.records {
		if rec.bannedUntil.After(now) {
			res[ip] = rec.bannedUntil
		}
	}
	return res
}

func (b *IPBlocker) cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-b.window)

	for ip, rec := range b.records {
		if rec.bannedUntil.After(now) {
			continue // keep active ban
		}
		hasRecent := false
		for _, t := range rec.timestamps {
			if t.After(cutoff) {
				hasRecent = true
				break
			}
		}
		if !hasRecent {
			delete(b.records, ip)
		}
	}
}

// Default404Blocker is the default global instance with 10 attempts / 5 mins / 7 days ban.
var (
	Default404Blocker *IPBlocker
	blockerOnce       sync.Once
)

// GetDefault404Blocker returns the singleton blocker instance.
func GetDefault404Blocker() *IPBlocker {
	blockerOnce.Do(func() {
		Default404Blocker = NewIPBlocker(10, 5*time.Minute, 7*24*time.Hour)
	})
	return Default404Blocker
}

// Anti404Scanner returns Gin middleware that blocks IPs with excessive 404 errors for 7 days.
func Anti404Scanner(blocker *IPBlocker) gin.HandlerFunc {
	if blocker == nil {
		blocker = GetDefault404Blocker()
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 1. Pre-handler check: if the IP is currently banned, block immediately
		if banned, remaining := blocker.IsBanned(clientIP); banned {
			days := int(remaining.Hours()) / 24
			hours := int(remaining.Hours()) % 24
			c.Header("Retry-After", fmt.Sprintf("%.0f", remaining.Seconds()))
			response.JSON(c, http.StatusForbidden, gin.H{
 "code": "IP_BLOCKED_SCANNER",
 "error": "Forbidden",
 "message": fmt.Sprintf("Access denied: Your IP (%s) is temporarily blocked for 7 days due to excessive invalid endpoint requests. Ban expires in %dd %dh.", clientIP, days, hours),
			})
			c.Abort()
			return
		}

		// 2. Process request
		c.Next()

		// 3. Post-handler check: if response is 404 Not Found, record and check threshold
		if c.Writer.Status() == http.StatusNotFound {
			if newlyBanned := blocker.Record404(clientIP); newlyBanned {
				path := c.Request.URL.Path
				logger.Warn(fmt.Sprintf("[SECURITY BAN] IP %s banned for 7 days after exceeding %d 404 errors within %s. Last path: %s", clientIP, blocker.max404s, blocker.window, path))

				if client := logger.GetProleadClient(); client != nil {
					traceID, _ := c.Get("trace_id")
					tID, _ := traceID.(string)
					client.Log(&logger.LogPayload{
						Level: "WARN",
						Message: fmt.Sprintf("[SECURITY BAN] IP %s automatically blocked for 7 days after 10x 404 errors (scanned: %s)", clientIP, path),
						TraceID: tID,
						HTTP: &logger.HTTPMetadata{
							Method:     c.Request.Method,
							Path:       path,
							StatusCode: http.StatusNotFound,
							ClientIP:   clientIP,
							UserAgent:  c.Request.UserAgent(),
						},
						Context: map[string]interface{}{
 "security_alert": "scanner_ip_ban",
 "client_ip":      clientIP,
 "ban_duration": "7_days",
 "reason": "Continuous 404 scanning detected",
						},
					})
				}
			}
		}
	}
}
