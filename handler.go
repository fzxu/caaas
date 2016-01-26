package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"

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
	// if file was cached, return directly
	// put it before creating the db session for higher performance
	if req.Method == "GET" {
		if id := h.getUUID(req.URL.Path); id != "" {
			cachedFile, _ := filepath.Abs(filepath.Clean(Config.Image.CacheDir + req.URL.Path))
			cachedData, err := ioutil.ReadFile(cachedFile)
			if err == nil {
				w.Write(cachedData)
				return
			}
		}
	}

	session, err := h.cluster.CreateSession()
	if err != nil {
		glog.Fatal(err)
	}
	defer session.Close()

	switch req.Method {
	case "GET":
		h.get(w, req, session)
	case "POST":
		if !CheckBasicAuth(req) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Authentication Failed")
			return
		}
		h.post(w, req, session)
	case "DELETE":
		if !CheckBasicAuth(req) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Authentication Failed")
			return
		}
		h.delete(w, req, session)
	}
}

func (h *ImgHandler) get(w http.ResponseWriter, req *http.Request, session *gocql.Session) {
	path := strings.Trim(req.URL.Path, "/")
	if id := h.getUUID(path); id != "" {
		width, mode, height := h.getSizes(path)

		asset, err := new(Asset).Find(session, id)
		if err != nil {
			glog.Fatal(err)
		}

		var buf []byte
		buffer := bytes.NewBuffer(buf)
		err = h.processImage(bytes.NewBuffer(asset.Binary), buffer, mode, width, height)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err.Error())
			return
		}

		// create cache image and wrap a multiwriter
		cachedFilePath, _ := filepath.Abs(filepath.Clean(Config.Image.CacheDir + req.URL.Path))
		cacheFile, _ := os.Create(cachedFilePath)
		defer cacheFile.Close()

		multiWriter := io.MultiWriter(w, cacheFile)
		multiWriter.Write(buffer.Bytes())
	} else {
		// list image under a path
		pathComma := strings.Join(strings.Split(path, "/"), ",")
		assets, err := new(Asset).FindByPath(session, pathComma)
		if err != nil {
			glog.Fatal(err)
		}
		data, err := json.Marshal(assets)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func (h *ImgHandler) post(w http.ResponseWriter, req *http.Request, session *gocql.Session) {
	path := req.URL.Path
	if path == "" || path == "/" {
		glog.Error("Please specify the path")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "please specify the path")
		return
	}

	req.ParseMultipartForm(10 << 20) // 10M
	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		glog.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}

	glog.Info(fileHeader.Filename)

	var buf []byte
	buffer := bytes.NewBuffer(buf)
	err = h.processImage(file, buffer, "z", Config.Image.StoreWidth, Config.Image.StoreHeight)
	if err != nil {
		glog.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	_, fn := filepath.Split(fileHeader.Filename)
	fileName, _ := url.QueryUnescape(fn)

	glog.Info("resized:", fileName)
	asset := &Asset{
		Name:        fileName,
		Path:        strings.Split(strings.Trim(path, "/"), "/"),
		ContentType: "image/jpeg",
		CreatedAt:   time.Now(),
		Binary:      buffer.Bytes(),
	}

	err = asset.Save(session)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info("saved:", fileName)
	data, err := json.Marshal(asset)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *ImgHandler) delete(w http.ResponseWriter, req *http.Request, session *gocql.Session) {
	path := strings.Trim(req.URL.Path, "/")
	var err error
	if id := h.getUUID(path); id != "" {
		err = new(Asset).Delete(session, id)
	} else {
		pathComma := strings.Join(strings.Split(path, "/"), ",")
		err = new(Asset).DeleteByPath(session, pathComma)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
	}
}

func (h *ImgHandler) getUUID(text string) string {
	r := regexp.MustCompile("[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}")
	if r.MatchString(text) {
		return r.FindStringSubmatch(text)[0]
	}
	return ""
}

// get the width, crop mode and height
func (h *ImgHandler) getSizes(path string) (int, string, int) {
	segments := strings.Split(path, "__")

	width := Config.Image.DefaultWidth
	height := Config.Image.DefaultHeight
	mode := "z"

	if len(segments) > 1 {
		reg := regexp.MustCompile(`([0-9]+)([z|x])([0-9]+)`)
		parts := reg.FindStringSubmatch(segments[1])
		widthInt, _ := strconv.Atoi(parts[1])
		mode = parts[2]
		heightInt, _ := strconv.Atoi(parts[3])

		if widthInt != 0 || heightInt != 0 {
			width = widthInt
			height = heightInt

			if widthInt == 0 {
				height = heightInt
				width = height
			}

			if heightInt == 0 {
				width = widthInt
				height = width
			}
		}
	}
	return width, mode, height
}

func (h *ImgHandler) processImage(in io.Reader, out io.Writer, mode string, width int, height int) error {
	if Config.Image.UseGoRoutine {
		ImageChannel <- 1
		defer func() {
			<-ImageChannel
		}()
	}

	img, _, err := image.Decode(in)
	if err != nil {
		glog.Fatal(err)
		return err
	}

	var m *image.NRGBA
	switch mode {
	case "z":
		m = imaging.Fit(img, width, height, imaging.Lanczos)
	case "x":
		m = imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)
	}

	jpeg.Encode(out, m, &jpeg.Options{Config.Image.ReadQuality})
	return nil
}
