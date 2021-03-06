package libkb

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

var ErrLoginSessionNotLoaded = errors.New("LoginSession not loaded")
var ErrLoginSessionCleared = errors.New("LoginSession already cleared")

type LoginSession struct {
	sessionFor      string // set by constructor
	salt            []byte // retrieved from server, or set by WithSalt constructor
	loginSessionB64 string
	loginSession    []byte // decoded from above parameter
	loaded          bool   // load state
	cleared         bool   // clear state
	Contextified
}

func NewLoginSession(emailOrUsername string, g *GlobalContext) *LoginSession {
	return &LoginSession{
		sessionFor:   emailOrUsername,
		Contextified: NewContextified(g),
	}
}

// Upon signup, a login session is created with a generated salt.
func NewLoginSessionWithSalt(emailOrUsername string, salt []byte, g *GlobalContext) *LoginSession {
	ls := NewLoginSession(emailOrUsername, g)
	ls.salt = salt
	ls.loaded = true
	ls.cleared = true
	return ls
}

func (s *LoginSession) Session() ([]byte, error) {
	if s == nil {
		return nil, ErrLoginSessionNotLoaded
	}
	if !s.loaded {
		return nil, ErrLoginSessionNotLoaded
	}
	if s.cleared {
		return nil, ErrLoginSessionCleared
	}
	return s.loginSession, nil
}

func (s *LoginSession) SessionEncoded() (string, error) {
	if s == nil {
		return "", ErrLoginSessionNotLoaded
	}
	if !s.loaded {
		return "", ErrLoginSessionNotLoaded
	}
	if s.cleared {
		return "", ErrLoginSessionCleared
	}
	return s.loginSessionB64, nil
}

func (s *LoginSession) ExistsFor(emailOrUsername string) bool {
	if s == nil {
		return false
	}
	if s.sessionFor != emailOrUsername {
		return false
	}
	if s.cleared {
		return false
	}
	if s.loginSession == nil {
		return false
	}
	return true
}

func (s *LoginSession) Clear() error {
	if s == nil {
		return nil
	}
	if !s.loaded {
		return ErrLoginSessionNotLoaded
	}
	s.loginSession = nil
	s.loginSessionB64 = ""
	s.cleared = true
	return nil
}

func (s *LoginSession) Salt() ([]byte, error) {
	if s == nil {
		return nil, ErrLoginSessionNotLoaded
	}
	if !s.loaded {
		return nil, ErrLoginSessionNotLoaded
	}
	return s.salt, nil
}

func (s *LoginSession) Dump() {
	if s == nil {
		fmt.Printf("LoginSession Dump: nil\n")
		return
	}
	fmt.Printf("sessionFor: %q\n", s.sessionFor)
	fmt.Printf("loaded: %v\n", s.loaded)
	fmt.Printf("cleared: %v\n", s.cleared)
	fmt.Printf("salt: %x\n", s.salt)
	fmt.Printf("loginSessionB64: %s\n", s.loginSessionB64)
	fmt.Printf("\n")
}

func (s *LoginSession) Load() error {
	if s.loaded {
		return fmt.Errorf("LoginSession already loaded for %s", s.sessionFor)
	}

	res, err := s.G().API.Get(APIArg{
		Endpoint:    "getsalt",
		NeedSession: false,
		Args: HTTPArgs{
			"email_or_username": S{Val: s.sessionFor},
		},
	})
	if err != nil {
		return err
	}

	shex, err := res.Body.AtKey("salt").GetString()
	if err != nil {
		return err
	}

	salt, err := hex.DecodeString(shex)
	if err != nil {
		return err
	}

	b64, err := res.Body.AtKey("login_session").GetString()
	if err != nil {
		return err
	}

	ls, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}

	s.salt = salt
	s.loginSessionB64 = b64
	s.loginSession = ls
	s.loaded = true

	return nil
}
