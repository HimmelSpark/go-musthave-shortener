package auth

import "net/http"

func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := Info{}

			cookie, err := r.Cookie(CookieName)
			if err == nil {
				info.CookiePresent = true
				if userID, ok := ParseCookieValue(cookie.Value, secret); ok {
					info.UserID = userID
					info.CookieValid = true
				}
			}

			if info.UserID == "" {
				info.UserID = NewUserID()
				IssueCookie(w, info.UserID, secret)
			}

			next.ServeHTTP(w, r.WithContext(WithInfo(r.Context(), info)))
		})
	}
}
