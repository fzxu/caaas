package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"net/http"
	"time"
)

func main() {
	cluster := gocql.NewCluster(Config.Db.Http.Host)
	cluster.ProtoVersion = 4
	cluster.NumConns = Config.Db.NumConns
	cluster.Timeout = time.Duration(2) * time.Second
	cluster.Keyspace = Config.Db.DBName

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", Config.Http.Host, Config.Http.Port),
		Handler:        &ImgHandler{cluster},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
