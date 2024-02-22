package main

import (
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Bucket struct {
	Id       int
	Name string
	URL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	Id       int
	Name string
	UserName string
	Email string
	Password string
	Inbox []*Share
	Drives []*Drive
	Friends []*User
	CreatedAt time.Time
	UpdatedAt time.Time
}

// stores details of a file share
type Share struct {
	Id int
	From *User
	To *User
	File *Object
	Note string
	Drive *Drive
	Sent time.Time
}

// represents a cloud: shared/personal
type Drive struct {
	Id       int
	Name string
	IsPersonal bool
	Owner *User
	Members []*User
	Files []*Object
	Bucket   *Bucket
	CreatedAt time.Time
	UpdatedAt time.Time
}

// can represent a file or a folder (if its a folder: IsDir is true)
type Object struct {
	Id       int
	Name string
	IsDir bool
	Drive *Drive
	Metadata map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

//  add the "all" operation
func (drive *Drive) Manager(action string) (err error) {
	timestamp := time.Now()

	switch action {
	case "create":
		err = Db.QueryRow(
			"insert into drives (name, is_personal, owner_id, bucket_id, created_at, updated_at) values $1, $2, $3 returning id", 
			drive.Name, drive.IsPersonal, drive.Owner.Id, drive.Bucket.Id, timestamp).Scan(&drive.Id)
		
		// add the drive id to the owner's drive in the db
		update_query := fmt.Sprintf("update users set drives = drives || '{%d}' where id=$1", drive.Id)
		_ = Db.QueryRow(update_query, drive.Owner.Id)

		// add the owner to the membas of the drive
		updateMemberQuery := fmt.Sprintf("update users set members = members || {%d} where id=$1", drive.Owner.Id)
		_ = Db.QueryRow(updateMemberQuery, drive.Id)

	case "retrieve":

		user := User{}
		bucket := Bucket{}
		members_id := new([]interface{})
		members := []*User{}
		files := []*Object{}

		err = Db.QueryRow("select name, is_personal, members, owner_id, bucket_id, created_at, updated_at from objects where id=$1", drive.Id).Scan(
			&drive.Name, &drive.IsPersonal, pq.Array(members_id), &user.Id, &bucket.Id, &drive.CreatedAt, &drive.UpdatedAt,
		)
		if err != nil {
			return err
		}
		
		drive.Owner = &user
		drive.Bucket = &bucket

		query := "select id, name, email, created_at from users where id in("
		for i := range *members_id{

			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf("$%d", i+1)
		}
		query += ")"

		rows, err := Db.Queryx(query, *members_id...)
		if err != nil {
			return err
		}
		for rows.Next() {
			user := User{}
			rows.StructScan(&user)
			members = append(members, &user)
		}

		drive.Members = members

		object_rows, err := Db.Queryx("select id, name, is_dir, created_at, updated_at from objects where drive_id=$1", drive.Id)
		if err != nil {
			return err
		}

		for object_rows.Next() {
			obj := Object{}
			object_rows.StructScan(&obj)
			files = append(files, &obj)
		}

		drive.Files = files

	case "update": 
		_, err = Db.Exec("update drives set name=$2, is_personal=$3, owner_id=$4, bucket_id=$5, updated_at=$6 where id=$1", 
		drive.Name, drive.IsPersonal, drive.Owner.Id, drive.Bucket.Id, timestamp)

	case "delete":
		_, err = Db.Exec("delete from users where id=$1", drive.Id)
	
	}
	
	return
}

func (object *Object) Manager(action string) (err error) {
	timestamp := time.Now()

	switch action {
	case "create":
		err = Db.QueryRow(
			"insert into objects (name, is_dir, drive_id, created_at) values $1, $2, $3 returning id", 
			object.Name, object.IsDir, object.Drive.Id, timestamp).Scan(&object.Id)

	case "retrieve":

		drive := Drive{}

		err = Db.QueryRow("select name, is_dir, drive_id, created_at, updated_at from objects where id=$1", object.Id).Scan(
			&object.Name, &object.IsDir, &drive.Id, &object.CreatedAt, &object.UpdatedAt,
		)
		if err != nil {
			return err
		}
		
		object.Drive = &drive

	case "update": 
		_, err = Db.Exec("update objects set name=$2 is_dir=$3 updated_at=$4 where id=$1", object.Id, object.Name, object.IsDir, timestamp)

	case "delete":
		_, err = Db.Exec("delete from objects where id=$1", object.Id)
	}

	return
}

func (user *User) Manager(action string) (err error) {

	timestamp := time.Now()

	switch action {
	case "create":

		err = Db.QueryRow(
			"insert into users (name, email, created_at) values $1, $2, $3 returning id", 
			user.Name, user.Email, timestamp).Scan(&user.Id)

	case "retrieve":

		drives := *new([]*Drive)
		drives_id := new([]interface{})

		err = Db.QueryRow("select id, name, email, drives, created_at, updated_at from users where id=$1", user.Id).Scan(&user.Id, 
		&user.Name, &user.Email, pq.Array(&drives_id), &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return err
		}
		
		query := "select id, name, is_personal, created_at from drives where id in("
		for i := range *drives_id{

			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf("$%d", i+1)
		}
		query += ")"

		rows, err := Db.Queryx(query, *drives_id...)
		if err != nil {
			return err
		}
		for rows.Next() {
			drive := Drive{}
			rows.StructScan(&drive)
			drives = append(drives, &drive)
		}

		user.Drives = drives

	case "update": 
		_, err = Db.Exec("update users set name=$2 email=$3 updated_at=$4 where id=$1", user.Id, user.Name, user.Email, timestamp)

	case "delete":
		_, err = Db.Exec("delete from users where id=$1", user.Id)
	}

	return 
}


func (bucket *Bucket) Manager(action string) (err error) {
	timestamp := time.Now()

	switch action {
	case "create":
		err = Db.QueryRow(
			"insert into buckets (name, url, created_at) values $1, $2, $3 returning id", 
			bucket.Name, bucket.URL, timestamp).Scan(&bucket.Id)

	case "retrieve":

		err = Db.QueryRow("select name, url, created_at, updated_at from buckets where id=$1", bucket.Id).Scan(
			&bucket.Name, &bucket.URL, &bucket.CreatedAt, &bucket.UpdatedAt,
		)
		if err != nil {
			return err
		}

	case "update": 
		_, err = Db.Exec("update buckets set name=$2 is_dir=$3 updated_at=$4 where id=$1", bucket.Id, bucket.Name, bucket.URL, timestamp)

	case "delete":
		_, err = Db.Exec("delete from buckets where id=$1", bucket.Id)
	}

	return
}
