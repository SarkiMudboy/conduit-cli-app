package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var conf aws.Config
var port = ":8000"

func init() {
	conf = getConfig("us-east-1")
}

func getUploadURL(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("%s - %s\n", r.Method, r.URL)
	filename := make(map[string]string)
	data := make([]byte, r.ContentLength)

	_, err := r.Body.Read(data)

	if err != nil && err != io.EOF{
		log.Println("Cannot read request data: ", err)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(data, &filename)
	if err != nil {
		log.Println("Cannot extract json data: ", err)
		return
	}
	presign_url := putPresignURL(conf, "conduitcli-test-bucket", filename["filename"])
	url := []byte(fmt.Sprintf(`{"presign_url": "%s"}`, presign_url))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(url)

}

func getDownloadURL(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("%s - %s (%d)\n", r.Method, r.URL, r.ContentLength)
	filename := make(map[string]string)

	data := make([]byte, r.ContentLength)
	_, err := r.Body.Read(data)

	if err != nil && err != io.EOF{
		log.Println("Cannot read request data: ", err)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(data, &filename)
	if err != nil {
		log.Println("Cannot extract json data: ", err)
		return
	}
	presign_url := getPresignURL(conf, "conduitcli-test-bucket", filename["filename"])
	url := []byte(fmt.Sprintf(`{"presign_url": "%s"}`, presign_url))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(url)

}

func main() {

	mux := http.NewServeMux()
	server := http.Server{
		Addr: port,
		Handler: mux,
	}
	mux.HandleFunc("/get-upload-url", getUploadURL)
	mux.HandleFunc("/get-download-url", getDownloadURL)
	err := server.ListenAndServe()

	if err != nil {
		fmt.Printf("Server error: %s\n", err)
	} else {
		fmt.Println("Server started: Listening on port ", port)
	}
}