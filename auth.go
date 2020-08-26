package blonk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Login describes a login request.
type Login struct {
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	UUID       uuid.UUID `json:"uuid,omitempty"`
	Device     string    `json:"device"`
	DeviceName string    `json:"client_name"`
}

// NewLogin returns a Login api request structure.
func (s *Session) NewLogin(email, password string) (*Request, error) {
	if len(email) == 0 || !strings.ContainsRune(email, '@') {
		return nil, errors.New("invalid email")
	}
	if len(password) == 0 || len(password) > 255 {
		return nil, errors.New("invalid password")
	}
	login := Login{Email: email, Password: password, UUID: s.uuid, Device: "Blonk", DeviceName: "Blonk golang client"}
	return NewRequest(s, "/api/v4/account/login", login, false)
}

// LoginReply describes the schema for the API's response to a Login request.
type LoginReply struct {
	Account struct {
		ID         uint64 `json:"id"`
		Verify     bool   `json:"verification_required"`
		NewAccount bool   `json:"new_account"`
	} `json:"account"`
	Client struct {
		ID     uint64 `json:"id"`
		Verify bool   `json:"verification_required"`
	} `json:"client"`
	AuthToken struct {
		AuthToken string `json:"authtoken"`
		Message   string `json:"message"`
	} `json:"authtoken"`
	LockoutRemaining      int64 `json:"lockout_time_remaining"`
	ForcePasswordReset    bool  `json:"force_password_reset"`
	AllowPinResendSeconds int64 `json:"allow_pin_resend_seconds"`
}

// LogoutReply is the schema for the API's response to a logout request.
type LogoutReply struct {
	Message string `json:"message"`
}

func newAccountRequest(session *Session, url string, body interface{}, withAuth bool) (request *Request, err error) {
	if session.accountID == 0 {
		err = errors.New("missing account id")
	} else if session.clientID == 0 {
		err = errors.New("missing client id")
	} else {
		endpoint := fmt.Sprintf("/api/v4/account/%d/client/%d/%s", session.accountID, session.clientID, url)
		request, err = NewRequest(session, endpoint, body, withAuth)
	}
	return
}

// NewLogout returns a new logout request.
func (s *Session) NewLogout(accountID, clientID uint64) (request *Request, err error) {
	return newAccountRequest(s, "logout", nil, true)
}

// VerifyPin is the schema for verifying login with a pin.
type VerifyPin struct {
	Pin string `json:"pin"`
}

// VerifyPinResult is the schema for the API's response to a VerifyPin request.
type VerifyPinResult struct {
	Valid         bool   `json:"valid"`
	RequireNewPin bool   `json:"require_new_pin"`
	Message       string `json:"message"`
	Code          int    `json:"code"`
}

// NewVerifyPin returns a pin verification request.
func (s *Session) NewVerifyPin(pin string) (request *Request, err error) {
	if len(pin) == 0 {
		err = errors.New("missing pin")
	} else {
		request, err = newAccountRequest(s, "pin/verify", &VerifyPin{pin}, true)
	}
	return
}
