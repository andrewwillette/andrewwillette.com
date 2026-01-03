package traffic

import (
	"net/http"
	"sync"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

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

type TrafficStore struct {
	mu       sync.RWMutex
	requests []Request
}

var store = &TrafficStore{
	requests: make([]Request, 0),
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
		RecordRequest(
			c.Request().URL.Path,
			c.RealIP(),
			c.Request().UserAgent(),
			c.Request().Referer(),
		)
		return next(c)
	}
}

type AdminPageData struct {
	CurrentYear int
	Buckets     []HourlyBucket
	TotalCount  int
	MemoryBytes int
}

func HandleAdminPage(c echo.Context) error {
	buckets := GetHourlyBuckets()
	total := 0
	for _, b := range buckets {
		total += len(b.Requests)
	}

	data := AdminPageData{
		CurrentYear: time.Now().Year(),
		Buckets:     buckets,
		TotalCount:  total,
		MemoryBytes: getMemoryUsage(),
	}
	return c.Render(http.StatusOK, "adminpage", data)
}

func getMemoryUsage() int {
	store.mu.RLock()
	defer store.mu.RUnlock()

	total := 0
	for _, req := range store.requests {
		total += len(req.Path) + len(req.IP) + len(req.UserAgent) + len(req.Referrer) + 24 // 24 bytes for time.Time
	}
	return total
}

func BasicAuthMiddleware() echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		log.Info().Msgf("password: %s, username: %s", config.C.AdminPassword, "empty")
		if password == config.C.AdminPassword && config.C.AdminPassword != "" {
			return true, nil
		}
		return false, nil
	})
}
