package openapi

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:generate cp openapi.yml dist/swagger.yml
//go:embed all:dist
var files embed.FS

func FileServer(path string) http.Handler {
	fsys, err := fs.Sub(files, "dist")
	if err != nil {
		log.Fatal(err)
	}
	filesystem := http.FS(fsys)

	// serve files stripped of the ui prefix
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, path)

		// try if file exists at path, if not append .html
		_, err := filesystem.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			path = fmt.Sprintf("%s.html", path)
		}
		r.URL.Path = path
		http.FileServer(filesystem).ServeHTTP(w, r)
	})
}
