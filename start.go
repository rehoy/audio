package main

import (
	"database/sql"
	_ "github.com/glebarez/go-sqlite"
	"fmt"
)

func main() {
	db, err := sql.Open("sqlite", "./my.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	var sqlVersion string
	db.QueryRow("select sqlite_version()").Scan(&sqlVersion)

	fmt.Printf("connected to sql and version is %s\n", sqlVersion)

	columns := map[string]string{
		"series_id": "INTEGER PRIMARY KEY",
		"name": "Text",
	}

	name := "series"
	err = createTable(db, name, columns)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("table %s created\n", name)


	

}

func createTable(db *sql.DB, tableName string, columns map[string]string) error {
	sql := "CREATE TABLE " + tableName + "("
	for name, dataType := range columns {
		sql += name + " " + dataType + ", "
	}
	sql = sql[:len(sql)-2] + ")"
	_, err := db.Exec(sql)

	return err
}
