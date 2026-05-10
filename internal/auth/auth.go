package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	CookieName  = "auth_token"
	userIDLen   = 21
	cookieSep   = "|"
	cookiePath  = "/"
	maxCookiePartCount = 2
)

func Sign(userID, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	return hex.EncodeToString(mac.Sum(nil))
}

func Verify(userID, signature, secret string) bool {
	got, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	expected := hmac.New(sha256.New, []byte(secret))
	expected.Write([]byte(userID))
	return hmac.Equal(expected.Sum(nil), got)
}

func ParseCookieValue(value, secret string) (string, bool) {
	parts := strings.SplitN(value, cookieSep, maxCookiePartCount+1)
	if len(parts) != maxCookiePartCount {
		return "", false
	}
	if parts[0] == "" {
		return "", false
	}
	if !Verify(parts[0], parts[1], secret) {
		return "", false
	}
	return parts[0], true
}

func NewUserID() string {
	id, err := gonanoid.New(userIDLen)
	if err != nil {
		panic(err)
	}
	return id
}

func IssueCookie(w http.ResponseWriter, userID, secret string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    userID + cookieSep + Sign(userID, secret),
		Path:     cookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
