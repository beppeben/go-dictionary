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
	ResetDB() error
	ResetCalendar() error
	GetCalendarEvents(month int, year int) (events []*CalendarEvent, err error)
	GetLangFromKey(key string) string
	Search(word, fromLang, toLang, baseLang string) (words []*Word, err error)
	GetWordsWithTerm(term string, lang1 string, lang2 string) (words []*SimpleWord, err error)
	GetLanguages(base string) []*Language
	GetWebTerm(lang, key string) string
}

type ServerConfig interface {
	GetHTTPDir() string
	GetAdminPass() string
	GetServerPort() string
}

type SysUtils interface {
	ExtractZipToHttpDir(file multipart.File, length int64) error
	CopyFileToExcelDir(file multipart.File, name string) error
}

type MessageUtils interface {
	//SendEmail(email, title, text string) error
	//SendEmailToAdmins(subject, body string) error
	SendToSlack(text string) error
}

type WebserviceHandler struct {
	repo     Repository
	frouter  *http.ServeMux
	mrouter  *router
	config   ServerConfig
	sutils   SysUtils
	msgutils MessageUtils
	stats    *StatsTracker
}

type router struct {
	*httprouter.Router
}

func NewWebHandler(repo Repository, c ServerConfig, s SysUtils, e MessageUtils) *WebserviceHandler {
	tracker := NewStatsTracker(e)
	return &WebserviceHandler{repo: repo, config: c, sutils: s, msgutils: e, stats: tracker}
}

func (h WebserviceHandler) StartServer() {
	commonHandlers := alice.New(context.ClearHandler, h.LoggingHandler, h.StatsHandler, h.RecoverHandler)
	commonHandlersNoStats := alice.New(context.ClearHandler, h.LoggingHandler, h.RecoverHandler)
	h.mrouter = NewRouter()
	h.frouter = http.NewServeMux()
	h.frouter.Handle("/", http.FileServer(http.Dir(h.config.GetHTTPDir())))

	h.mrouter.Get("/services/autocomplete/:langkey", commonHandlersNoStats.ThenFunc(h.Autocomplete))
	h.mrouter.Post("/services/deployFront", commonHandlersNoStats.Append(h.BasicAuth).ThenFunc(h.DeployFront))
	h.mrouter.Post("/services/deployDb", commonHandlersNoStats.Append(h.BasicAuth).ThenFunc(h.DeployDb))
	h.mrouter.Post("/services/deployCal", commonHandlersNoStats.Append(h.BasicAuth).ThenFunc(h.DeployCal))
	h.mrouter.Get("/services/notify", commonHandlersNoStats.ThenFunc(h.Notify))
	h.mrouter.Get("/search/:langkey/:term", commonHandlers.ThenFunc(h.IndexHTML))
	h.mrouter.Get("/calendar", commonHandlers.ThenFunc(h.CalendarHTMLDefault))
	h.mrouter.Get("/calendar/:year/:month", commonHandlers.ThenFunc(h.CalendarHTML))
	h.mrouter.Get("/index.html", commonHandlers.ThenFunc(h.IndexHTML))
	h.mrouter.Get("/terms.html", commonHandlers.ThenFunc(h.TermsHTML))
	h.mrouter.Get("/about.html", commonHandlers.ThenFunc(h.AboutHTML))
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
	if strings.HasPrefix(r.URL.Path, "/css") || strings.HasPrefix(r.URL.Path, "/js") ||
		strings.HasPrefix(r.URL.Path, "/deploy.html") || strings.HasPrefix(r.URL.Path, "/media") ||
		strings.HasPrefix(r.URL.Path, "/google") {

		// serve static files
		handler.frouter.ServeHTTP(w, r)
	} else {
		// serve dynamic files
		handler.mrouter.ServeHTTP(w, r)
	}
}
