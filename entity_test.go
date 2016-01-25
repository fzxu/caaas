package main

import (
	"log"
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

	asset := &Asset{
		Name:        "test1",
		Path:        []string{"a", "b"},
		ContentType: "image/jpeg",
	}
	err = asset.Save(session)
	if err != nil {
		log.Panic(err)
	}

	err = asset.Delete(session, asset.Id.String())
	if err != nil {
		log.Panic(err)
	}
	log.Println(asset)
}
