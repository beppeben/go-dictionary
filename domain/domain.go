package domain

import (
	"strings"
)

type Word struct {
	Word         string
	LangKey      string
	Field        string
	FieldDesc    string
	Description  string
	Definition   string
	Locality     string
	Synonyms     []*Word
	Correlated   []*Word
	Translations []*Word
}

type SimpleWord struct {
	Word    string `json:"w"`
	LangTag string `json:"t"`
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
		prefix_i := strings.HasPrefix(strings.ToLower(a.Words[i].Word), a.Term)
		prefix_j := strings.HasPrefix(strings.ToLower(a.Words[j].Word), a.Term)
		if prefix_i && !prefix_j {
			return true
		} else if prefix_j && !prefix_i {
			return false
		}
	}
	return shortestAlphabetical(a.Words[i].Word, a.Words[j].Word)
}

func shortestAlphabetical(first, second string) bool {
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
	return shortestAlphabetical(a.Words[i].Word, a.Words[j].Word)
}
