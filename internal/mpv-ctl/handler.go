package mpvctl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/blang/mpv"
)

type handler struct {
	mpvc *mpv.Client
}

func GetNewHandler(mpvc *mpv.Client) http.Handler {
	return middleHandler()((&handler{mpvc}).createRouter())
}

func (h *handler) createRouter() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/append", h.appendHandler)
	return router
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

func writeDefaultResponse(w *http.ResponseWriter) {
	if err := json.NewEncoder(*w).Encode(ResponseModel{true}); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
func writeError(w *http.ResponseWriter, msg string) {
	if err := json.NewEncoder(*w).Encode(ErrorResponse{msg, false}); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
func (h *handler) appendHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("path") {
		writeError(&w, "'path' param should be present")
		return
	}
	flag := mpv.LoadFileModeReplace
	if r.URL.Query().Has("flag") {
		switch r.URL.Query().Get("flag") {
		case mpv.LoadFileModeAppend:
			flag = mpv.LoadFileModeAppend
		case mpv.LoadFileModeAppendPlay:
			flag = mpv.LoadFileModeAppendPlay
		}
	}
	if err := h.mpvc.Loadfile(r.URL.Query().Get("path"), flag); err != nil {
		errorMsg := fmt.Sprintf("append error: %s", err.Error())
		log.Println(errorMsg)
		writeError(&w, errorMsg)
	} else {
		writeDefaultResponse(&w)
	}
}

func (h *handler) loadPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("path") {
		writeError(&w, "'path' param should be present")
		return
	}
	flag := mpv.LoadFileModeReplace
	if r.URL.Query().Has("flag") {
		switch r.URL.Query().Get("flag") {
		case mpv.LoadFileModeAppend:
			flag = mpv.LoadFileModeAppend
		case mpv.LoadFileModeAppendPlay:
			flag = mpv.LoadFileModeAppendPlay
		}
	}
	if err := h.mpvc.LoadList(r.URL.Query().Get("path"), flag); err != nil {
		errorMsg := fmt.Sprintf("load playlist error: %s", err.Error())
		log.Println(errorMsg)
		writeError(&w, errorMsg)
	} else {
		writeDefaultResponse(&w)
	}
}
