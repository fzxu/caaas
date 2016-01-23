package main

import (
	//"log"
	"testing"

	"github.com/gocql/gocql"
)

func TestCreate(t *testing.T) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.ProtoVersion = 4
	cluster.Keyspace = "aaas"
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//asset, err := new(Asset).Find(session, "668a43b0-b82b-11e5-9ac5-f7654b5743d8")
	if err != nil {
		//log.Panic(err)
	}
	//log.Println(asset)
}
