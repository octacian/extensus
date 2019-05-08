package routes

import (
	"html/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/octacian/extensus/core"

	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
)

var templates = template.New("")
var templatePath = core.Abs(filepath.Join("public", "templates"))

type (
	// Data type used to pass to RenderTemplate.
	Data map[string]interface{}

	// TmplName type represents the relative path to a template.
	TmplName string

	// TmplTitle type represents the title of a page.
	TmplTitle string
)

// GetTemplateName takes a path and returns the simplest, unique template name
// possible.
func GetTemplateName(path string) string {
	name := strings.Replace(path, templatePath, "", 1)[1:]
	return name[:len(name)-5]
}

// IsTemplate returns true if the path is a valid template. If an error occurs
// while opening the path or while fetching its FileInfo, panic is called.
func IsTemplate(path string) bool {
	if !core.GetFileInfo(path).IsDir() && filepath.Ext(path) == ".html" {
		return true
	}

	return false
}

// ParseTemplate ensures that the file located at the path specified is a valid
// template before parsing it into the loaded templates. If any errors occur
// panic is called. If the path is a directory or does not end with .html,
// nothing happens.
func ParseTemplate(path string) {
	if IsTemplate(path) && filepath.Ext(path) == ".html" {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			log.Panicf("ParseTemplate: got error while reading %s:\n%s", path, err)
		}

		name := GetTemplateName(path)
		var tmpl *template.Template
		if tmpl = templates.Lookup(name); tmpl == nil {
			tmpl = templates.New(name)
		}

		if _, err := tmpl.Parse(string(contents)); err != nil {
			log.Panic("ParseTemplate: got error: ", err)
		}
	}
}

// ParseAllTemplates recursively parses all templates within public/templates
// and its sub-directories. If any errors occur panic is called.
func ParseAllTemplates() {
	if err := filepath.Walk(templatePath, func(path string, _ os.FileInfo, err error) error {
		ParseTemplate(path)
		return nil
	}); err != nil {
		log.Panic("ParseAllTemplates: got error:\n", err)
	}
}

// RenderTemplate renders a template by name.
func RenderTemplate(w http.ResponseWriter, tmpl TmplName, title TmplTitle, data Data) {
	if data == nil {
		data = Data{}
	}

	data["Title"] = title

	if err := templates.ExecuteTemplate(w, string(tmpl), data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// WatchTemplates watches the templates directory for changes and re-executes
// templates when they change. If any errors occur panic is called.
func WatchTemplates() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panic("WatchTemplates: got error while creating watcher:\n", err)
	}
	defer watcher.Close()

	if err := filepath.Walk(templatePath, func(path string, file os.FileInfo, err error) error {
		if file.IsDir() {
			return watcher.Add(path)
		}

		return nil
	}); err != nil {
		log.Panic("WatchTemplates: got error while walking templates directory:\n", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case _ = <-watcher.Events:
				log.Info("Templates changed, parsing again")
				templates = template.New("")
				ParseAllTemplates()
			case err := <-watcher.Errors:
				log.Warn("WatchTemplates: got error:", err)
			}
		}
	}()

	log.Info("Started template watcher")
	<-done
}

// Serve starts the HTTP server. If any errors occur, Serve panics.
func Serve() {
	ParseAllTemplates()
	if os.Getenv("MODE") == "DEV" {
		go WatchTemplates()
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
		})
	})

	ServeFiles(router)

	address := core.GetConfig().Address
	log.WithFields(log.Fields{"address": address}).Info("HTTP server listening")
	log.Fatal(http.ListenAndServe(address, router))
}

// ServeFiles starts a http.FileServer to serve static files from public.
func ServeFiles(router chi.Router) {
	fs := http.StripPrefix("/public", http.FileServer(http.Dir(core.Abs("public"))))

	router.Get("/public", http.RedirectHandler("/public/", 301).ServeHTTP)

	router.Get("/public/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
