package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// DelionSession stores d_delion_id session info
type DelionSession struct {
	UserID    uint
	CreatedAt time.Time
}

// Session store for d_delion_id (in production, use Redis with TTL)
var DelionSessionStore = make(map[string]*DelionSession)

// SessionTimeout for d_delion_id (7 days)
const DelionSessionTimeout = 7 * 24 * time.Hour

// GenerateDelionID generates 32-byte random hex string
func GenerateDelionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// IsExpired checks if session is expired
func (s *DelionSession) IsExpired() bool {
	return time.Since(s.CreatedAt) > DelionSessionTimeout
}
