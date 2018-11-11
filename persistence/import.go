package persistence

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/go-dictionary/domain"
	"github.com/beppeben/go-dictionary/excel"
)

type DbHandler interface {
	Conn() *sql.DB
	Transact(txFunc func(*sql.Tx) (interface{}, error)) (interface{}, error)
	TransactNoRet(txFunc func(*sql.Tx) error) error
}

type SqlRepo struct {
	handler   DbHandler
	reader    *excel.ExcelReader
	languages []string
	//from key to english language name
	langMap map[string]string
	//allWords["lang1" + "lang2"] contains all the words in that specific dictionary
	allWords map[string]*SimpleWordsPair
	//langMatrix["lang1" + "lang2"] contains the translation of lang1 into lang2
	langMatrix map[string]string
	//webStrings["lang" + "key"] contains web entry "key" in language "lang"
	webStrings map[string]string
}

type ImportOptions struct {
	CheckHeaders bool
	AutoId       bool
	Square       bool
}

func NewRepo(h DbHandler, r *excel.ExcelReader) *SqlRepo {
	repo := &SqlRepo{handler: h, reader: r}
	repo.refreshLanguages()
	repo.refreshLanguageMaps()
	repo.refreshWordsCache()
	return repo
}

func (r *SqlRepo) GetLanguages(base string) []*Language {
	result := make([]*Language, len(r.languages))
	for i, _ := range r.languages {
		lang := r.langMatrix[r.languages[i][:3]+base[:3]]
		newLang := &Language{Language: strings.Title(lang), Tag: r.languages[i][:3]}
		if r.languages[i] == base {
			result[i] = result[0]
			result[0] = newLang
		} else {
			result[i] = newLang
		}
	}
	return result
}

func getDbType(sample string, title string) string {
	if title == "id" {
		_, err := strconv.ParseInt(sample, 10, 64)
		if err == nil {
			return "INT"
		}
	}
	if title == "parent" || title == "synonyms" || title == "field" || title == "genre" {
		return "INT"
	}
	if title == "english_id" {
		return "INT NOT NULL"
	}
	if title == "description" || title == "definition" || title == "about_html" || title == "terms_html" {
		return "VARCHAR(10000)"
	}
	return "VARCHAR(255)"
}

