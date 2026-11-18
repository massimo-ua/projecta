package web

import (
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
)

type noDirListFS struct {
	fs http.FileSystem
}

func (n noDirListFS) Open(name string) (http.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if stat.IsDir() {
		indexFile, err := n.fs.Open(path.Join(name, "index.html"))
		if err != nil {
			_ = f.Close()
			return nil, os.ErrPermission
		}
		_ = indexFile.Close()
	}
	return f, nil
}

func createSwaggerHandler(r *mux.Router) {
	r.PathPrefix("/swagger").Handler(http.StripPrefix("/swagger", http.FileServer(noDirListFS{fs: http.Dir("swagger")})))
}
