package domain

import (
	"strings"
	"time"
)

type CalendarEvent struct {
	Id          int64
	StartDate   time.Time
	EndDate     time.Time
	Tag         string
	Title       string
	Description string
}

type Word struct {
	//LangKey      string
	Word         string
	Lang         *Language
	Field        string
	FieldDesc    string
	Genre        string
	Description  string
	Definition   string
	Locality     string
	Synonyms     []*Word
	Correlated   []*Word
	Translations []*Word
}

type Language struct {
	Language string
	Tag      string
}

type SimpleWord struct {
	Word        string `json:"w"`
	WordASCII   string `json:"-"`
	NumSubWords int    `json:"-"`
	LangTag     string `json:"t"`
}

type SimpleWordsPair struct {
	First  []*SimpleWord
	Second []*SimpleWord
}

type LeastWordsAlphabeticSimple struct {
	Words     []*SimpleWord
	Term      string
	LangFirst string
}

func (a LeastWordsAlphabeticSimple) Len() int {
	return len(a.Words)
}

func (a LeastWordsAlphabeticSimple) Swap(i, j int) {
	a.Words[i], a.Words[j] = a.Words[j], a.Words[i]
}

func (a LeastWordsAlphabeticSimple) Less(i, j int) bool {
	if a.LangFirst != "" {
		lang_i := strings.HasPrefix(a.LangFirst, a.Words[i].LangTag)
		lang_j := strings.HasPrefix(a.LangFirst, a.Words[j].LangTag)
		if lang_i && !lang_j {
			return true
		} else if lang_j && !lang_i {
			return false
		}
	}
	if a.Term != "" {
		prefix_i := strings.HasPrefix(a.Words[i].WordASCII, a.Term)
		prefix_j := strings.HasPrefix(a.Words[j].WordASCII, a.Term)
		if prefix_i && !prefix_j {
			return true
		} else if prefix_j && !prefix_i {
			return false
		}
	}
	return shortestAlphabeticalWord(a.Words[i], a.Words[j])
}

func shortestAlphabeticalWord(first, second *SimpleWord) bool {
	if first.NumSubWords < second.NumSubWords {
		return true
	} else if first.NumSubWords > second.NumSubWords {
		return false
	}

	return first.Word < second.Word
}

func shortestAlphabeticalString(first, second string) bool {
	nwords_i := strings.Count(first, " ")
	nwords_j := strings.Count(second, " ")
	if nwords_i < nwords_j {
		return true
	} else if nwords_i > nwords_j {
		return false
	}

	return first < second
}

type LeastWordsAlphabetic struct {
	Words []*Word
}

func (a LeastWordsAlphabetic) Len() int {
	return len(a.Words)
}

func (a LeastWordsAlphabetic) Swap(i, j int) {
	a.Words[i], a.Words[j] = a.Words[j], a.Words[i]
}

func (a LeastWordsAlphabetic) Less(i, j int) bool {
	return shortestAlphabeticalString(a.Words[i].Word, a.Words[j].Word)
}
