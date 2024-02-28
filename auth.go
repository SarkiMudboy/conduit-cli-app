package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)


func signUp(w http.ResponseWriter, r *http.Request) {

	// collect the email and password 
	fmt.Printf("%s - %s\n", r.Method, r.URL)
	auth := make(map[string]string)
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
	email, password, fullname := auth["email"], auth["password"], auth["fullname"]

	// generate a username
	userName := getUserName(fullname)
	// hash the password 
	encrypted_password := Encrypt(password)
	
	user := User{
		Name: fullname,
		UserName: userName,
		Email: email,
		Password: encrypted_password,
	}

	// store in Db
	err = user.Manager("create")
	if err != nil {
		log.Println(err)
	}

	bucket := Bucket{Id:1}
	err = bucket.Manager("retrieve")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(bucket)

	// create a personal drive for user
	drive := Drive{
		Name: user.UserName + "'s drive",
		IsPersonal: true,
		Owner: &user,
		Members: []*User{&user},
		Bucket: &bucket,
	}

	err = drive.Manager("create")
	if err != nil {
		log.Println(err)
	}

	// return username in response
	url := []byte(fmt.Sprintf(`{"username": "%s"}`, userName))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(url)

	// later ..
	// token
	// validation
	// verification of email
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