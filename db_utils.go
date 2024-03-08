package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jmoiron/sqlx"
)


func getObjectsByIds(Db sqlx.DB, query string, TargetObject interface{}, IdList *[]interface{}) (objects []interface{}, err error) {
	// utility for getting row entries provided a list of their ids
	
	for i := range *IdList{

		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+1)
	}
	query += ")"

	rows, err := Db.Queryx(query, *IdList...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		rows.StructScan(&TargetObject)
		objects = append(objects, &TargetObject)
	}

	return
}


func (share *Share) SaveToCache() (err error) {
	// var shares map[string]Share

	shares := make(map[string]Share)
	file , err := os.Open("shares_cache.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	entries, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	if len(entries) != 0 {
		err = json.Unmarshal(entries, &shares)
	}

	shares[share.File.Name] = *share
	shareJSON, err := json.MarshalIndent(shares, "", "")

	if err != nil {
		return fmt.Errorf("cannot marshal (serialize) data: %s", err)
	}

	err = file.Truncate(0)
	if err != nil {
		fmt.Println(err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
	}
	_, err = file.Write(shareJSON)

	if err != nil {
		return fmt.Errorf("cannot save data: %s", err)
	}

	_ = file.Close()

	return
}