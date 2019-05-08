package routes

import "net/http"

const (
	tmplDashboardName  TmplName  = "dashboard" // path to login template
	tmplDashboardTitle TmplTitle = "Dashboard" // title of login page
)

// Dashboard renders the dashboard page.
func Dashboard(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, tmplDashboardName, tmplDashboardTitle, nil)
}
