package persistence

import (
	"database/sql"
	"fmt"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
)

type MySqlConfig interface {
	GetPassDB() string
	GetDBName() string
}

type MySqlHandler struct {
	Connection *sql.DB
}

func NewMySqlHandler(c MySqlConfig) *MySqlHandler {
	info := fmt.Sprintf("password=%s dbname=%s sslmode=disable", c.GetPassDB(), c.GetDBName())
	conn, err := sql.Open("postgres", info)
	if err != nil {
		panic(err.Error())
	}
	err = conn.Ping()
	if err != nil {
		panic(err.Error())
	} else {
		log.Infoln("Established connection with database")
	}
	return &MySqlHandler{Connection: conn}
}

func (handler *MySqlHandler) Conn() *sql.DB {
	return handler.Connection
}

func (handler *MySqlHandler) Transact(txFunc func(*sql.Tx) (interface{}, error)) (obj interface{}, err error) {
	tx, err := handler.Connection.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			log.Infoln("Rolling back")
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	return txFunc(tx)
}

func (handler *MySqlHandler) TransactNoRet(txFunc func(*sql.Tx) error) error {
	f := func(tx *sql.Tx) (interface{}, error) {
		return nil, txFunc(tx)
	}
	_, err := handler.Transact(f)
	return err
}
