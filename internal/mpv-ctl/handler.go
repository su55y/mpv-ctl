package mpvctl

import (
	"encoding/json"
	"log"
	"net/http"
)

type handler struct {
	service *Service
}

func GetNewHandler(service *Service) http.Handler {
	return middleHandler()((&handler{service}).createRouter())
}

func (h *handler) createRouter() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/append", h.appendHandler)
	return router
}

func writeDefaultResponse(w *http.ResponseWriter) {
	if err := json.NewEncoder(*w).Encode(ResponseModel{true}); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
func writeError(w *http.ResponseWriter, msg string) {
	log.Println(msg)
	if err := json.NewEncoder(*w).Encode(ErrorResponse{msg, false}); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}

func middleHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			log.Printf("%s %s\n", r.Method, r.RequestURI)
			if r.Method != http.MethodGet {
				writeError(&w, "Method Not Allowed")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (h *handler) appendHandler(w http.ResponseWriter, r *http.Request) {
	h.service.LoadFile(r.URL.Query(), &w)
}

func (h *handler) loadPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	h.service.LoadPlaylist(r.URL.Query(), &w)
}
