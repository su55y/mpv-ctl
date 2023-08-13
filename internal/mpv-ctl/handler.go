package mpvctl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type IHTTPService interface {
	LoadFileHttpHandler(url.Values, *http.ResponseWriter)
	ControlHttpHandler(url.Values, *http.ResponseWriter)
	SetProperty(string, string) error
	GetProperty(string) (string, error)
}

type handler struct {
	service IHTTPService
}

func NewHandler(service IHTTPService) http.Handler {
	return middleHandler()((&handler{service}).createRouter())
}

func (h *handler) createRouter() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/append", h.appendHandler)
	router.HandleFunc("/playlist", h.loadPlaylistHandler)
	router.HandleFunc("/control", h.controlHandler)
	router.HandleFunc("/set", h.setPropertyHandler)
	router.HandleFunc("/get", h.getPropertyHandler)
	return router
}

func writeResponse(w *http.ResponseWriter, resp interface{}) {
	if err := json.NewEncoder(*w).Encode(resp); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}

func writeDefaultResponse(w *http.ResponseWriter) {
	writeResponse(w, ResponseModel{true})
}

func writePropertyResponse(w *http.ResponseWriter, resp PropertyResponse) {
	writeResponse(w, resp)
}

func writeError(w *http.ResponseWriter, msg string) {
	log.Println(msg)
	writeResponse(w, ErrorResponse{msg, false})
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

func (h *handler) getPropertyHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("name") {
		writeError(&w, fmt.Sprintf("missing 'name' parameter"))
		return
	}
	var key string
	if key := query.Get("name"); len(key) == 0 {
		writeError(&w, fmt.Sprintf("invalid property name %+s", key))
		return
	}
	prop, err := h.service.GetProperty(key)
	if err != nil {
		writeError(&w, fmt.Sprintf("can't get property %+s: %v", key, err))
		return
	}
	writePropertyResponse(&w, PropertyResponse{true, key, prop})
}

func (h *handler) setPropertyHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("name") {
		writeError(&w, fmt.Sprintf("missing 'name' parameter"))
		return
	}
	if !query.Has("value") {
		writeError(&w, fmt.Sprintf("missing 'value' parameter"))
		return
	}
	var key, prop string
	if key := query.Get("name"); len(key) == 0 {
		writeError(&w, fmt.Sprintf("invalid property 'name' %+s", key))
		return
	}
	if prop = query.Get("value"); len(prop) == 0 {
		writeError(&w, fmt.Sprintf("invalid property value %+s", prop))
		return
	}
	if err := h.service.SetProperty(key, prop); err != nil {
		writeError(&w, fmt.Sprintf("set property %+s (%+v) error: %v", key, prop, err))
		return
	}
	writeDefaultResponse(&w)
}
