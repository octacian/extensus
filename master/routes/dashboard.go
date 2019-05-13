package routes

import (
	"net/http"

	"github.com/octacian/extensus/master/template"
)

const (
	tmplDashboardName  template.Name  = "dashboard" // path to login template
	tmplDashboardTitle template.Title = "Dashboard" // title of login page
)

// Dashboard renders the dashboard page.
func Dashboard(w http.ResponseWriter, r *http.Request) {
	template.Render(w, tmplDashboardName, tmplDashboardTitle, template.Data{
		"List": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	})
}
