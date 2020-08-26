package blonk

import (
	"encoding/json"
	"errors"
)

// Request describes a blink-api request.
type Request struct {
	session *Session
	URL     string
	Headers map[string]string
	Body    []byte
}

// NewRequest creates and populates a new blink-api request packet.
func NewRequest(session *Session, url string, body interface{}, withAuth bool) (*Request, error) {
	if withAuth && len(session.authToken) == 0 {
		return nil, errors.New("missing auth token")
	}
	request := Request{
		session: session, URL: session.host + url, Headers: make(map[string]string, 4),
	}
	if body != nil {
		request.Headers["Content-Type"] = "application/json"
		marshalled, error := json.Marshal(body)
		if error != nil {
			return nil, error
		}
		request.Body = marshalled
	}
	if withAuth {
		request.Headers["TOKEN_AUTH"] = session.authToken
	}
	return &request, nil
}
