package traffic

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
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

var db *sql.DB

func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		ip TEXT NOT NULL,
		user_agent TEXT,
		referrer TEXT,
		timestamp DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS suspicious_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		ip TEXT NOT NULL,
		user_agent TEXT,
		timestamp DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS failed_auths (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip TEXT NOT NULL,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_requests_timestamp ON requests(timestamp);
	CREATE INDEX IF NOT EXISTS idx_suspicious_timestamp ON suspicious_requests(timestamp);
	CREATE INDEX IF NOT EXISTS idx_failed_auths_timestamp ON failed_auths(timestamp);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	log.Info().Msgf("traffic: database initialized at %s", dbPath)
	return nil
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

func isSuspiciousPath(path string) bool {
	lowerPath := strings.ToLower(path)
	for _, suspicious := range suspiciousPaths {
		if strings.Contains(lowerPath, suspicious) {
			return true
		}
	}
	return false
}

func RecordRequest(path, ip, userAgent, referrer string) {
	if db == nil {
		return
	}
	_, err := db.Exec(
		"INSERT INTO requests (path, ip, user_agent, referrer, timestamp) VALUES (?, ?, ?, ?, ?)",
		path, ip, userAgent, referrer, time.Now(),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to record request")
	}
}

func RecordSuspiciousRequest(path, ip, userAgent string) {
	if db == nil {
		return
	}
	_, err := db.Exec(
		"INSERT INTO suspicious_requests (path, ip, user_agent, timestamp) VALUES (?, ?, ?, ?)",
		path, ip, userAgent, time.Now(),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to record suspicious request")
	}
}

func RecordFailedAuth(ip string) {
	if db == nil {
		return
	}
	_, err := db.Exec(
		"INSERT INTO failed_auths (ip, timestamp) VALUES (?, ?)",
		ip, time.Now(),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to record failed auth")
	}
}

func parseTimestamp(s string) time.Time {
	// SQLite stores timestamps in various formats, try common ones
	formats := []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func GetFailedAuthSummary() []FailedAuthSummary {
	if db == nil {
		return nil
	}

	rows, err := db.Query(`
		SELECT ip, COUNT(*) as count, MAX(timestamp) as last
		FROM failed_auths
		GROUP BY ip
		ORDER BY last DESC
	`)
	if err != nil {
		log.Error().Err(err).Msg("failed to query failed auths")
		return nil
	}
	defer rows.Close()

	var results []FailedAuthSummary
	for rows.Next() {
		var summary FailedAuthSummary
		var lastStr string
		if err := rows.Scan(&summary.IP, &summary.Count, &lastStr); err != nil {
			log.Error().Err(err).Msg("failed to scan failed auth row")
			continue
		}
		summary.Last = parseTimestamp(lastStr)
		results = append(results, summary)
	}
	return results
}

func GetSuspiciousSummary() []SuspiciousSummary {
	if db == nil {
		log.Warn().Msg("GetSuspiciousSummary called with nil DB")
		return nil
	}

	rows, err := db.Query(`
		SELECT ip, path, COUNT(*) as count, MIN(timestamp) as first, MAX(timestamp) as last
		FROM suspicious_requests
		GROUP BY ip, path
		ORDER BY last DESC
	`)
	if err != nil {
		log.Error().Err(err).Msg("failed to query suspicious requests")
		return nil
	}
	defer rows.Close()

	var results []SuspiciousSummary
	for rows.Next() {
		var summary SuspiciousSummary
		var firstStr, lastStr string
		if err := rows.Scan(&summary.IP, &summary.Path, &summary.Count, &firstStr, &lastStr); err != nil {
			log.Error().Err(err).Msg("failed to scan suspicious request row")
			continue
		}
		summary.First = parseTimestamp(firstStr)
		summary.Last = parseTimestamp(lastStr)
		results = append(results, summary)
	}
	return results
}

func GetHourlyBuckets() []HourlyBucket {
	if db == nil {
		return nil
	}

	rows, err := db.Query(`
		SELECT path, ip, user_agent, referrer, timestamp
		FROM requests
		ORDER BY timestamp DESC
		LIMIT 1000
	`)
	if err != nil {
		log.Error().Err(err).Msg("failed to query requests")
		return nil
	}
	defer rows.Close()

	bucketMap := make(map[string][]Request)
	for rows.Next() {
		var req Request
		var userAgent, referrer sql.NullString
		var timestampStr string
		if err := rows.Scan(&req.Path, &req.IP, &userAgent, &referrer, &timestampStr); err != nil {
			log.Error().Err(err).Msg("failed to scan request row")
			continue
		}
		req.UserAgent = userAgent.String
		req.Referrer = referrer.String
		req.Timestamp = parseTimestamp(timestampStr)

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

func getTotalRequestCount() int {
	if db == nil {
		return 0
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM requests").Scan(&count)
	if err != nil {
		log.Error().Err(err).Msg("failed to count requests")
		return 0
	}
	return count
}

func getDBSize() int64 {
	if config.C.TrafficDBPath == "" {
		return 0
	}
	info, err := os.Stat(config.C.TrafficDBPath)
	if err != nil {
		return 0
	}
	return info.Size()
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div),
		"KMGTPE"[exp],
	)
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
	DBSize             string
	SuspiciousRequests []SuspiciousSummary
	FailedAuths        []FailedAuthSummary
}

func HandleAdminPage(c echo.Context) error {
	log.Info().Msg("HandleAdminPage")
	buckets := GetHourlyBuckets()

	data := AdminPageData{
		CurrentYear:        time.Now().Year(),
		Buckets:            buckets,
		TotalCount:         getTotalRequestCount(),
		DBSize:             humanBytes(getDBSize()),
		SuspiciousRequests: GetSuspiciousSummary(),
		FailedAuths:        GetFailedAuthSummary(),
	}
	log.Info().Msg("Rendering admin page")
	return c.Render(http.StatusOK, "adminpage", data)
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
