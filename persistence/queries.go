package persistence

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
	. "github.com/beppeben/go-dictionary/utils"
)

const (
	translationsAndForeignSynonymsFromAnyBase = "WITH RECURSIVE toeng(:lang1, english, syn, enid) AS" +
		"(SELECT :lang1.word, english.word, english.synonyms, english.id  FROM :lang1 " +
		"INNER JOIN english ON :lang1.english_id=english.id " +
		"UNION " +
		"SELECT CAST('' as VARCHAR(255)), english.word, english.synonyms, english.id " +
		"FROM english JOIN toeng ON toeng.syn=english.id OR toeng.enid=english.synonyms) "

	translationsAndForeignSynonymsFromAnyToAny = translationsAndForeignSynonymsFromAnyBase +
		"SELECT :lang1, :lang2.word AS :lang2 FROM toeng " +
		"INNER JOIN :lang2 ON toeng.enid=:lang2.english_id;"

	translationsAndForeignSynonymsFromAnyToEng = translationsAndForeignSynonymsFromAnyBase +
		"SELECT :lang1, english FROM toeng;"

	translationsAndForeignSynonymsFromEng = "SELECT english.word, :lang.word FROM english " +
		"INNER JOIN :lang ON english.id=:lang.english_id;"

	allWords = "SELECT word, word from :lang;"

	searchWithSynonymsBase = "WITH RECURSIVE toeng(syn, enid) AS" +
		"(SELECT synonyms, id FROM english WHERE id=$1 " +
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
	} else if lang1 == "english" {
		s = strings.Replace(translationsAndForeignSynonymsFromEng, ":lang", lang2, -1)
		//s = strings.Replace(translationsAndForeignSynonymsFromAnyToEng, ":lang1", lang, -1)
	} else if lang2 == "english" {
		s = strings.Replace(translationsAndForeignSynonymsFromAnyToEng, ":lang1", lang1, -1)
	} else {
		s = translationsAndForeignSynonymsFromAnyToAny
		s = strings.Replace(s, ":lang1", lang1, -1)
		s = strings.Replace(s, ":lang2", lang2, -1)
	}
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
	term = MapToASCII(term)
	words1, words2, err := r.GetWords(lang1, lang2)
	if err != nil {
		return nil, err
	}
	words = make([]*SimpleWord, 0)
	for _, w := range words1 {
		if strings.Contains(MapToASCII(w.Word), term) {
			words = append(words, w)
		}
	}
	for _, w := range words2 {
		if strings.Contains(MapToASCII(w.Word), term) {
			words = append(words, w)
		}
	}
	sort.Sort(LeastWordsAlphabeticSimple{Words: words, Term: term, LangFirst: lang1})
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
	r.queryAndAddToSets(translationsAndForeignSynonymsStmt(lang1, lang2), set1, set2)
	r.queryAndAddToSets(translationsAndForeignSynonymsStmt(lang2, lang1), set2, set1)
	for word, _ := range set1 {
		words1 = append(words1, &SimpleWord{word, lang1[:3]})
	}
	sort.Sort(LeastWordsAlphabeticSimple{Words: words1})
	for word, _ := range set2 {
		words2 = append(words2, &SimpleWord{word, lang2[:3]})
	}
	sort.Sort(LeastWordsAlphabeticSimple{Words: words2})
	r.allWords[lang1+lang2] = &SimpleWordsPair{First: words1, Second: words2}
	r.allWords[lang2+lang1] = &SimpleWordsPair{First: words2, Second: words1}
	return
}

func (r *SqlRepo) Search(word, fromLang, toLang, baseLang string) (words []*Word, err error) {
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
		lang := &Language{Language: r.langMatrix[fromLang[:3]+baseLang[:3]], Tag: fromLang[:3]}
		w := &Word{Word: word, Description: description, Definition: definition,
			Locality: loc, Lang: lang}
		translations, err := r.translate(w, toLang, baseLang, enId)
		if err != nil {
			log.Infoln(err.Error())
			return nil, err
		}
		w.Translations = translations
		synonyms, err := r.translate(w, fromLang, baseLang, enId)
		if err != nil {
			log.Infoln(err.Error())
			return nil, err
		}
		w.Synonyms = synonyms
		words = append(words, w)
		statement = "SELECT fields." + baseLang + ", fields_expl." + baseLang + " FROM english " +
			"INNER JOIN fields on english.field=fields.id " +
			"INNER JOIN fields_expl ON fields.id=fields_expl.id WHERE english.id=$1"
		log.Debug(statement)
		frows, err := r.handler.Conn().Query(statement, enId)
		if err != nil {
			log.Infoln(err.Error())
			return nil, err
		}
		defer frows.Close()
		var field, desc string
		for frows.Next() {
			frows.Scan(&field, &desc)
			w.Field = field
			w.FieldDesc = desc
		}
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("Word %s not found in %s table", word, fromLang)
	}
	sort.Sort(LeastWordsAlphabetic{Words: words})
	return words, nil
}

func (r *SqlRepo) translate(word *Word, toLang string, baseLang string, enId int64) (words []*Word, err error) {
	var statement string
	if toLang == "english" {
		statement = searchWithSynonymsToEng
	} else {
		statement = strings.Replace(searchWithSynonymsToAny, ":lang", toLang, -1)
	}
	rows, err := r.handler.Conn().Query(statement, enId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var wrd, description, definition, loc string
	for rows.Next() {
		rows.Scan(&wrd, &description, &definition, &loc)
		if word.Word != wrd || word.Lang.Tag != toLang[:3] {
			lang := &Language{Language: r.langMatrix[toLang[:3]+baseLang[:3]], Tag: toLang[:3]}
			w := &Word{Word: wrd, Description: description, Definition: definition,
				Locality: loc, Lang: lang}
			words = append(words, w)
		}
	}
	return words, nil
}
