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
	langMap   map[string]string
	allWords  map[string]*SimpleWordsPair
}

type ImportOptions struct {
	CheckHeaders bool
	AutoId       bool
	Square       bool
}

func NewRepo(h DbHandler, r *excel.ExcelReader) *SqlRepo {
	repo := &SqlRepo{handler: h, reader: r}
	repo.refreshLanguages()
	repo.refreshLanguageMap()
	repo.refreshWordsCache()
	return repo
}

func (r *SqlRepo) GetLanguages() []string {
	result := make([]string, len(r.languages))
	for i, _ := range r.languages {
		result[i] = strings.Title(r.languages[i])
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
	if title == "parent" || title == "synonyms" || title == "field" {
		return "INT"
	}
	if title == "english_id" {
		return "INT NOT NULL"
	}
	if title == "description" || title == "definition" {
		return "VARCHAR(5000)"
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

func (r *SqlRepo) refreshLanguageMap() {
	log.Info("Refreshing language map")
	r.langMap = make(map[string]string)
	for _, lang := range r.languages {
		r.langMap[lang[:3]] = lang
	}
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

	_, err = tx.Exec(st_insert, vals...)
	checkError(err, title)
}

func (r *SqlRepo) checkLanguageHeaders(title string, headers []string) error {
	if len(headers) != len(r.languages) {
		return fmt.Errorf("Table %v does not contain all the language columns: "+
			"expected %d, got %d", title, len(r.languages), len(headers))
	}
	headers_sorted := make([]string, len(headers))
	languages_sorted := make([]string, len(r.languages))
	for i := 0; i < len(headers); i++ {
		headers_sorted[i] = strings.ToLower(headers[i])
		languages_sorted[i] = strings.ToLower(r.languages[i])
	}
	sort.Strings(headers_sorted)
	sort.Strings(languages_sorted)
	for i := 0; i < len(headers); i++ {
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
			tx.Exec("DROP TABLE fields_expl")
			tx.Exec("DROP TABLE languages")
			for _, lang := range r.languages {
				if lang == "english" {
					continue
				}
				tx.Exec("DROP TABLE " + lang)
			}
			tx.Exec("DROP TABLE english")
			tx.Exec("DROP TABLE fields")
		}
		var err error
		if err = r.reader.RefreshFile(); err != nil {
			panic(err.Error())
		}
		r.createTable(tx, "languages", &ImportOptions{Square: true})
		r.refreshLanguagesFromTx(tx)
		r.createTable(tx, "fields", &ImportOptions{CheckHeaders: true})
		r.createTable(tx, "fields_expl", &ImportOptions{CheckHeaders: true})

		tx.Exec("ALTER TABLE fields_expl ADD FOREIGN KEY(id) REFERENCES fields(id)")

		//english is the master table, with synonyms and parent ids
		//in the other language tables every word has an english equivalent
		r.createTable(tx, "english", &ImportOptions{})

		tx.Exec("ALTER TABLE english ADD FOREIGN KEY(synonyms) REFERENCES english(id)")
		tx.Exec("ALTER TABLE english ADD FOREIGN KEY(parent) REFERENCES english(id)")
		tx.Exec("ALTER TABLE english ADD FOREIGN KEY(field) REFERENCES fields(id)")
		tx.Exec("ALTER TABLE fields_expl ADD FOREIGN KEY(id) REFERENCES fields(id)")

		tx.Exec("CREATE INDEX idx ON english(synonyms)")
		tx.Exec("CREATE INDEX wrd_eng ON english(word)")

		for _, lang := range r.languages {
			if lang == "english" {
				continue
			}
			r.createTable(tx, lang, &ImportOptions{AutoId: true})

			tx.Exec("ALTER TABLE " + lang +
				" ADD FOREIGN KEY(english_id) REFERENCES english(id)")

			tx.Exec("CREATE INDEX wrd_" + lang[:3] + " ON " + lang + "(word)")
		}
		return err
	})
	if err == nil {
		r.refreshLanguages()
		r.refreshLanguageMap()
		r.refreshWordsCache()
	}
	return err
}
