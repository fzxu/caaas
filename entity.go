package main

import (
	"errors"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/glog"
)

type Asset struct {
	Id          gocql.UUID `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Path        []string   `json:"path,omitempty"`
	ContentType string     `json:"content_type,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	Binary      []byte     `json:"-"`
}

func (asset *Asset) Find(session *gocql.Session, assetId string) (*Asset, error) {
	var id gocql.UUID
	var name string
	var path string
	var contentType string
	var createdAt time.Time
	var binary []byte

	// Check if the assetId is an valid UUID
	idCheck, err := gocql.ParseUUID(assetId)
	if err != nil {
		return nil, err
	}

	if idCheck.Timestamp() == 0 {
		return nil, errors.New("Invalid UUID")
	}

	if err := session.Query(`SELECT id, name, path, contenttype, createdat, binary FROM assets WHERE id = ? LIMIT 1`,
		assetId).Consistency(gocql.One).Scan(&id, &name, &path, &contentType, &createdAt, &binary); err != nil {
		return nil, err
	}

	return &Asset{id, name, strings.Split(path, ","), contentType, createdAt, binary}, nil
}

func (asset Asset) FindByPath(session *gocql.Session, path string) ([]Asset, error) {
	var id gocql.UUID
	var name string
	var assets = make([]Asset, 0)
	iter := session.Query(`SELECT id, name FROM assetbypaths WHERE path = ?`, path).Iter()
	for iter.Scan(&id, &name) {
		assets = append(assets, Asset{Id: id, Name: name, Path: strings.Split(path, ",")})
	}
	return assets, nil
}

func (asset *Asset) Save(session *gocql.Session) error {
	if asset.Id.Timestamp() == 0 {
		asset.Id = gocql.TimeUUID()
		if err := session.Query(`INSERT INTO assets (id, name, path, contenttype, createdat, binary) VALUES (?, ?, ?, ?, ?, ?)`,
			asset.Id, asset.Name, strings.Join(asset.Path, ","), asset.ContentType, asset.CreatedAt, asset.Binary).Exec(); err != nil {
			glog.Fatal(err)
			return err
		}

		if err := session.Query(`INSERT INTO assetbypaths (path, id, name) VALUES (?, ?, ?)`,
			strings.Join(asset.Path, ","), asset.Id, asset.Name).Exec(); err != nil {
			glog.Fatal(err)
			return err
		}
		return nil
	} else {
		if err := session.Query(`UPDATE assets SET name = ?, path = ?, contenttype = ? WHERE id = ?`,
			asset.Name, strings.Join(asset.Path, ","), asset.ContentType, asset.Id).Exec(); err != nil {
			glog.Fatal(err)
			return err
		}

		if err := session.Query(`UPDATE assetbypaths SET name = ?, path = ? WHERE id = ?`,
			asset.Name, strings.Join(asset.Path, ","), asset.Id).Exec(); err != nil {
			glog.Fatal(err)
			return err
		}
		return nil
	}
}

func (asset Asset) Delete(session *gocql.Session, id string) error {
	a, err := asset.Find(session, id)
	if err != nil {
		return err
	}

	var errRet error
	if err := session.Query(`DELETE FROM assets WHERE id = ?`, id).Exec(); err != nil {
		glog.Fatal(err)
		errRet = err
	}
	if err := session.Query(`DELETE FROM assetbypaths WHERE path = ? AND id = ?`,
		strings.Join(a.Path, ","), id).Exec(); err != nil {
		glog.Fatal(err)
		errRet = err
	}
	return errRet
}

func (asset Asset) DeleteByPath(session *gocql.Session, path string) error {
	assets, err := asset.FindByPath(session, path)
	if err != nil {
		return err
	}

	var errRet error
	for _, a := range assets {
		err = asset.Delete(session, a.Id.String())
		if err != nil {
			errRet = err
		}
	}
	return errRet
}
