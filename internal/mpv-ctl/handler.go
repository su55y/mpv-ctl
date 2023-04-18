package mpvctl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

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

type ParamsType1 struct {
	url  string
	flag string
}

func parseParamsType1(query url.Values) (*ParamsType1, error) {
	if !query.Has("url") {
		return nil, fmt.Errorf("'url' param should be present")
	}
	flag := mpv.LoadFileModeReplace
	if query.Has("flag") {
		switch query.Get("flag") {
		case mpv.LoadFileModeAppend:
			flag = mpv.LoadFileModeAppend
		case mpv.LoadFileModeAppendPlay:
			flag = mpv.LoadFileModeAppendPlay
		}
	}
	return &ParamsType1{query.Get("url"), flag}, nil
}

func (h *handler) appendHandler(w http.ResponseWriter, r *http.Request) {
	params, err := parseParamsType1(r.URL.Query())
	if err != nil {
		writeError(&w, err.Error())
		return
	}
	if err := h.mpvc.Loadfile(params.url, params.flag); err != nil {
		errorMsg := fmt.Sprintf("append error: %s", err.Error())
		log.Println(errorMsg)
		writeError(&w, errorMsg)
	} else {
		writeDefaultResponse(&w)
	}
}

func (h *handler) loadPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	params, err := parseParamsType1(r.URL.Query())
	if err != nil {
		writeError(&w, err.Error())
		return
	}
	if err := h.mpvc.LoadList(params.url, params.flag); err != nil {
		errorMsg := fmt.Sprintf("load playlist error: %s", err.Error())
		log.Println(errorMsg)
		writeError(&w, errorMsg)
	} else {
		writeDefaultResponse(&w)
	}
}
