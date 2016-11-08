package web

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"

	log "github.com/Sirupsen/logrus"
)

func (handler WebserviceHandler) BasicAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authError := func() {
			w.Header().Set("WWW-Authenticate", "Basic realm")
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
		}

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			authError()
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			authError()
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || !handler.ValidateAdmin(pair[0], pair[1]) {
			authError()
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (handler WebserviceHandler) ValidateAdmin(username, password string) bool {
	if username == "admin" && password == handler.config.GetAdminPass() {
		return true
	}
	return false
}

func (handler WebserviceHandler) RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Warnf("%v", err)
				http.Error(w, http.StatusText(500)+": "+fmt.Sprintf("%v", err), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (handler WebserviceHandler) LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//useful to test the frontend locally, remove in prod
		w.Header().Add("Access-Control-Allow-Origin", "*")
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Infof("%s request to %q: time %v", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func (handler WebserviceHandler) StatsHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		ip := strings.Split(r.RemoteAddr, ":")[0]
		key := ip + r.UserAgent()
		ps := context.Get(r, "params").(httprouter.Params)
		term := ps.ByName("term")
		agent := r.UserAgent()
		if !strings.Contains(agent, "Google") && !strings.Contains(agent, "Scanbot") && agent != "" {
			go handler.stats.NotifyUser(&User{Ip: ip, Referer: r.Referer(),
				UserAgent: agent, LastUri: r.RequestURI}, key, term)
		}
	}
	return http.HandlerFunc(fn)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (handler WebserviceHandler) GzipJsonHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzr, r)
	}
	return http.HandlerFunc(fn)
}
