package main

import "fmt"

func getObjectsByIds(query string, TargetObject interface{}, IdList *[]interface{}) (objects []interface{}, err error) {
	// utility for getting 
	
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