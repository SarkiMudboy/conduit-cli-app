package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/jmoiron/sqlx"
)

var conf aws.Config
var port = ":8000"
var dependency Dependency

type Payload interface {
	Handle(*sqlx.DB, string) ([]byte, error)
}

type uploadData struct {
	Db *sqlx.DB
	From int `json:"from"`
	To int `json:"to"`
	FileName string `json:"filename"`
	IsDir bool `json:"is_dir"`
	Note string `json:"note"`
}

func(u uploadData) Handle(db *sqlx.DB, action string) (url []byte, err error) {
	switch action {
	case "POST":
		// check cache here...abstrct into Handle
		u.Db = db
		presign_url, err := putPresignURL(conf, "conduitcli-test-bucket", u.FileName)
		url = []byte(fmt.Sprintf(`{"presign_url": "%s"}`, presign_url))
		if err != nil {
			return nil, err
		}
		// save to cache...
		err = SaveFileToCache(u)
		// goroutine that waits for the time to elapse and queries AWS for file
		
	}
	return
}

func init() {
	conf = getConfig("us-east-1")
}

func Upload(d Dependency) http.HandlerFunc {

	return func (w http.ResponseWriter, r *http.Request) {
		// data in -> filename, from_username (drive_member), to_username, note, drive, is_dir 
		
		fmt.Printf("%s - %s\n", r.Method, r.URL)
		upload := d.payload
		data := make([]byte, r.ContentLength)

		_, err := r.Body.Read(data)

		if err != nil && err != io.EOF{
			log.Println("Cannot read request data: ", err)
			return
		}

		defer r.Body.Close()

		err = json.Unmarshal(data, &upload)
		if err != nil {
			log.Println("Cannot extract json data: ", err)
			return
		}

		url, err := upload.Handle(d.Db, r.Method)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(url)
	}
}


// func WaitForUpload(wait time.Duration, objectData) {
	// time.Sleep(wait)
	// check aws for the file
// }


func SaveFileToCache(data uploadData) error {
	// save to cache
	// get users
	Db := data.Db
	from := DriveMember{
		Db: Db,
		Id: data.From,
	}
	to := DriveMember{
		Db: Db,
		Id: data.To,
	}
	// fix to see if members and drive do not exist
	err := from.Manager("retrieve")
	if err == sql.ErrNoRows{
		from_user := User{Db: Db, Id: from.Id}
		to_user := User{Db: Db, Id: to.Id}
		err = from_user.Manager("retrieve")
		if err != nil {
			return err
		}
		err = to_user.Manager("retrieve")
		if err != nil {
			return err
		}

		drive := Drive{
			Db: Db,
			Name: from_user.UserName + to.User.UserName + " -drive",
			IsPersonal: false,
			Owner: &from_user,
			Bucket: &Bucket{Id: 1},
		}
		_ = drive.Manager("create")

		from.User = &from_user
		from.Drive = &drive
		to.User = &to_user
		to.Drive = &drive

		from.Manager("create")
		to.Manager("create")

	} else {
		_ = to.Manager("retrieve")
	}

	drive := from.Drive
	object := Object{
		Db: Db,
		Name: data.FileName,
		IsDir: data.IsDir,
		Drive: drive,
	}

	share := Share{
		Db: Db,
		From: &from,
		To: &to,
		File: &object,
		Note: data.Note,
		Saved: false,
	}
	err = share.SaveToCache()
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
	presign_url, err := getPresignURL(conf, "conduitcli-test-bucket", share.File.Name)
	url := []byte(fmt.Sprintf(`{"presign_url": "%s"}`, presign_url))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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
	mux.HandleFunc("/upload", Upload(dependency))
	mux.HandleFunc("/download", Download)

	mux.HandleFunc("/register", signUp(dependency))
	
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Server error: %s\n", err)
	} else {
		fmt.Println("Server started: Listening on port ", port)
	}
}