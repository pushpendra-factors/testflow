package helpers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	C "factors/config"

	"github.com/gorilla/securecookie"
)

const (
	SecondsInOneDay      = 86400
	SecondsInFifteenDays = SecondsInOneDay * 15
	SecondsInOneMonth    = SecondsInOneDay * 30
	ExpireCookie         = -1
)

var (
	ErrExpired = errors.New("expired")
)

type ProtectedFields struct {
	Email string `json:"e"`
	ExpAt int64  `json:"exp"`
}

type AuthData struct {
	AgentUUID       string `json:"au"`
	ProtectedFields string `json:"pf"`
}

func GetAuthData(email, agentUUID, key string, dur time.Duration) (string, error) {

	if email == "" || agentUUID == "" || key == "" {
		return "", errors.New("missing params")
	}

	pf := ProtectedFields{Email: email, ExpAt: time.Now().UTC().Add(dur).Unix()}

	encPfBytes, err := createSecureData([]byte(key), pf)
	if err != nil {
		return "", err
	}

	ad := AuthData{
		AgentUUID:       agentUUID,
		ProtectedFields: string(encPfBytes),
	}

	adBytes, err := json.Marshal(ad)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(adBytes), nil
}

func ParseAuthData(data string) (*AuthData, error) {

	if data == "" {
		return nil, errors.New("missing params")
	}

	decode, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	ad := AuthData{}
	if err = json.Unmarshal(decode, &ad); err != nil {
		return nil, err
	}

	return &ad, nil
}

func ParseAndDecryptProtectedFields(key string, lastLoggedOut int64, lastPasswordUpdated int64, protectedFields string, cookieExpiry int64) (string, string, error) {
	pf, err := decodeSecureData([]byte(key), protectedFields)
	if err != nil {
		return "", "Tampering", err
	}

	now := time.Now().UTC().Unix()

	if now > pf.ExpAt {
		return "", "ExpiredKey", ErrExpired
	}

	cookieCreatedAt := pf.ExpAt - cookieExpiry
	if cookieCreatedAt < lastLoggedOut || cookieCreatedAt < lastPasswordUpdated {
		return "", "CookieInvalid", ErrExpired
	}
	return pf.Email, "", nil
}

func createSecureData(key []byte, pf ProtectedFields) (string, error) {
	s := securecookie.New(key, key)
	s = s.SetSerializer(securecookie.JSONEncoder{})
	str, er := s.Encode(C.GetFactorsCookieName(), pf)
	return str, er
}

func decodeSecureData(key []byte, value string) (ProtectedFields, error) {
	s := securecookie.New(key, key)
	s = s.SetSerializer(securecookie.JSONEncoder{})
	pf := ProtectedFields{}
	err := s.Decode(C.GetFactorsCookieName(), value, &pf)
	return pf, err
}
