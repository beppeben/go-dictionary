package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
	"github.com/beppeben/go-dictionary/utils"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

var MONTHS = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

type HtmlContent struct {
	Languages   []*Language
	Results     []*Word
	Fields      []string
	FieldDescs  []string
	BaseLangTag string
}

type CalendarDay struct {
	Day     int
	Active  bool
	IsToday bool
}

type CalendarEventHtml struct {
	Id             int
	Tag            string
	Title          string
	Description    string
	Level          int
	Row            int
	Column         int
	Span           int
	IsContinuation bool
}

type CalendarContent struct {
	Month     string
	Year      int
	PrevMonth int
	PrevYear  int
	NextMonth int
	NextYear  int
	Days      []CalendarDay
	Events    []CalendarEventHtml
}

func (handler WebserviceHandler) getHelpers(baseLang string) template.FuncMap {
	return template.FuncMap{
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
		"getString": func(key string) string {
			return handler.repo.GetWebTerm(baseLang, key)
		},
		"getHtml": func(key string) template.HTML {
			return template.HTML(handler.repo.GetWebTerm(baseLang, key))
		},
	}
}

func (handler WebserviceHandler) CalendarHTMLDefault(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	url := "/calendar/" + strconv.Itoa(now.Year()) + "/" + strconv.Itoa(int(now.Month()))
	http.Redirect(w, r, url, http.StatusFound)
}

func positive(a float64) int {
	if a > 0 {
		return int(a)
	} else {
		return 0
	}
}

func (handler WebserviceHandler) CalendarHTML(w http.ResponseWriter, r *http.Request) {
	baseLang := handler.getBaseLanguage(r.FormValue("lang"))
	ps := context.Get(r, "params").(httprouter.Params)
	yearStr := ps.ByName("year")
	monthStr := ps.ByName("month")
	log.Infof("Year %s month %s", yearStr, monthStr)
	yearNum, err := strconv.Atoi(yearStr)
	if err != nil {
		panic("Bad format")
	}
	monthNum, err := strconv.Atoi(monthStr)
	if err != nil {
		panic("Bad format")
	}

	htmlHelpers := handler.getHelpers(baseLang)
	t := template.Must(template.New("calendar.html").Funcs(htmlHelpers).ParseFiles(handler.config.GetHTTPDir() + "calendar.html"))

	// compute previus/next month/year
	nextMonth := monthNum + 1
	nextYear := yearNum
	if nextMonth > 12 {
		nextMonth = 1
		nextYear = yearNum + 1
	}
	prevMonth := monthNum - 1
	prevYear := yearNum
	if prevMonth < 1 {
		prevMonth = 12
		prevYear = yearNum - 1
	}

	// compute days in current month
	d, err := time.Parse("2/1/2006", "1/"+monthStr+"/"+yearStr)
	if err != nil {
		panic(err)
	}
	start_lag := (int(d.Weekday()) + 6) % 7
	first_day := d.AddDate(0, 0, -start_lag)
	days := make([]CalendarDay, 0)
	i := start_lag
	for ; i > 0; i-- {
		day := CalendarDay{Day: d.AddDate(0, 0, -i).Day(), Active: false}
		days = append(days, day)
	}
	now := time.Now()
	for ; d.AddDate(0, 0, i).Month() == d.Month(); i++ {
		day := CalendarDay{Day: i + 1, Active: true}
		if now.Day() == day.Day && now.Month() == d.Month() {
			day.IsToday = true
		}
		days = append(days, day)
	}
	end_lag := (7 - (len(days) % 7)) % 7
	for i = 0; i < end_lag; i++ {
		day := CalendarDay{Day: i + 1, Active: false}
		days = append(days, day)
	}
	max_row := (len(days)-1)/7 + 2

	// place events
	events, err := handler.repo.GetCalendarEvents(monthNum, yearNum)
	if err != nil {
		panic(err)
	}
	log.Infof("%d events", len(events))
	html_events := make([]CalendarEventHtml, 0)
	counter := 1
	for i, _ := range events {
		html_event := CalendarEventHtml{Id: counter, Tag: events[i].Tag, Title: events[i].Title, Description: events[i].Description}
		counter = counter + 1
		for j := 0; j < i; j++ {
			if !events[j].EndDate.Before(events[i].StartDate) && !events[j].StartDate.After(events[i].EndDate) {
				html_event.Level = html_event.Level + 1
			}
		}
		/*
			html_event.Row = (events[i].StartDate.Day()+start_lag-1)/7 + 2
			html_event.Column = (events[i].StartDate.Day()+start_lag-1)%7 + 1
			html_event.Span = int(events[i].EndDate.Sub(events[i].StartDate).Hours()/24) + 1
		*/
		html_event.Row = (positive(events[i].StartDate.Sub(first_day).Hours()/24))/7 + 2
		html_event.Column = (positive(events[i].StartDate.Sub(first_day).Hours()/24))%7 + 1
		if events[i].StartDate.After(first_day) {
			html_event.Span = int(events[i].EndDate.Sub(events[i].StartDate).Hours()/24) + 1
		} else {
			html_event.Span = int(events[i].EndDate.Sub(first_day).Hours()/24) + 1
			html_event.IsContinuation = true
		}
		html_events = append(html_events, html_event)

		// place copies over multiple rows
		row := html_event.Row
		column := html_event.Column
		span := html_event.Span
		for column+span > 7 && row < max_row {
			span = span - (7 - column)
			column = 1
			row = row + 1
			copy_event := html_event
			copy_event.Span = span
			copy_event.Column = column
			copy_event.Row = row
			copy_event.Id = counter
			copy_event.IsContinuation = true
			counter = counter + 1
			html_events = append(html_events, copy_event)
		}
	}

	content := &CalendarContent{Month: MONTHS[monthNum-1], Year: yearNum, NextMonth: nextMonth, NextYear: nextYear,
		PrevMonth: prevMonth, PrevYear: prevYear, Days: days, Events: html_events}
	t.Execute(w, content)
}

