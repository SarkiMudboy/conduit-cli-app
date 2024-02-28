package main

import (
	"bytes"
	// "encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func uploadFile(filepath string, url string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, file); err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPut, url, buff)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "multipart/form-data")
	client := &http.Client{}
	r, err := client.Do(request)

	fmt.Println(r.Status)
	return err
}

func downloadFile(filename, url string) (err error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	fmt.Println(response.Status)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := io.Copy(file, response.Body)
	fmt.Printf("Downloaded size -> %d\n", n)
	return err
}

func signup(data string) error {

	buff := bytes.NewBuffer(nil)
	url := "http://localhost:8000/register"
	reader := bytes.NewReader([]byte(data))
	if _, err := io.Copy(buff, reader); err != nil {
		return err
	}
	
	request, err := http.NewRequest(http.MethodGet, url, buff)
	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	fmt.Println(response.Status)
	return nil
}

// func main() {

// 	// sign up

// 	details := `{
// 		"email": "sharonresser33@gmail.com",
// 		"password": "45789",
// 		"fullname": "Sharon Resser"
// 	}`

// 	err := signup(details)
// 	if err != nil {
// 		panic(err)
// 	}

	// Download
	// filename := "cart.json"
	// file := []byte(`{"filename":"cart.json"}`)
	// reader := bytes.NewReader(file)

	// request, _ := http.NewRequest(http.MethodPost, "http://localhost:8000/get-download-url", reader)
	// client := &http.Client{}
	// response, err := client.Do(request)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(response.Status, response.ContentLength)
	
	// data := make(map[string]string)
	// dec := json.NewDecoder(response.Body)
	// err = dec.Decode(&data)
	
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// url := data["presign_url"]

	// err = downloadFile(filename, url)
	// if err != nil {
	// 	panic(err)

	// Upload

	// file := []byte(`{"filename":"cart.json"}`)
	// reader := bytes.NewReader(file)

	// request, _ := http.NewRequest(http.MethodPut, "http://localhost:8000/get-upload-url", reader)
	// client := &http.Client{}
	// response, err := client.Do(request)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(response.Status, response.ContentLength)

	// data := make(map[string]string)
	// dec := json.NewDecoder(response.Body)
	// err = dec.Decode(&data)
	
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// url := data["presign_url"]

	// err = uploadFile("cart.json", url)
	// if err != nil {
	// 	panic(err)
	// }

// }
