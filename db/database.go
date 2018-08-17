package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tacusci/berrycms/util"

	"github.com/tacusci/logging"

	//blank import to make sure right SQL driver is used to talk to DB
	_ "github.com/go-sql-driver/mysql"
)

var SchemaName string
var Conn *sql.DB

//Connect connects to database
func Connect(sqlDriver string, dbRoute string, schemaName string) {
	SchemaName = schemaName
	db, err := sql.Open(sqlDriver, dbRoute+SchemaName)
	if err != nil {
		logging.ErrorNnl(fmt.Sprintf(" DB error: %s\n", err.Error()))
	}
	err = db.Ping()
	if err != nil {
		logging.ErrorAndExit((fmt.Sprintf(" Error connecting to DB: %s", err.Error())))
		return
	}
	logging.GreenOutput(" Connected...\n")
	Conn = db
}

func Close() {
	if Conn != nil {
		Conn.Close()
	}
}

//CreateTestData fill database with known test data for development/testing purposes
func CreateTestData() {
	usersTable := &UsersTable{}
	err := usersTable.Insert(Conn, User{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "jdoe",
		AuthHash:  util.HashAndSalt([]byte("iamjohndoe")),
		Email:     "person@place.com",
	})
	rootUser := User{}
	rows, err := usersTable.Select(Conn, "UUID", fmt.Sprintf("userid = 1 AND userroleid = %d", ROOT))
	if err != nil {
		logging.Error("Unable to fetch root user...")
		return
	}
	for rows.Next() {
		rows.Scan(&rootUser.UUID)
	}
	pagesTable := &PagesTable{}
	err = pagesTable.Insert(Conn, Page{
		AuthorUUID: rootUser.UUID,
		Title:      "Add New",
		Route:      "/addnew",
		Content:    "<html><body><h2>Adding Carbonite page...</h2></body></html>",
	})
	if err != nil {
		logging.Error(err.Error())
	}
}

func Heartbeat() {
	for {
		<-time.After(time.Second * 60)
		err := Conn.Ping()
		if err != nil {
			logging.Error(fmt.Sprintf("DB Ping error -> %s", err.Error()))
		}
	}
}

//Wipe drops all database tables
func Wipe() error {
	logging.Debug("Dropping/wiping all existing tables...")
	for _, tableToDrop := range getTables() {
		logging.Debug(fmt.Sprintf("Dropping %s table...", tableToDrop.Name()))
		dropSmt := fmt.Sprintf("DROP TABLE %s;", tableToDrop.Name())
		_, err := Conn.Exec(dropSmt)
		if err != nil {
			return err
		}
	}
	return nil
}

//Setup constructs all the tables etc.,
func Setup() {
	if Conn == nil {
		return
	}
	logging.Info("Setting up DB...")
	createTables(Conn)
}

func createTables(db *sql.DB) {
	logging.Debug("Creating all database tables...")
	tablesToCreate := getTables()
	for _, tableToCreate := range tablesToCreate {
		tableCreateStatement := createStatement(tableToCreate)

		logging.Debug(fmt.Sprintf("Creating table %s...", tableToCreate.Name()))
		logging.Debug(fmt.Sprintf("Running create statement: \"%s\"", tableCreateStatement))

		_, err := db.Exec(tableCreateStatement)
		tableToCreate.Init(db)

		if err != nil {
			logging.Error(err.Error())
		}
	}
}

func getTables() []Table {
	return []Table{&UsersTable{}, &PagesTable{}, &AuthSessionsTable{}}
}
