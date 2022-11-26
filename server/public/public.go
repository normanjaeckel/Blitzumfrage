/*
Package public provides a http handler for the the public client files like
index.html and all assets (JS, CSS, ...).
*/
package public

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed files
var files embed.FS

func Files() http.Handler {
	return handler{}
}

type handler struct{}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filesHandler().ServeHTTP(w, r)
}

func filesHandler() http.Handler {
	root, err := fs.Sub(files, "files")
	if err != nil {
		panic("Error when getting subtree of embedded filesystem. " +
			"This should never ever happen.")
	}
	return http.FileServer(http.FS(root))
}
