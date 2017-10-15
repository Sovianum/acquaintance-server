package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"github.com/Sovianum/acquaintance-server/config"
	"github.com/Sovianum/acquaintance-server/server"
	"net/http"
	"github.com/gorilla/handlers"
	"fmt"
)

const (
	confFile = "resources/config.json"
)

func main() {
	var conf = getConf()
	var db, err = connectDB(conf)
	if err != nil {
		panic(err)
	}

	var env = server.NewEnv(db, conf)
	var router = server.GetRouter(env)
	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, router))
}

func connectDB(conf config.Conf) (*sql.DB, error) {
	var db *sql.DB
	var err error

	if db, err = sql.Open(conf.DB.DriverName, conf.DB.GetEnvAuthString()); err != nil {
		return nil, err
	}
	if err = db.Ping(); err == nil {
		fmt.Println("Authorized via env")
		return db, nil
	}
	if db, err = sql.Open(conf.DB.DriverName, conf.DB.GetAuthStr()); err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("Authorized via conf")
	return db, err
}

func getConf() config.Conf {
	var file, confErr = os.Open(confFile)
	if confErr != nil {
		panic(confErr)
	}
	defer file.Close()

	var conf, err = config.ReadConf(file)
	if err != nil {
		panic(err)
	}

	return conf
}
