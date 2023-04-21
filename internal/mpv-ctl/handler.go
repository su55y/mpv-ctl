package mpvctl

import (
	"encoding/json"
	"log"
	"net/http"
)

type handler struct {
	service IHTTPService
}

func GetNewHandler(service IHTTPService) http.Handler {
	return middleHandler()((&handler{service}).createRouter())
}

func (h *handler) createRouter() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/append", h.appendHandler)
	router.HandleFunc("/playlist", h.loadPlaylistHandler)
	router.HandleFunc("/control", h.controlHandler)
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
	q := r.URL.Query()
	q.Set("type", "video")
	h.service.LoadFileHttpHandler(q, &w)
}

func (h *handler) loadPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Set("type", "playlist")
	h.service.LoadFileHttpHandler(q, &w)
}

func (h *handler) controlHandler(w http.ResponseWriter, r *http.Request) {
	h.service.ControlHttpHandler(r.URL.Query(), &w)
}
