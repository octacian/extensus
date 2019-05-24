package template

import (
	"html/template"

	"github.com/octacian/extensus/master/models"
	"github.com/octacian/extensus/shared"

	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
)

var (
	templates    = template.New("")
	templatePath = shared.Abs("templates")
)

type (
	// Data type used to pass data to templates with Render.
	Data map[string]interface{}

	// Name type represents the relative path to a template.
	Name string

	// Title type represents the title of a page.
	Title string
)

// GetName takes a path to a template file and returns a path relative to the
// top-level templates directory with the file extension removed.
func GetName(path string) string {
	name := strings.Replace(path, templatePath, "", 1)[1:]
	return name[:len(name)-5]
}

// Assert returns true if the path is a valid template. If an error occurs
// while opening the path or while fetching its FileInfo, panic is called.
func Assert(path string) bool {
	if !shared.GetFileInfo(path).IsDir() && filepath.Ext(path) == ".html" {
		return true
	}

	return false
}

// Parse ensures that the file located at the path specified is a valid
// template before parsing it into the loaded templates. If any errors occur
// panic is called. If the path is a directory or does not end with .html,
// nothing happens.
func Parse(path string) {
	if Assert(path) && filepath.Ext(path) == ".html" {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			log.Panicf("template.Parse: got error while reading %s:\n%s", path, err)
		}

		name := GetName(path)
		var tmpl *template.Template
		if tmpl = templates.Lookup(name); tmpl == nil {
			tmpl = templates.New(name)
		}

		if _, err := tmpl.Parse(string(contents)); err != nil {
			log.Panic("template.Parse: got error: ", err)
		}
	}
}

// ParseAll recursively parses all templates within the top-level template
// directory and its sub-directories. If any errors occur panic is called.
func ParseAll() {
	if err := filepath.Walk(templatePath, func(path string, _ os.FileInfo, err error) error {
		Parse(path)
		return nil
	}); err != nil {
		log.Panic("template.ParseAll: got error:\n", err)
	}
}

// Render renders a template given its name and an arbitrary title.
func Render(w http.ResponseWriter, r *http.Request, tmpl Name, title Title, data Data) {
	if data == nil {
		data = Data{}
	}

	data["Title"] = title

	if user, ok := models.UserFromContext(r.Context()); ok {
		data["User"] = user
	}

	if err := templates.ExecuteTemplate(w, string(tmpl), data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// WatchAll watches the templates directory for changes and re-executes
// templates when they change. If any errors occur panic is called.
func WatchAll() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panic("template.WatchAll: got error while creating watcher:\n", err)
	}
	defer watcher.Close()

	if err := filepath.Walk(templatePath, func(path string, file os.FileInfo, err error) error {
		if file.IsDir() {
			return watcher.Add(path)
		}

		return nil
	}); err != nil {
		log.Panic("template.WatchAll: got error while walking templates directory:\n", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case _ = <-watcher.Events:
				log.Info("Templates changed, parsing again")
				templates = template.New("")
				ParseAll()
			case err := <-watcher.Errors:
				log.Warn("template.WatchAll: got error:", err)
			}
		}
	}()

	log.Info("Started template watcher")
	<-done
}
