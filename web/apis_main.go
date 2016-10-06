package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
	"github.com/beppeben/go-dictionary/utils"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

type HtmlContent struct {
	Languages  []*Language
	Results    []*Word
	Fields     []string
	FieldDescs []string
}

var htmlHelpers = template.FuncMap{
	"oddOrEven": func(num int) string {
		if math.Mod(float64(num), 2) != 0 {
			return "odd"
		} else {
			return "even"
		}
	},
	"toUpper": func(text string) string {
		return strings.ToUpper(text[0:1]) + text[1:]
	},
	"dec": func(num int) int {
		return num - 1
	},
}

func (handler WebserviceHandler) IndexHTML(w http.ResponseWriter, r *http.Request) {
	baseLang := handler.getBaseLanguage(r.FormValue("lang"))
	ps := context.Get(r, "params").(httprouter.Params)
	key := ps.ByName("langkey")
	term := ps.ByName("term")
	htmlHelpers["getString"] = func(key string) string {
		return handler.repo.GetWebTerm(baseLang, key)
	}
	t := template.Must(template.New("index.html").Funcs(htmlHelpers).ParseFiles(handler.config.GetHTTPDir() + "index.html"))
	langs := handler.repo.GetLanguages(baseLang)
	content := &HtmlContent{Languages: langs}
	if key != "" && term != "" {
		fromLang, toLang := handler.getLanguagesFromRequest(ps.ByName("langkey"))
		results, err := handler.repo.Search(term, fromLang, toLang, baseLang)
		if err != nil {
			//trying the other way around
			results, err = handler.repo.Search(term, toLang, fromLang, baseLang)
			if err != nil {
				panic(fmt.Sprintf("Word %s does not exist in %s/%s dictionary", term, fromLang, toLang))
			} else {
				url := "/search/" + toLang[:3] + fromLang[:3] + "/" + term
				if baseLang != "" {
					url += "?lang=" + baseLang[:3]
				}
				http.Redirect(w, r, url, http.StatusFound)
			}
		}
		content.Results = results
		//list of (non repeating) fields for all the words
		for _, word := range results {
			if !utils.Contains(content.Fields, word.Field) {
				content.Fields = append(content.Fields, word.Field)
				content.FieldDescs = append(content.FieldDescs, word.FieldDesc)
			}
		}
	}

	t.Execute(w, content)
}

func (handler WebserviceHandler) AboutHTML(w http.ResponseWriter, r *http.Request) {
	baseLang := handler.getBaseLanguage(r.FormValue("lang"))
	htmlHelpers["getString"] = func(key string) string {
		return handler.repo.GetWebTerm(baseLang, key)
	}
	htmlHelpers["getHtml"] = func(key string) template.HTML {
		return template.HTML(handler.repo.GetWebTerm(baseLang, key))
	}
	t := template.Must(template.New("about.html").Funcs(htmlHelpers).ParseFiles(handler.config.GetHTTPDir() + "about.html"))

	t.Execute(w, "")
}

func (handler WebserviceHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	ps := context.Get(r, "params").(httprouter.Params)
	fromLang, toLang := handler.getLanguagesFromRequest(ps.ByName("langkey"))
	term := strings.ToLower(r.FormValue("term"))
	result, err := handler.repo.GetWordsWithTerm(term, fromLang, toLang)
	if err != nil {
		panic(err.Error())
	}
	enc := json.NewEncoder(w)
	enc.Encode(result)
}

func (handler WebserviceHandler) Notify(w http.ResponseWriter, r *http.Request) {
	fromLang, toLang := handler.getLanguagesFromRequest(r.FormValue("langkey"))
	term := r.FormValue("word")
	if term == "" {
		panic("No word inserted")
	}
	message := "Word: " + term + "\nDictionary: " + fromLang + "-" + toLang
	go func() {
		err := handler.eutils.SendEmailToAdmin("Suggestion received", message)
		if err != nil {
			log.Info(err.Error())
		}
	}()
}

func (handler WebserviceHandler) getLanguagesFromRequest(key string) (string, string) {
	if len(key) != 6 {
		panic("Invalid key")
	}
	fromLang := handler.repo.GetLangFromKey(key[0:3])
	toLang := handler.repo.GetLangFromKey(key[3:6])
	if fromLang == "" || toLang == "" {
		panic("Invalid language keys")
	}
	return fromLang, toLang
}

func (handler WebserviceHandler) getBaseLanguage(key string) string {
	baseLang := "english"
	if key != "" {
		baseLang = handler.repo.GetLangFromKey(key)
		if baseLang == "" {
			panic("Invalid language")
		}
	}
	return baseLang
}
