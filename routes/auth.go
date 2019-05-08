package routes

import (
	"net/http"
	"path"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/octacian/extensus/core"
	"github.com/octacian/extensus/core/models"
)

const (
	tmplLoginName  TmplName  = "login"  // path to login template
	tmplLoginTitle TmplTitle = "Log In" // title of login page

	tmplForgotName  TmplName  = "forgot"          // path to forgot password template
	tmplForgotTitle TmplTitle = "Forgot Password" // title of forgot password page
)

// SignIn renders the sign in page.
func SignIn(w http.ResponseWriter, r *http.Request) {
	returnAfter := r.URL.Query()["return"]
	RenderTemplate(w, tmplLoginName, tmplLoginTitle, Data{"Return": returnAfter, "Query": "?" + r.URL.RawQuery})
}

// AuthenticationClaims holds JWT claims information.
type AuthenticationClaims struct {
	ID uint64 `json:"id"`
	jwt.StandardClaims
}

// SignInPost handles sign in requests.
func SignInPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if user, err := models.AuthenticateUser(email, password); err != nil {
		if models.IsErrNoEntry(err) {
			RenderTemplate(w, tmplLoginName, tmplLoginTitle, Data{"Failed": true, "Email": email})
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		// Token will expire in five days from now
		expirationTime := time.Now().Add(time.Hour * 24 * 5)
		claims := &AuthenticationClaims{
			ID: user.ID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(core.GetConfig().Secret))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})

		redirect := "dashboard"
		if returnAfter, ok := r.URL.Query()["return"]; ok && len(returnAfter) > 0 && returnAfter[0] != "" {
			redirect = path.Join(r.URL.Host, returnAfter[0])
		}

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

// Logout removes the stored token and redirects to the sign in page.
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Unix(0, 0),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Forgot renders the forgot password page.
func Forgot(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, tmplForgotName, tmplForgotTitle, nil)
}

// ForgotPost handles forgot password requests.
func ForgotPost(w http.ResponseWriter, r *http.Request) {
	value := r.FormValue("email")
	res := models.ValidUserEmail.MatchString(value)
	RenderTemplate(w, tmplForgotName, tmplForgotTitle, Data{"Email": value, "Valid": res})
}