func (handler WebserviceHandler) IndexHTML(w http.ResponseWriter, r *http.Request) {
	baseLang := handler.getBaseLanguage(r.FormValue("lang"))
	ps := context.Get(r, "params").(httprouter.Params)
	key := ps.ByName("langkey")
	term := ps.ByName("term")
	htmlHelpers := handler.getHelpers(baseLang)
	t := template.Must(template.New("index.html").Funcs(htmlHelpers).ParseFiles(handler.config.GetHTTPDir() + "index.html"))
	langs := handler.repo.GetLanguages(baseLang)
	content := &HtmlContent{Languages: langs, BaseLangTag: baseLang[:3]}
	if key != "" && term != "" {
		fromLang, toLang := handler.getLanguagesFromRequest(ps.ByName("langkey"))
		results, err := handler.repo.Search(term, fromLang, toLang, baseLang)
		if err != nil {
			//trying the other way around
			results, err = handler.repo.Search(term, toLang, fromLang, baseLang)
			if err != nil {
				panic(fmt.Sprintf("Word %s does not exist in %s/%s dictionary", term, fromLang, toLang))
				//http.Redirect(w, r, "/", http.StatusFound)
			} else {
				url := "/search/" + toLang[:3] + fromLang[:3] + "/" + term
				if baseLang != "" {
					url += "?lang=" + baseLang[:3]
				}
				http.Redirect(w, r, url, http.StatusFound)
				return
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
	handler.executeBasicTemplate(w, r, "about.html")
}

func (handler WebserviceHandler) TermsHTML(w http.ResponseWriter, r *http.Request) {
	handler.executeBasicTemplate(w, r, "terms.html")
}

func (handler WebserviceHandler) executeBasicTemplate(w http.ResponseWriter, r *http.Request, name string) {
	baseLang := handler.getBaseLanguage(r.FormValue("lang"))
	htmlHelpers := handler.getHelpers(baseLang)
	t := template.Must(template.New(name).Funcs(htmlHelpers).ParseFiles(handler.config.GetHTTPDir() + name))
	langs := handler.repo.GetLanguages(baseLang)
	content := &HtmlContent{Languages: langs, BaseLangTag: baseLang[:3]}
	t.Execute(w, content)
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
	numResults := len(result)
	// limit autocomplete results to 10
	if numResults > 10 {
		numResults = 10
	}
	enc.Encode(result[:numResults])
}

func (handler WebserviceHandler) Notify(w http.ResponseWriter, r *http.Request) {
	fromLang, toLang := handler.getLanguagesFromRequest(r.FormValue("langkey"))
	term := r.FormValue("word")
	if term == "" {
		panic("No word inserted")
	}
	message := "Word: " + term + "\nDictionary: " + fromLang + "-" + toLang
	go func() {
		err := handler.eutils.SendEmailToAdmins("Suggestion received", message)
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
		lang := handler.repo.GetLangFromKey(key)
		if lang != "" {
			baseLang = lang
		}
	}
	return baseLang
}
