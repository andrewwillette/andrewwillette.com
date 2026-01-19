-- Traffic monitoring schema

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

CREATE TABLE IF NOT EXISTS bad_ips (
    ip TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    threat_level INTEGER DEFAULT 1,
    added_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_requests_timestamp ON requests(timestamp);
CREATE INDEX IF NOT EXISTS idx_requests_ip ON requests(ip);
CREATE INDEX IF NOT EXISTS idx_suspicious_timestamp ON suspicious_requests(timestamp);
CREATE INDEX IF NOT EXISTS idx_suspicious_ip ON suspicious_requests(ip);
CREATE INDEX IF NOT EXISTS idx_failed_auths_timestamp ON failed_auths(timestamp);
CREATE INDEX IF NOT EXISTS idx_failed_auths_ip ON failed_auths(ip);
