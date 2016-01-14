package main

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gocql/gocql"
)

func init() {
	// only support 3 type of images
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
}

type ImgHandler struct {
	cluster *gocql.ClusterConfig
}

func (h *ImgHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	session, err := h.cluster.CreateSession()
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	switch req.Method {
	case "GET":
		h.get(w, req, session)
	case "POST":
		h.post(w, req, session)
	}
}

func (h *ImgHandler) get(w http.ResponseWriter, req *http.Request, session *gocql.Session) {
	log.Println(req.URL.Path)
	id := strings.TrimPrefix(req.URL.Path, "/")
	asset, err := new(Asset).find(session, id)
	if err != nil {
		log.Panic(err)
	}
	w.Write(asset.Binary)
}

func (h *ImgHandler) post(w http.ResponseWriter, req *http.Request, session *gocql.Session) {
	req.ParseMultipartForm(10 << 20) // 10M
	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(fileHeader.Filename)

	img, format, err := image.Decode(file)
	if err != nil {
		log.Panic(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	m := imaging.Fit(img, Config.Image.StoreWidth, Config.Image.StoreHeight, imaging.Lanczos)

	out, err := os.Create("test_resized.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	log.Println(format)
	// write new image to file
	jpeg.Encode(out, m, &jpeg.Options{98})
}
