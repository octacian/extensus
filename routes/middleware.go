package routes

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/octacian/extensus/core"
)

// authorized returns true if the request is authorized. Any unhandled errors
// are returned. In certain circumstances a specific HTTP status is also
// returned.
func authorized(r *http.Request) (bool, int, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return false, 0, nil
		}
		return false, http.StatusBadRequest, err
	}

	rawToken := cookie.Value
	if rawToken == "" {
		return false, 0, nil
	}

	claims := &AuthenticationClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: invalid signing method")
		}
		return []byte(core.GetConfig().Secret), nil
	})

	if err != nil {
		return false, http.StatusInternalServerError, err
	}

	if _, ok := token.Claims.(*AuthenticationClaims); ok {
		if token.Valid {
			return true, 0, nil
		}
		return false, 0, nil
	}

	return false, http.StatusBadRequest, nil
}

// NoAuthorization ensures that requests do not contain valid JWT tokens.
func NoAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowed, status, err := authorized(r); !allowed {
			next.ServeHTTP(w, r)
		} else if err != nil {
			if status == 0 {
				status = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), status)
		} else if status != 0 {
			w.WriteHeader(status)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	})
}

// Authorization ensures that requests contain a valid JWT token.
func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowed, status, err := authorized(r); allowed {
			next.ServeHTTP(w, r)
		} else if err != nil {
			if status == 0 {
				status = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), status)
		} else if status != 0 {
			w.WriteHeader(status)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/?return=%s", r.RequestURI), http.StatusSeeOther)
		}
	})
}
