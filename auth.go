package main

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func CheckBasicAuth(r *http.Request) bool {
	if r.Header["Authorization"] == nil {
		return false
	}

	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		return false
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 || !Validate(pair[0], pair[1]) {
		return false
	}
	return true
}

func Validate(username, password string) bool {
	if username == Config.Auth.Username && password == Config.Auth.Password {
		return true
	}
	return false
}
