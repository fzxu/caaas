package main

import (
	"github.com/gocql/gocql"
	"log"
	"time"
)

type Asset struct {
	Id          gocql.UUID
	Name        string
	Path        string
	ContentType string
	CreatedAt   time.Time
	Binary      []byte
}

func (a *Asset) find(session *gocql.Session, assetId string) (*Asset, error) {
	var id gocql.UUID
	var name string
	var path string
	var contentType string
	var createdAt time.Time
	var binary []byte

	if err := session.Query(`SELECT id, name, path, contenttype, createdat, binary FROM assets WHERE id = ? LIMIT 1`,
		assetId).Consistency(gocql.One).Scan(&id, &name, &path, &contentType, &createdAt, &binary); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Asset{id, name, path, contentType, createdAt, binary}, nil
}

func (a *Asset) Save(session *gocql.Session, asset Asset) (gocql.UUID, error) {
	if asset.Id.String() == "" {
		id := gocql.TimeUUID()
		if err := session.Query(`INSERT INTO assets (id, name, path, contenttype, createdat, binary) VALUES (?, ?, ?, ?, ?, ?)`,
			id, asset.Name, asset.Path, asset.ContentType, asset.CreatedAt, asset.Binary).Exec(); err != nil {
			log.Fatal(err)
			return id, err
		}
		return id, nil
	} else {
		return asset.Id, nil
	}
}
