package routes

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/octacian/extensus/master/core"
	"github.com/octacian/extensus/master/template"
	"github.com/octacian/extensus/shared"

	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// Serve starts the HTTP server. If any errors occur, Serve panics.
func Serve() {
	template.ParseAll()
	if os.Getenv("MODE") == "DEV" {
		go template.WatchAll()
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/", func(router chi.Router) {
		router.Group(func(router chi.Router) {
			router.Use(NoAuthorization)
			router.Get("/", SignIn)
			router.Post("/", SignInPost)
			router.Get("/forgot", Forgot)
			router.Post("/forgot", ForgotPost)
		})

		router.Group(func(router chi.Router) {
			router.Use(Authorization)
			router.Get("/logout", Logout)
			router.Get("/dashboard", Dashboard)
			router.Get("/nodes", Nodes)
		})
	})

	ServeFiles(router)

	address := core.GetConfig().Address
	log.WithFields(log.Fields{"address": address}).Info("HTTP server listening")
	log.Fatal(http.ListenAndServe(address, router))
}

// ServeFiles starts a http.FileServer to serve static files from public.
func ServeFiles(router chi.Router) {
	fs := http.StripPrefix("/public", http.FileServer(http.Dir(shared.Abs("public"))))

	router.Get("/public", http.RedirectHandler("/public/", 301).ServeHTTP)

	router.Get("/public/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
