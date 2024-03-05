package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	// "os"
	"strings"
	"testing"
)


var mux *http.ServeMux
var writer *httptest.ResponseRecorder

func setUp() {
	mux = http.NewServeMux()
	writer = httptest.NewRecorder()
}

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	os.Exit(code)
}

func TestRegister(t *testing.T) {
	// t.Skip("Skip the register test")
	// mux := http.NewServeMux()
	mux.HandleFunc("/register", signUp)

	uploadData := strings.NewReader(`{
		"email": "Derek@hey.com",
		"fullname": "Derek Ficher",
		"password": "password"
	}`)

	// writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/register", uploadData)

	if err != nil {
		t.Fatal("Cannot create request for endpoint")
	}

	mux.ServeHTTP(writer, request)
	if writer.Code != 200 {
		t.Error("Could not create user")
	}

	var username map[string]string
	json.Unmarshal(writer.Body.Bytes(), &username)
	t.Log(username["username"])
}

func TestUpload(t *testing.T) {
	// mux := http.NewServeMux()
	mux.HandleFunc("/upload", Upload)

	uploadData := strings.NewReader(`{
		"from": "Sarki Ihima101",
		"to": "Sharon Resser101",
		"drive": 5,
		"filename": "proposal.txt",
		"is_dir": false,
		"Note": "this is a new proposal for you"
	}`)

	// writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/upload", uploadData)

	if err != nil {
		t.Fatal("Cannot create request for endpoint")
	}

	mux.ServeHTTP(writer, request)
	if writer.Code != 200 {
		t.Error("Could not get url")
	}
	var url map[string]string
	err = json.Unmarshal(writer.Body.Bytes(), &url)
	if err != nil {
		t.Error(err)
	}

	t.Log(url["presign_url"])
}