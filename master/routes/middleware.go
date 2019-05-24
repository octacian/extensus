package routes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/octacian/extensus/master/core"
	"github.com/octacian/extensus/master/models"
	log "github.com/sirupsen/logrus"
)

// authorized returns true if the request is authorized. Any unhandled errors
// are returned. In certain circumstances a specific HTTP status is also
// returned.
func authorized(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	user, status, err := func() (*models.User, int, error) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				return nil, 0, nil // Unsuccessful.
			}
			return nil, http.StatusBadRequest, err // Error occured.
		}

		rawToken := cookie.Value
		if rawToken == "" {
			return nil, 0, nil // Unsuccessful.
		}

		claims := &AuthenticationClaims{}
		token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("jwt: invalid signing method")
			}
			return []byte(core.GetConfig().Secret), nil
		})

		if err != nil {
			return nil, 0, err // Error occured.
		}

		if claims, ok := token.Claims.(*AuthenticationClaims); ok {
			if token.Valid {
				cached, err := models.Cache(&models.User{}, int(claims.ID))
				if err != nil {
					return nil, http.StatusBadRequest, err
				}
				user, ok := cached.(*models.User)
				if !ok {
					return nil, 0, errors.New("authenticate: failed perform type assertion on received claims")
				}

				return user, 0, nil // Successful.
			}
			return nil, 0, nil // Unsuccessful.
		}

		return nil, http.StatusBadRequest, nil // Error occurred.
	}()

	if user != nil {
		return user, nil // Successful.
	}
	if user == nil && status == 0 && err == nil {
		return nil, nil
	}

	if err != nil {
		if status == 0 {
			status = http.StatusInternalServerError
		}
		http.Error(w, fmt.Sprintf("authenticate: got error: %s", err.Error()), status)
		return nil, err // Error occurred.
	}

	if status != 0 {
		w.WriteHeader(status) // Error occurred.
		return nil, fmt.Errorf("authenticate: got illegal return status of %d but no error message", status)
	}

	return nil, fmt.Errorf("authenticate: received empty user (%v), response status (%d), and error (%s)",
		user, status, err)
}

// NoAuthorization ensures that requests do not contain valid JWT tokens.
func NoAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := authorized(w, r); err != nil { // Error occurred.
			log.WithFields(log.Fields{"error": err.Error()}).Error("NoAuthorization failed with an unexpected error")
		} else if user == nil && err == nil { // Authentication unsuccessful, serve request.
			next.ServeHTTP(w, r)
		} else { // Authentication successful, redirect to dashboard.
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	})
}

// Authorization ensures that requests contain a valid JWT token.
func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := authorized(w, r); err != nil { // Error occurred.
			log.WithFields(log.Fields{"error": err.Error()}).Error("Authorization failed with an unexpected error")
		} else if user == nil && err == nil { // Authentication unsuccessful, redirect to login.
			http.Redirect(w, r, fmt.Sprintf("/?return=%s", r.RequestURI), http.StatusSeeOther)
		} else { // Authentication successful, serve request.
			newRequest := r.WithContext(user.NewContext(r.Context()))
			next.ServeHTTP(w, newRequest)
		}
	})
}
