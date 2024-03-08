package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type AuthData struct {
	Db *sqlx.DB
	FullName string `json:"full_name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

func(auth *AuthData) Handle(db *sqlx.DB, action string) (username []byte, err error) {
	auth.Db = db
	// generate a username
	userName := getUserName(auth.FullName)
	// hash the password 
	encrypted_password := Encrypt(auth.Password)
	
	user := User{
		Db: auth.Db,
		Name: auth.FullName,
		UserName: userName,
		Email: auth.Email,
		Password: encrypted_password,
	}

	// store in Db
	err = user.Manager("create")
	if err != nil {
		return
	}

	bucket := Bucket{Db: auth.Db, Id:1}
	err = bucket.Manager("retrieve")
	if err != nil {
		return
	}
	fmt.Println(bucket)

	// create a personal drive for user
	drive := Drive{
		Db: auth.Db,
		Name: user.UserName + "'s drive",
		IsPersonal: true,
		Owner: &user,
		Bucket: &bucket,
	}

	err = drive.Manager("create")
	// return username in response
	username = []byte(fmt.Sprintf(`{"username": "%s"}`, userName))

	return
}


func signUp(d Dependency) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := d.payload
		// collect the email and password 
		fmt.Printf("%s - %s\n", r.Method, r.URL)
		data := make([]byte, r.ContentLength)

		_, err := r.Body.Read(data)

		if err != nil && err != io.EOF{
			log.Println("Cannot read request data: ", err)
			return
		}

		err = json.Unmarshal(data, &auth)
		if err != nil {
			log.Println(err)
		}
		log.Println(auth)
		// email, password, fullname := auth["email"], auth["password"], auth["fullname"]

		response, err := auth.Handle(d.Db, "POST")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(response)

		// later ..
		// token
		// validation
		// verification of email
	}

	
}

func createBucket() (bucket Bucket) {
	bucket = Bucket{
		Name: "conduitcli-test-bucket",
		URL: "",
	}
	err := bucket.Manager("create")
	if err != nil {
		log.Println(err)
	}
	return
}

// func authenticate(w http.ResponseWriter, r *http.Request) {

// }

func getUserName(fullname string) string {
	return fullname + "101" // handle this better...
}

func Encrypt(plaintext string) (cryptext string) {
	cryptext = fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
	return cryptext
}