type QueryObj interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func (r *SqlRepo) refreshWordsCache() {
	log.Info("Refreshing words cache")
	r.allWords = make(map[string]*SimpleWordsPair)
	for i := 1; i < len(r.languages); i++ {
		for j := 0; j < i; j++ {
			l1 := r.languages[i]
			l2 := r.languages[j]
			_, _, err := r.GetWords(l1, l2)
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

func (r *SqlRepo) GetWebTerm(lang, key string) string {
	return r.webStrings[lang[:3]+key]
}

func (r *SqlRepo) saveWebTerms(lang string) {
	rows, err := r.handler.Conn().Query("SELECT * FROM web WHERE lower(id)=$1", lang)
	if err != nil {
		return
		//panic(err.Error())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}
	terms := make([]string, len(columns))
	pointers := make([]interface{}, len(columns))
	for i, _ := range columns {
		pointers[i] = &terms[i]
	}
	if rows.Next() {
		err = rows.Scan(pointers...)
		if err != nil {
			panic(err.Error())
		}
		for i, _ := range columns {
			r.webStrings[lang[:3]+columns[i]] = terms[i]
		}
	}
}

func (r *SqlRepo) fillMissingWebTerms() {
	for key := range r.webStrings {
		if key[:3] == "eng" {
			continue
		}
		if r.webStrings[key] == "" {
			r.webStrings[key] = r.webStrings["eng"+key[3:]]
		}
	}
}

func (r *SqlRepo) refreshLanguageMaps() {
	log.Info("Refreshing language maps")
	r.langMap = make(map[string]string)
	r.langMatrix = make(map[string]string)
	r.webStrings = make(map[string]string)

	for _, lang := range r.languages {
		r.saveWebTerms(lang)
		r.langMap[lang[:3]] = lang
		for _, other := range r.languages {
			rows, err := r.handler.Conn().Query("SELECT "+lang+" FROM languages WHERE lower(id)=$1", other)
			if err != nil {
				panic(err.Error())
			}
			defer rows.Close()
			var tran string
			rows.Next()
			rows.Scan(&tran)
			r.langMatrix[lang[:3]+other[:3]] = strings.ToLower(tran)
		}
	}

	r.fillMissingWebTerms()
}

func (r *SqlRepo) GetLangFromKey(key string) string {
	return r.langMap[key]
}

func (r *SqlRepo) refreshLanguagesFromTx(tx QueryObj) error {
	rows, err := tx.Query("SELECT id FROM languages")
	if err != nil {
		return err
	}
	defer rows.Close()
	result := make([]string, 0)
	var lang string
	for rows.Next() {
		rows.Scan(&lang)
		result = append(result, strings.ToLower(lang))
	}
	r.languages = result
	return nil
}

func (r *SqlRepo) refreshLanguages() error {
	return r.refreshLanguagesFromTx(r.handler.Conn())
}

func checkError(err error, table string) {
	if err != nil {
		panic(fmt.Sprintf("Error in table %s: %v", table, err))
	}
}

func (r *SqlRepo) createTableFromMatrix(tx *sql.Tx, title string, matrix [][]string, autoId bool) {
	st_create := "CREATE TABLE " + title + "(id "
	st_insert := "INSERT INTO " + title + "("
	db_types := make([]string, len(matrix[0]))
	offset := 1
	if autoId {
		st_create += "SERIAL PRIMARY KEY,"
		offset = 0
	} else {
		db_types[0] = getDbType(matrix[1][0], "id")
		st_create += db_types[0] + " PRIMARY KEY"
		st_insert += "id"
	}
	for i := offset; i < len(matrix[0]); i++ {
		db_types[i] = getDbType(matrix[1][i], strings.ToLower(matrix[0][i]))
		if i > 0 {
			st_create += ","
			st_insert += ","
		}
		st_create += matrix[0][i] + " " + db_types[i]
		st_insert += matrix[0][i]
	}
	st_create += ")"
	log.Debug(st_create)
	_, err := tx.Exec(st_create)
	checkError(err, title)

	st_insert += ") VALUES "
	vals := make([]interface{}, 0)
	var value interface{}
	var buffer bytes.Buffer
	for i := 1; i < len(matrix); i++ {
		buffer.WriteString("(")
		for j := 0; j < len(matrix[0]); j++ {
			buffer.WriteString("$" + strconv.Itoa((i-1)*len(matrix[0])+j+1))
			if j < len(matrix[0])-1 {
				buffer.WriteString(",")
			}
			value = matrix[i][j]
			if value == "" && db_types[j] == "INT" {
				if db_types[j] == "INT" {
					value = sql.NullInt64{}
				} else {
					value = sql.NullString{}
				}
			}
			vals = append(vals, value)
		}
		buffer.WriteString("),")
	}
	st_insert += buffer.String()
	st_insert = st_insert[0 : len(st_insert)-1]
	log.Debug(st_insert)
	log.Debug(vals)
	_, err = tx.Exec(st_insert, vals...)
	checkError(err, title)
}

func (r *SqlRepo) checkLanguageHeaders(title string, headers []string) error {
	if len(headers) < len(r.languages) {
		return fmt.Errorf("Table %v does not contain all the language columns: "+
			"expected %d, got %d", title, len(r.languages), len(headers))
	}
	headers_sorted := make([]string, len(r.languages))
	languages_sorted := make([]string, len(r.languages))
	for i := 0; i < len(r.languages); i++ {
		headers_sorted[i] = strings.ToLower(headers[i])
		languages_sorted[i] = strings.ToLower(r.languages[i])
	}
	sort.Strings(headers_sorted)
	sort.Strings(languages_sorted)
	for i := 0; i < len(r.languages); i++ {
		if headers_sorted[i] != languages_sorted[i] {
			return fmt.Errorf("Some languages are missing from Table %v", title)
		}
	}
	return nil
}

func (r *SqlRepo) createTable(tx *sql.Tx, title string, opts *ImportOptions) {
	log.Infof("Creating %v table", title)
	matrix, err := r.reader.GetMatrix(title)
	checkError(err, title)
	if opts.Square {
		for i := 1; i < len(matrix[0]); i++ {
			if matrix[0][i] != matrix[i][0] {
				panic(fmt.Errorf("Row and column ids of %v matrix must coincide", title))
			}
		}
	}
	if opts.CheckHeaders {
		err = r.checkLanguageHeaders(title, matrix[0][1:])
		if err != nil {
			panic(err.Error())
		}
	}
	r.createTableFromMatrix(tx, title, matrix, opts.AutoId)
}

func (r *SqlRepo) ResetDB() error {
	err := r.reader.RefreshFile()
	if err != nil {
		return err
	}
	err = r.handler.TransactNoRet(func(tx *sql.Tx) error {
		log.Debugf("%d languages currently stored", len(r.languages))
		if len(r.languages) > 0 {
			log.Info("Removing all tables")
			tx.Exec("DROP TABLE IF EXISTS web")
			tx.Exec("DROP TABLE IF EXISTS fields_expl")
			tx.Exec("DROP TABLE IF EXISTS languages")
			for _, lang := range r.languages {
				if lang == "english" {
					continue
				}
				tx.Exec("DROP TABLE IF EXISTS " + lang)
			}
			tx.Exec("DROP TABLE IF EXISTS english")
			tx.Exec("DROP TABLE IF EXISTS fields")
			tx.Exec("DROP TABLE IF EXISTS genre")
		}
		var err error
		if err = r.reader.RefreshFile(); err != nil {
			panic(err.Error())
		}
		r.createTable(tx, "languages", &ImportOptions{Square: true})
		r.refreshLanguagesFromTx(tx)
		r.createTable(tx, "fields", &ImportOptions{CheckHeaders: true})
		r.createTable(tx, "fields_expl", &ImportOptions{CheckHeaders: true})
		_, err = tx.Exec("ALTER TABLE fields_expl ADD FOREIGN KEY(id) REFERENCES fields(id)")
		checkError(err, "fields_expl")
		r.createTable(tx, "genre", &ImportOptions{CheckHeaders: true})

		r.createTable(tx, "web", &ImportOptions{})
		_, err = tx.Exec("ALTER TABLE web ADD FOREIGN KEY(id) REFERENCES languages(id)")
		checkError(err, "web")

		//english is the master table, with synonyms and parent ids
		//in the other language tables every word has an english equivalent
		r.createTable(tx, "english", &ImportOptions{})

		_, err = tx.Exec("ALTER TABLE english ADD FOREIGN KEY(synonyms) REFERENCES english(id)")
		checkError(err, "english")
		_, err = tx.Exec("ALTER TABLE english ADD FOREIGN KEY(parent) REFERENCES english(id)")
		checkError(err, "english")
		_, err = tx.Exec("ALTER TABLE english ADD FOREIGN KEY(field) REFERENCES fields(id)")
		checkError(err, "english")
		_, err = tx.Exec("ALTER TABLE english ADD FOREIGN KEY(genre) REFERENCES genre(id)")
		checkError(err, "english")

		tx.Exec("CREATE INDEX idx ON english(synonyms)")
		tx.Exec("CREATE INDEX wrd_eng ON english(word)")

		for _, lang := range r.languages {
			if lang == "english" {
				continue
			}
			r.createTable(tx, lang, &ImportOptions{AutoId: true})

			_, err = tx.Exec("ALTER TABLE " + lang +
				" ADD FOREIGN KEY(english_id) REFERENCES english(id)")
			checkError(err, lang)

			_, err = tx.Exec("ALTER TABLE " + lang + " ADD FOREIGN KEY(genre) REFERENCES genre(id)")
			checkError(err, lang)

			tx.Exec("CREATE INDEX wrd_" + lang[:3] + " ON " + lang + "(word)")
		}
		return err
	})
	if err == nil {
		r.refreshLanguages()
		r.refreshLanguageMaps()
		r.refreshWordsCache()
	}
	return err
}
