package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strings"

	//log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

type HtmlContent struct {
	Languages []string
	Results   []*Word
}

func (handler WebserviceHandler) IndexHTML(w http.ResponseWriter, r *http.Request) {
	ps := context.Get(r, "params").(httprouter.Params)
	key := ps.ByName("langkey")
	term := ps.ByName("term")
	funcMap := template.FuncMap{
		"langsToKey": func(l1, l2 string) string {
			return strings.ToLower(l1[:3]) + strings.ToLower(l2[:3])
		},
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
	}
	t := template.Must(template.New("index.html").Funcs(funcMap).ParseFiles(handler.config.GetHTTPDir() + "index.html"))
	languages := handler.repo.GetLanguages()
	content := &HtmlContent{Languages: languages}
	if key != "" && term != "" {
		fromLang, toLang := handler.getLanguagesFromRequest(ps.ByName("langkey"))
		results, err := handler.repo.Search(term, fromLang, toLang)
		if err != nil {
			//trying the other way around
			results, err = handler.repo.Search(term, toLang, fromLang)
			if err != nil {
				panic(fmt.Sprintf("Word %s does not exist in %s/%s dictionary", term, fromLang, toLang))
			} else {
				http.Redirect(w, r, "/search/"+toLang[0:3]+fromLang[0:3]+"/"+term, http.StatusFound)
			}
		}
		content.Results = results
	}

	t.Execute(w, content)
}

/*
func (handler WebserviceHandler) Search(w http.ResponseWriter, r *http.Request) {
	fromLang, toLang := handler.getLanguagesFromRequest(r)
	word := r.FormValue("word")
	if word == "" {
		panic("Empty word")
	}
	words, err := handler.repo.Search(word, fromLang, toLang)
	if err != nil {
		panic(err.Error())
	}
	for _, wd := range words {
		for _, tr := range wd.Translations {
			fmt.Fprintln(w, tr.Word)
		}
	}
}
*/

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
