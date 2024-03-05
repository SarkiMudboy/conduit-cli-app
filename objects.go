package main

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
)

type Bucket struct {
	Db *sqlx.DB
	Id       int
	Name string
	URL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	Db *sqlx.DB
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
	Db *sqlx.DB
	Id int `json:"id"`
	From *User `json:"from"`
	To *User `json:"user"`
	File *Object `json:"object"`
	Note string `json:"note"`
	Drive *Drive `json:"drive"`
	Sent time.Time `json:"sent"`
}

// represents a cloud: shared/personal
type Drive struct {
	Db *sqlx.DB
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
	Db *sqlx.DB
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
		err = drive.Db.QueryRowx(
			"insert into drives (name, is_personal, owner_id, bucket_id, created_at) values ($1, $2, $3, $4, $5) returning id", 
			drive.Name, drive.IsPersonal, drive.Owner.Id, drive.Bucket.Id, timestamp).StructScan(&drive)
		
		// add the drive id to the owner's drive in the db
		update_query := fmt.Sprintf("update users set drives = drives || '{%d}' where id=$1", drive.Id)
		_, err = drive.Db.Exec(update_query, drive.Owner.Id)

		// add the owner to the membas of the drive
		updateMemberQuery := fmt.Sprintf("update drives set members = members || '{%d}' where id=$1", drive.Owner.Id)
		_, err = drive.Db.Exec(updateMemberQuery, drive.Id)

	case "retrieve":

		user := User{}
		bucket := Bucket{}

		var members []interface{}
		members_id := new([]interface{})

		var files []*Object

		err = drive.Db.QueryRowx("select name, is_personal, members, owner_id, bucket_id, created_at, updated_at from objects where id=$1", drive.Id).Scan(
			&drive.Name, &drive.IsPersonal, pq.Array(members_id), &user.Id, &bucket.Id, &drive.CreatedAt, &drive.UpdatedAt,
		)
		if err != nil {
			return err
		}
		
		drive.Owner = &user
		drive.Bucket = &bucket

		query := "select id, name, email, created_at from users where id in("
		members, _ = getObjectsByIds(query, User{}, members_id)

		for _, user := range members {
			if u, ok := user.(*User); ok {
				drive.Members = append(drive.Members, u)
			}
		}

		object_rows, err := drive.Db.Queryx("select id, name, is_dir, created_at, updated_at from objects where drive_id=$1", drive.Id)
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
		_, err = drive.Db.Exec("update drives set name=$2, is_personal=$3, owner_id=$4, bucket_id=$5, updated_at=$6 where id=$1", 
		drive.Name, drive.IsPersonal, drive.Owner.Id, drive.Bucket.Id, timestamp)

	case "delete":
		_, err = drive.Db.Exec("delete from users where id=$1", drive.Id)
	default: 
		return errors.New("invalid command")
	}
	
	return
}

func (object *Object) Manager(action string) (err error) {
	timestamp := time.Now()

	switch action {
	case "create":
		err = object.Db.QueryRowx(
			"insert into objects (name, is_dir, drive_id, created_at) values $1, $2, $3 returning id", 
			object.Name, object.IsDir, object.Drive.Id, timestamp).Scan(&object.Id)

	case "retrieve":

		drive := Drive{}

		err = object.Db.QueryRowx("select name, is_dir, drive_id, created_at, updated_at from objects where id=$1", object.Id).Scan(
			&object.Name, &object.IsDir, &drive.Id, &object.CreatedAt, &object.UpdatedAt,
		)
		if err != nil {
			return err
		}
		
		object.Drive = &drive

	case "update": 
		_, err = object.Db.Exec("update objects set name=$2 is_dir=$3 updated_at=$4 where id=$1", object.Id, object.Name, object.IsDir, timestamp)

	case "delete":
		_, err = object.Db.Exec("delete from objects where id=$1", object.Id)
	default: 
		return errors.New("invalid command")
	}

	return
}

