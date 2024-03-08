package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
)

type Model interface {
	Manager(string) error
}

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
	CreatedAt time.Time
	UpdatedAt time.Time
}

// stores details of a file share
type Share struct {
	Db *sqlx.DB
	Id int `json:"id"`
	From *DriveMember `json:"from"`
	To *DriveMember `json:"to"`
	File *Object `json:"object"`
	Note string `json:"note"`
	Saved bool `json:"saved"`
	Sent time.Time `json:"sent"`
}

// represents a cloud: shared/personal
type Drive struct {
	Db *sqlx.DB
	Id       int
	Name string
	IsPersonal bool
	Owner *User
	Members []*DriveMember
	Files []*Object
	Bucket   *Bucket
	CreatedAt time.Time
	UpdatedAt time.Time
}

// A member of a Drive
type DriveMember struct {
	Db *sqlx.DB
	Id int
	User *User
	Drive *Drive
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
		
		// // add the drive id to the owner's drive in the db
		// update_query := fmt.Sprintf("update users set drives = drives || '{%d}' where id=$1", drive.Id)
		// _, err = drive.Db.Exec(update_query, drive.Owner.Id)

		// add the owner to the membas of the drive
		// updateMemberQuery := fmt.Sprintf("update drives set members = members || '{%d}' where id=$1", drive.Owner.Id)
		// _, err = drive.Db.Exec(updateMemberQuery, drive.Id)

	case "retrieve":
		user := User{}
		bucket := Bucket{}
		var members []*DriveMember
		var files []*Object

		err = drive.Db.QueryRowx("select name, is_personal, owner_id, bucket_id, created_at, updated_at from objects where id=$1", drive.Id).Scan(
			&drive.Name, &drive.IsPersonal, &user.Id, &bucket.Id, &drive.CreatedAt, &drive.UpdatedAt,
		)
		if err != nil {
			return err
		}

		err = user.Manager("retrieve")
		err = bucket.Manager("retrieve")
		drive.Owner = &user
		drive.Bucket = &bucket

		rows, err := drive.Db.Queryx("select user_id, created_at, updated_at from drive_members where drive_id=$1", drive.Id)

		for rows.Next() {
			driveMember := DriveMember{Drive: drive}
			rows.StructScan(&driveMember)
			members = append(members, &driveMember)
		}
		drive.Members = members

		object_rows, err := drive.Db.Queryx("select id, name, is_dir, created_at, updated_at from objects where drive_id=$1", drive.Id)
		if err != nil {
			return err
		}

		for object_rows.Next() {
			obj := Object{Drive: drive}
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
		err = object.Db.QueryRowx("select name, is_dir, drive_id, created_at, updated_at from objects where id=$1", object.Id).Scan(
			&object.Name, &object.IsDir, &object.Drive.Id, &object.CreatedAt, &object.UpdatedAt,
		)
		if err != nil {
			return err
		}
		_ = object.Drive.Manager("retrieve")

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

		err = user.Db.QueryRow(
			"insert into users (name, username, email, password, created_at) values ($1, $2, $3, $4, $5) returning id", 
			user.Name, user.UserName, user.Email, user.Password, timestamp).Scan(&user.Id)

	case "retrieve":
		err = user.Db.QueryRowx(
			"select id, name, username, email, password, created_at, updated_at from users where id=$1", user.Id).StructScan(&user)
		if err != nil {
			return err
		}

		rows, err := user.Db.Query("select drive_id from drive_members where user_id=$1", user.Id) 
		if err != nil {
			return err
		}
		for rows.Next() {
			// try to use goruotines here
			var drive Drive
			err = rows.Scan(&drive.Id)
			if err != nil {
				return err
			}
			_ = drive.Manager("retrieve")
			user.Drives = append(user.Drives, &drive)
		}

		rows, err = user.Db.Query("select id, from_user, to_user, file, note, saved, sent from shares where saved=$1", false)
		for rows.Next() {
			var share Share
			err = rows.Scan(&share.Id, &share.From.Id, &share.To.Id, &share.File.Id, &share.Note, &share.Saved, &share.Sent)
			if err != nil {
				return err
			}
			// use go routines for each manager here
			_ = share.From.Manager("retrieve")
			_ = share.To.Manager("retrieve")
			_ = share.File.Manager("retrieve")
			user.Inbox = append(user.Inbox, &share)
		}
		// query := "select id, name, is_personal, created_at from drives where id in("
		// drives, _ = getObjectsByIds(query, Drive{}, drives_id)
		// for _, drive := range drives {
		// 	if d, ok := drive.(*Drive); ok {
		// 		user.Drives = append(user.Drives, d)
		// 	}
		// }

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
	timestamp := time.Now()

	switch action {
	case "create":
		err = share.Db.QueryRow("insert into shares (from_user, to_user, file, note, saved, sent) values ($1, $2, $3, $4, $5) returning id", 
		&share.From.Id, &share.To.Id, &share.File.Id, &share.Note, &share.Saved, &share.Sent).Scan(&share.Id)

	case "retrieve":
		err := share.Db.QueryRow("select from_user, to_user, file, note, saved, sent from shares where id=$1", share.Id).Scan(
			&share.From.Id, &share.To.Id, &share.File.Id, &share.Note, &share.Saved, &share.Sent, 
		)
		if err != nil {
			return err
		}

		err = share.From.Manager("retrieve")
		err = share.To.Manager("retrieve")
		_ = share.File.Manager("retrieve")

	case "update": 
		_, err = share.Db.Exec("update shares set note=$2 saved=$3 sent=$4 updated_at=$5 where id=$1", share.Id, share.Note, share.Saved, share.Sent, timestamp)

	case "delete":
		_, err = share.Db.Exec("delete from shares where id=$1", share.Id)
	default: 
		return errors.New("invalid command")
	}
	return
}

func (driveMember *DriveMember) Manager(action string) (err error) {
	timestamp := time.Now()
	switch action{
	case "create":
		err = driveMember.Db.QueryRow("insert into drive_members (user_id, drive_id, created_at) values ($1, $2, $3) returning id", driveMember.User.Id,
			driveMember.Drive.Id, timestamp).Scan(&driveMember.Id)
	case "retrieve":
		 err = driveMember.Db.QueryRow("select user_id, drive_id, created_at, updated_at from drive_members where id=$1", driveMember.Id).Scan(
			&driveMember.User.Id, &driveMember.Drive.Id, &driveMember.CreatedAt, &driveMember.UpdatedAt)
		if err != nil || err == sql.ErrNoRows {
			return
		}
		_ = driveMember.User.Manager("retrieve")
		_ = driveMember.Drive.Manager("retrieve")
	}
	
	return
}