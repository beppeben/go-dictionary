package persistence

import (
	"fmt"
	"sort"
	"strings"

	//log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
)

const (
	translationsAndForeignSynonymsFromAny = "WITH RECURSIVE toeng(:lang1, english, syn, enid) AS" +
		"(SELECT :lang1.word, english.word, english.synonyms, english.id  FROM :lang1 " +
		"INNER JOIN english ON :lang1.english_id=english.id " +
		"UNION " +
		"SELECT CAST('' as VARCHAR(255)), english.word, english.synonyms, english.id " +
		"FROM english JOIN toeng ON toeng.syn=english.id OR toeng.enid=english.synonyms) " +
		"SELECT :lang1, :lang2.word AS :lang2 FROM toeng " +
		"INNER JOIN :lang2 ON toeng.enid=:lang2.english_id;"

	translationsAndForeignSynonymsFromEng = "SELECT english.word, :lang.word FROM english " +
		"INNER JOIN :lang ON english.id=:lang.english_id;"

	allWords = "SELECT word, word from :lang;"

	searchWithSynonymsBase = "WITH RECURSIVE toeng(syn, enid) AS" +
		"(SELECT synonyms, id  FROM english WHERE id=$1 " +
		"UNION " +
		"SELECT synonyms, id " +
		"FROM english JOIN toeng ON toeng.syn=english.id OR toeng.enid=english.synonyms) "

	searchWithSynonymsToAny = searchWithSynonymsBase +
		"SELECT word, description, definition, loc FROM toeng " +
		"INNER JOIN :lang ON toeng.enid=:lang.english_id;"

	searchWithSynonymsToEng = searchWithSynonymsBase +
		"SELECT word, description, definition, loc FROM toeng " +
		"INNER JOIN english ON toeng.enid=english.id;"
)

func translationsAndForeignSynonymsStmt(lang1 string, lang2 string) string {
	var s string
	if lang1 == lang2 {
		s = strings.Replace(allWords, ":lang", lang1, -1)
	} else if lang1 == "english" || lang2 == "english" {
		lang := lang1
		if lang1 == "english" {
			lang = lang2
		}
		s = strings.Replace(translationsAndForeignSynonymsFromEng, ":lang", lang, -1)
	} else {
		s = translationsAndForeignSynonymsFromAny
		s = strings.Replace(s, ":lang1", lang1, -1)
		s = strings.Replace(s, ":lang2", lang2, -1)
	}
	//log.Info(s)
	return s
}

func (r *SqlRepo) queryAndAddToSets(statement string, set1 map[string]bool, set2 map[string]bool) {
	rows, err := r.handler.Conn().Query(statement)
	if err != nil {
		return
	}
	defer rows.Close()
	var l1, l2 string
	for rows.Next() {
		rows.Scan(&l1, &l2)
		if l1 != "" {
			set1[l1] = true
		}
		if l2 != "" {
			set2[l2] = true
		}
	}
}

func (r *SqlRepo) GetWordsWithTerm(term string, lang1 string, lang2 string) (words []*SimpleWord, err error) {
	words1, words2, err := r.GetWords(lang1, lang2)
	if err != nil {
		return nil, err
	}
	words = make([]*SimpleWord, 0)
	for _, w := range words1 {
		if strings.Contains(strings.ToLower(w.Word), strings.ToLower(term)) {
			words = append(words, w)
		}
	}
	for _, w := range words2 {
		if strings.Contains(strings.ToLower(w.Word), strings.ToLower(term)) {
			words = append(words, w)
		}
	}
	return
}

func (r *SqlRepo) GetWords(lang1 string, lang2 string) (words1 []*SimpleWord, words2 []*SimpleWord, err error) {
	words := r.allWords[lang1+lang2]
	if words != nil {
		words1 = words.First
		words2 = words.Second
		return
	}
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	var swap bool
	if lang1 == "english" || lang2 == "english" {
		if lang2 == "english" {
			lang2 = lang1
			lang1 = "english"
			swap = true
		}
		r.queryAndAddToSets(translationsAndForeignSynonymsStmt(lang1, lang2), set1, set2)
		if swap {
			temp := set2
			set2 = set1
			set1 = temp
			lang1 = lang2
			lang2 = "english"
		}
	} else {
		r.queryAndAddToSets(translationsAndForeignSynonymsStmt(lang1, lang2), set1, set2)
		r.queryAndAddToSets(translationsAndForeignSynonymsStmt(lang2, lang1), set2, set1)
	}
	for word, _ := range set1 {
		words1 = append(words1, &SimpleWord{word, lang1[:3]})
	}
	sort.Sort(Alphabetic(words1))
	for word, _ := range set2 {
		words2 = append(words2, &SimpleWord{word, lang2[:3]})
	}
	sort.Sort(Alphabetic(words2))
	r.allWords[lang1+lang2] = &SimpleWordsPair{First: words1, Second: words2}
	r.allWords[lang2+lang1] = &SimpleWordsPair{First: words2, Second: words1}
	return
}

func (r *SqlRepo) Search(word string, fromLang string, toLang string) (words []*Word, err error) {
	var statement string
	if fromLang == "english" {
		statement = "SELECT id, description, definition, loc FROM english WHERE WORD=$1"
	} else {
		statement = "SELECT english_id, description, definition, loc FROM :lang WHERE WORD=$1"
		statement = strings.Replace(statement, ":lang", fromLang, -1)
	}
	rows, err := r.handler.Conn().Query(statement, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	words = make([]*Word, 0)
	var enId int64
	var description, definition, loc string
	for rows.Next() {
		rows.Scan(&enId, &description, &definition, &loc)
		w := &Word{Word: word, Description: description, Definition: definition, Locality: loc, LangKey: fromLang}
		r.translate(w, toLang, enId)
		words = append(words, w)
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("Word %s not found in %s table", word, fromLang)
	}
	return words, nil
}

func (r *SqlRepo) translate(word *Word, toLang string, enId int64) error {
	var statement string
	if toLang == "english" {
		statement = searchWithSynonymsToEng
	} else {
		statement = strings.Replace(searchWithSynonymsToAny, ":lang", toLang, -1)
	}
	rows, err := r.handler.Conn().Query(statement, enId)
	if err != nil {
		return err
	}
	defer rows.Close()
	words := make([]*Word, 0)
	var wrd, description, definition, loc string
	for rows.Next() {
		rows.Scan(&wrd, &description, &definition, &loc)
		w := &Word{Word: wrd, Description: description, Definition: definition, Locality: loc, LangKey: toLang}
		words = append(words, w)
	}
	word.Translations = words
	return nil
}
