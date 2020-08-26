package blonk

import (
	"errors"

	"github.com/google/uuid"
)

const (
	// DefaultHost is the schema/hostname default for auth sessions.
	DefaultHost = "https://rest-prod.immedia-semi.com"
)

// Session encapsulates a Blink API session.
type Session struct {
	host      string
	uuid      uuid.UUID
	accountID uint64
	clientID  uint64
	authToken string
}

// NewSession will return a new Session object with a given uuid.
func NewSession(host string, uniqueID uuid.UUID) (*Session, error) {
	if uniqueID == uuid.Nil {
		uniqueID = uuid.New()
	}
	return &Session{host: host, uuid: uniqueID}, nil
}

// Authed will register account/client/token information for a session.
func (s *Session) Authed(account, client uint64, token string) error {
	if s.accountID != 0 || s.clientID != 0 || s.authToken != "" {
		return errors.New("session already authed")
	}
	s.accountID, s.clientID, s.authToken = account, client, token
	return nil
}

// Close will end a session and clear it's parameters.
func (s *Session) Close() error {
	if s.uuid == uuid.Nil {
		return errors.New("session already closed")
	}
	s.authToken = ""
	s.clientID = 0
	s.accountID = 0
	s.uuid = uuid.Nil
	return nil
}

// UUID provides the current UUID of this session.
func (s Session) UUID() uuid.UUID {
	return s.uuid
}
