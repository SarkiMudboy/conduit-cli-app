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

type UploadData struct {
	From string `json:"from"`
	To string `json:"to"`
	Drive int `json:"drive"`
	FileName string `json:"filename"`
	IsDir bool `json:"is_dir"`
	Note string `json:"note"`
}



func init() {
	conf = getConfig("us-east-1")
}

func Upload(w http.ResponseWriter, r *http.Request) {

	// data in -> filename, from_username (user), to_username, note, drive, is_dir 
	
	fmt.Printf("%s - %s\n", r.Method, r.URL)
	uploadData := new(UploadData)
	data := make([]byte, r.ContentLength)

	_, err := r.Body.Read(data)

	if err != nil && err != io.EOF{
		log.Println("Cannot read request data: ", err)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(data, uploadData)
	if err != nil {
		log.Println("Cannot extract json data: ", err)
		return
	}
	// check cache here...

	presign_url, expires := putPresignURL(conf, "conduitcli-test-bucket", uploadData.FileName)
	url := []byte(fmt.Sprintf(`{"presign_url": "%s"}`, presign_url))

	// save to cache...
	err = SaveFileToCache(*uploadData)
	// goroutine that waits for the time to elapse and queries AWS for file
	if err!= nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(url)
}


func SaveFileToCache(data UploadData) error {
	// save to cache
	// get users
	from := User{
		UserName: data.From,
	}
	to := User{
		UserName: data.To,
	}

	from.GetObjectByField("username")
	to.GetObjectByField("username")

	drive := Drive{Id: data.Drive}
	_ = drive.Manager("retrieve")

	object := Object{
		Name: data.FileName,
		IsDir: data.IsDir,
		Drive: &drive,
	}

	share := Share{
		From: &from,
		To: &to,
		File: &object,
		Note: data.Note,
		Drive: &drive,
	}
	err := share.SaveToCache()
	if err != nil {
		return err
	}
	return nil
}

func Download(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("%s - %s (%d)\n", r.Method, r.URL, r.ContentLength)
	download := make(map[string]int)

	data := make([]byte, r.ContentLength)
	_, err := r.Body.Read(data)

	if err != nil && err != io.EOF{
		log.Println("Cannot read request data: ", err)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(data, &download)
	if err != nil {
		log.Println("Cannot extract json data: ", err)
		return
	}
	// get share
	share := Share{Id: download["share_id"]}
	err = share.Manager("retrieve")

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	}
	presign_url, expires := getPresignURL(conf, "conduitcli-test-bucket", share.File.Name)
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
	mux.HandleFunc("/upload", Upload)
	mux.HandleFunc("/download", Download)

	mux.HandleFunc("/register", signUp)
	
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Server error: %s\n", err)
	} else {
		fmt.Println("Server started: Listening on port ", port)
	}
}