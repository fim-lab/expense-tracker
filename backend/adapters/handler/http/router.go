package http

import (
	"io/fs"
	"net/http"
	"strings"
)

func NewRouter(adapter *Adapter, staticFiles fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adapter.GetTransactionsHandler(w, r)
		case http.MethodPost:
			adapter.CreateTransactionHandler(w, r)
		case http.MethodDelete:
			adapter.DeleteTransactionHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// serves frontend (static)
	fileServer := http.FileServer(http.FS(staticFiles))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}