func (user *User) Manager(action string) (err error) {

	timestamp := time.Now()

	switch action {
	case "create":

		err = user.Db.QueryRowx(
			"insert into users (name, username, email, password, created_at) values ($1, $2, $3, $4, $5) returning id", 
			user.Name, user.UserName, user.Email, user.Password, timestamp).Scan(&user.Id)

	case "retrieve":

		var drives []interface{}
		drives_id := new([]interface{})
		shares_id := new([]interface{})
		friends_id := new([]interface{})

		err = user.Db.QueryRowx(
			"select id, name, username, email, password, inbox, drives, friends, created_at, updated_at from users where id=$1", user.Id).Scan(
				&user.Id, &user.Name, &user.UserName, &user.Email, &user.Password, pq.Array(&shares_id), pq.Array(&drives_id), 
					pq.Array(&friends_id), &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return err
		}
		
		query := "select id, name, is_personal, created_at from drives where id in("
		drives, _ = getObjectsByIds(query, Drive{}, drives_id)
		for _, drive := range drives {
			if d, ok := drive.(*Drive); ok {
				user.Drives = append(user.Drives, d)
			}
		}

		query = "select id, name, username from users where id in("
		friends, _ := getObjectsByIds(query, User{}, friends_id)
		for _, friend := range friends {
			if u, ok := friend.(*User); ok {
				user.Friends = append(user.Friends, u)
			}
		}

	case "update": 
		_, err = user.Db.Exec("update users set name=$2 username=$3 email=$4 updated_at=$5 where id=$1", user.Id, user.Name, user.UserName, user.Email, timestamp)

	case "delete":
		_, err = user.Db.Exec("delete from users where id=$1", user.Id)
	default: 
		return errors.New("invalid command")
	}

	return 
}


func (user *User) GetObjectByField(fieldname string) error {
	// check to see if field exists
	u := reflect.ValueOf(user).Elem()
	field := u.FieldByName(fieldname)
	if !field.IsValid() {
		return errors.New("invalid field")
	}
	query := fmt.Sprintf("select id, name, email, password, inbox, drives, friends, created_at, updated_at from users where %s=$1", fieldname)
	return user.Db.QueryRowx(query, field.Interface()).StructScan(&user)
}


func (bucket *Bucket) Manager(action string) (err error) {
	timestamp := time.Now()

	switch action {
	case "create":
		err = bucket.Db.QueryRowx(
			"insert into buckets (name, url, created_at) values ($1, $2, $3) returning id", 
			bucket.Name, bucket.URL, timestamp).Scan(&bucket.Id)

	case "retrieve":

		err = bucket.Db.QueryRowx("select name, url, created_at, updated_at from buckets where id=$1", bucket.Id).Scan(
			&bucket.Name, &bucket.URL, &bucket.CreatedAt, &bucket.UpdatedAt,
		)
		if err != nil {
			return err
		}

	case "update": 
		_, err = bucket.Db.Exec("update buckets set name=$2 is_dir=$3 updated_at=$4 where id=$1", bucket.Id, bucket.Name, bucket.URL, timestamp)

	case "delete":
		_, err = bucket.Db.Exec("delete from buckets where id=$1", bucket.Id)
	default: 
		return errors.New("invalid command")
	}

	return
}


func (share *Share) Manager(action string) (err error) {
	switch action {
	case "create":
		err = share.Db.QueryRow("insert into shares (from_user, to_user, file, note, drive) values ($1, $2, $3, $4, $5) returning id", 
		share.From, share.To, share.File.Id, share.Note, share.Drive.Id).Scan(&share.Id)

		// add the file to the user's inbox
		q := fmt.Sprintf("update users set inbox = inbox || '{%d}' where id=$1", share.Id)
		_, err = share.Db.Exec(q, share.To.Id)

	default: 
		return errors.New("invalid command")
	}
	return
}