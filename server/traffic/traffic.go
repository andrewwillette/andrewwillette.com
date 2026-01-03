package traffic

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

var suspiciousPaths = []string{
	"/wp-admin", "/wp-login", "/wp-content", "/wordpress",
	"/.env", "/.git", "/.gitignore", "/.htaccess",
	"/phpmyadmin", "/pma", "/mysql", "/adminer",
	"/admin", "/administrator", "/login", "/signin",
	"/config", "/backup", "/db", "/database",
	"/shell", "/cmd", "/eval", "/exec",
	"/api/", "/xmlrpc.php", "/wp-json",
}

type Request struct {
	Path      string
	IP        string
	UserAgent string
	Referrer  string
	Timestamp time.Time
}

type HourlyBucket struct {
	Hour     string
	Requests []Request
}

type SuspiciousRequest struct {
	Path      string
	IP        string
	UserAgent string
	Timestamp time.Time
}

type FailedAuth struct {
	IP        string
	Timestamp time.Time
}

type Store struct {
	mu                 sync.RWMutex
	requests           []Request
	suspiciousRequests []SuspiciousRequest
	failedAuths        []FailedAuth
}

var store = &Store{
	requests:           make([]Request, 0),
	suspiciousRequests: make([]SuspiciousRequest, 0),
	failedAuths:        make([]FailedAuth, 0),
}

func RecordRequest(path, ip, userAgent, referrer string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	req := Request{
		Path:      path,
		IP:        ip,
		UserAgent: userAgent,
		Referrer:  referrer,
		Timestamp: time.Now(),
	}

	store.requests = append(store.requests, req)
}

func isSuspiciousPath(path string) bool {
	lowerPath := strings.ToLower(path)
	for _, suspicious := range suspiciousPaths {
		if strings.Contains(lowerPath, suspicious) {
			return true
		}
	}
	return false
}

func RecordSuspiciousRequest(path, ip, userAgent string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.suspiciousRequests = append(store.suspiciousRequests, SuspiciousRequest{
		Path:      path,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	})
}

func RecordFailedAuth(ip string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.failedAuths = append(store.failedAuths, FailedAuth{
		IP:        ip,
		Timestamp: time.Now(),
	})
}

type FailedAuthSummary struct {
	IP    string
	Count int
	Last  time.Time
}

type SuspiciousSummary struct {
	IP    string
	Path  string
	Count int
	First time.Time
	Last  time.Time
}

func GetFailedAuthSummary() []FailedAuthSummary {
	store.mu.RLock()
	defer store.mu.RUnlock()

	counts := make(map[string]*FailedAuthSummary)
	for _, fa := range store.failedAuths {
		if summary, exists := counts[fa.IP]; exists {
			summary.Count++
			if fa.Timestamp.After(summary.Last) {
				summary.Last = fa.Timestamp
			}
		} else {
			counts[fa.IP] = &FailedAuthSummary{
				IP:    fa.IP,
				Count: 1,
				Last:  fa.Timestamp,
			}
		}
	}

	result := make([]FailedAuthSummary, 0, len(counts))
	for _, summary := range counts {
		result = append(result, *summary)
	}
	return result
}

func GetSuspiciousSummary() []SuspiciousSummary {
	store.mu.RLock()
	defer store.mu.RUnlock()

	groups := make(map[string]*SuspiciousSummary)
	for _, req := range store.suspiciousRequests {
		key := req.IP + "|" + req.Path
		if summary, exists := groups[key]; exists {
			summary.Count++
			if req.Timestamp.Before(summary.First) {
				summary.First = req.Timestamp
			}
			if req.Timestamp.After(summary.Last) {
				summary.Last = req.Timestamp
			}
		} else {
			groups[key] = &SuspiciousSummary{
				IP:    req.IP,
				Path:  req.Path,
				Count: 1,
				First: req.Timestamp,
				Last:  req.Timestamp,
			}
		}
	}

	result := make([]SuspiciousSummary, 0, len(groups))
	for _, summary := range groups {
		result = append(result, *summary)
	}
	return result
}

func GetHourlyBuckets() []HourlyBucket {
	store.mu.RLock()
	defer store.mu.RUnlock()

	bucketMap := make(map[string][]Request)

	for _, req := range store.requests {
		hourKey := req.Timestamp.Format("2006-01-02 15:00")
		bucketMap[hourKey] = append(bucketMap[hourKey], req)
	}

	buckets := make([]HourlyBucket, 0, len(bucketMap))
	for hour, reqs := range bucketMap {
		buckets = append(buckets, HourlyBucket{
			Hour:     hour,
			Requests: reqs,
		})
	}

	// Sort buckets by hour descending (most recent first)
	for i := 0; i < len(buckets)-1; i++ {
		for j := i + 1; j < len(buckets); j++ {
			if buckets[i].Hour < buckets[j].Hour {
				buckets[i], buckets[j] = buckets[j], buckets[i]
			}
		}
	}

	return buckets
}

func TrackingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		ip := c.RealIP()
		userAgent := c.Request().UserAgent()

		RecordRequest(path, ip, userAgent, c.Request().Referer())

		if isSuspiciousPath(path) {
			RecordSuspiciousRequest(path, ip, userAgent)
		}

		return next(c)
	}
}

type AdminPageData struct {
	CurrentYear        int
	Buckets            []HourlyBucket
	TotalCount         int
	MemoryBytes        int
	SuspiciousRequests []SuspiciousSummary
	FailedAuths        []FailedAuthSummary
}

func HandleAdminPage(c echo.Context) error {
	buckets := GetHourlyBuckets()
	total := 0
	for _, b := range buckets {
		total += len(b.Requests)
	}

	data := AdminPageData{
		CurrentYear:        time.Now().Year(),
		Buckets:            buckets,
		TotalCount:         total,
		MemoryBytes:        getMemoryUsage(),
		SuspiciousRequests: GetSuspiciousSummary(),
		FailedAuths:        GetFailedAuthSummary(),
	}
	return c.Render(http.StatusOK, "adminpage", data)
}

func getMemoryUsage() int {
	store.mu.RLock()
	defer store.mu.RUnlock()

	total := 0

	// Regular requests
	for _, req := range store.requests {
		total += len(req.Path) + len(req.IP) + len(req.UserAgent) + len(req.Referrer) + 24
	}

	// Suspicious requests
	for _, req := range store.suspiciousRequests {
		total += len(req.Path) + len(req.IP) + len(req.UserAgent) + 24
	}

	// Failed auths
	for _, fa := range store.failedAuths {
		total += len(fa.IP) + 24
	}

	return total
}

func BasicAuthMiddleware() echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if password == config.C.AdminPassword && config.C.AdminPassword != "" {
			return true, nil
		}
		RecordFailedAuth(c.RealIP())
		log.Info().Msgf("failed auth attempt from IP: %s", c.RealIP())
		return false, nil
	})
}
