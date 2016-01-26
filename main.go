package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"

	"github.com/gocql/gocql"
)

func main() {
	flag.Parse()
	cluster := gocql.NewCluster(Config.Db.Hosts...)
	cluster.ProtoVersion = 4
	cluster.NumConns = Config.Db.NumConns
	cluster.Timeout = time.Duration(Config.Db.Timeout) * time.Second
	cluster.Keyspace = Config.Db.DBName

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", Config.Http.Host, Config.Http.Port),
		Handler:        &ImgHandler{cluster},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	glog.Info("Starting server on:", Config.Http.Host, Config.Http.Port)
	glog.Fatal(s.ListenAndServe())
}
