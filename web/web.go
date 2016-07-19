package web

import (
	"mime/multipart"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

type Repository interface {
	//GetAllWords(lang1 string, lang2 string) (words []*SimpleWord, err error)
	ResetDB() error
	GetLangFromKey(key string) string
	Search(word string, fromLang string, toLang string) (words []*Word, err error)
	GetWordsWithTerm(term string, lang1 string, lang2 string) (words []*SimpleWord, err error)
	GetLanguages() []string
}

type ServerConfig interface {
	GetServerUrl() string
	GetHTTPDir() string
	GetAdminPass() string
	GetServerPort() string
}

type SysUtils interface {
	ExtractZipToHttpDir(file multipart.File, length int64) error
	CopyFileToExcelDir(file multipart.File) error
}

type WebserviceHandler struct {
	repo    Repository
	frouter *http.ServeMux
	mrouter *router
	config  ServerConfig
	sutils  SysUtils
}

type router struct {
	*httprouter.Router
}

func NewWebHandler(repo Repository, c ServerConfig, s SysUtils) *WebserviceHandler {
	return &WebserviceHandler{repo: repo, config: c, sutils: s}
}

func (h WebserviceHandler) StartServer() {
	commonHandlers := alice.New(context.ClearHandler, h.LoggingHandler, h.RecoverHandler)
	h.mrouter = NewRouter()
	h.frouter = http.NewServeMux()
	h.frouter.Handle("/", http.FileServer(http.Dir(h.config.GetHTTPDir())))

	h.mrouter.Get("/services/autocomplete/:langkey", commonHandlers.ThenFunc(h.Autocomplete))
	h.mrouter.Post("/services/deployFront", commonHandlers.Append(h.BasicAuth).ThenFunc(h.DeployFront))
	h.mrouter.Post("/services/deployDb", commonHandlers.Append(h.BasicAuth).ThenFunc(h.DeployDb))
	h.mrouter.Get("/search/:langkey/:term", commonHandlers.ThenFunc(h.IndexHTML))
	h.mrouter.Get("/", commonHandlers.ThenFunc(h.IndexHTML))

	var err error

	r := http.NewServeMux()
	r.HandleFunc("/", h.FrontHandler)
	go func() {
		log.Infof("Server launched on port %s", h.config.GetServerPort())
		err = http.ListenAndServe(":"+h.config.GetServerPort(), r)
		if err != nil {
			panic(err.Error())
		}
	}()

}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

func NewRouter() *router {
	return &router{httprouter.New()}
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}

func (handler WebserviceHandler) FrontHandler(w http.ResponseWriter, r *http.Request) {
	/*
		if strings.HasPrefix(r.URL.Path, "/services") {
			handler.mrouter.ServeHTTP(w, r)
		} else {
			handler.frouter.ServeHTTP(w, r)
		}
	*/
	if strings.HasPrefix(r.URL.Path, "/css") || strings.HasPrefix(r.URL.Path, "/js") ||
		strings.HasPrefix(r.URL.Path, "/deploy.html") {
		handler.frouter.ServeHTTP(w, r)
	} else {
		handler.mrouter.ServeHTTP(w, r)
	}
}
