package web

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

type App struct {
	staticHandler http.Handler
}

func New() *App {
	frontendDir, err := detectFrontendDir()
	if err != nil {
		panic(err)
	}

	return &App{
		staticHandler: http.FileServer(http.Dir(frontendDir)),
	}
}

func (a *App) LoginPage(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/login.html"
	a.staticHandler.ServeHTTP(w, r)
}

func (a *App) AdminPage(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/admin.html"
	a.staticHandler.ServeHTTP(w, r)
}

func (a *App) Static(w http.ResponseWriter, r *http.Request) {
	cleanPath := strings.TrimPrefix(r.URL.Path, "/static/")
	r.URL.Path = "/" + cleanPath
	a.staticHandler.ServeHTTP(w, r)
}

func detectFrontendDir() (string, error) {
	candidates := []string{
		"frontend",
		"../frontend",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("frontend directory not found; expected one of: %s", strings.Join(candidates, ", "))
}
