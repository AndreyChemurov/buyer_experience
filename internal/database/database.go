package database

import (
	"buyer_experience/internal/types"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq" //
)

const (
	host     string = "db"
	port     string = "5432"
	user     string = "postgres"
	password string = "postgres"
	dbname   string = "postgres"
)

var (
	psqlinfo string = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db  *sql.DB
	err error
)

func openDB() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", psqlinfo)

	if err != nil {
		return db, err
	}

	return db, nil
}

// CreateTables создает таблицы БД, "subscribers" и "users",
// если они еще не были созданы.
func CreateTables() (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(types.CreateTables)

	if err != nil {
		return err
	}

	return nil
}

// CreateUser проверяет, был ли пользователь с данным email'ом подписан на какое-то объявление (создан)
// Если нет, то создается новый пользователь.
func CreateUser(user *types.SubscriberResponse, price int) (id int, err error) {
	if db, err = openDB(); err != nil {
		return -1, err
	}

	defer db.Close()

	var (
		// SQLStmt string   = `SELECT EXISTS(SELECT 1 FROM "users" WHERE email=$1);`
		SQLStmt string   = `SELECT EXISTS(SELECT * FROM "subscribers" WHERE user_email=$1);`
		row     *sql.Row = db.QueryRow(SQLStmt, user.Mail)
		r       bool     // false, если пользователя с таким email ни на что не подписан
	)

	if err = row.Scan(&r); err != nil {
		return -1, err
	}

	if r != false {
		return -1, errors.New("User has subscription for this offer")
	}

	// Пользователь с таким email не имеет подписки на объявление

	SQLStmt = `ISERT INTO "subscribers" VALUES (DEFAULT, $1, $2, $3) RETURNING id;`

	if err = db.QueryRow(SQLStmt, user.Link, user.Mail, price).Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

// CheckSubscriptionExists ...
func CheckSubscriptionExists(link string) (price int, err error) {
	if db, err = openDB(); err != nil {
		return -1, err
	}

	defer db.Close()

	var (
		SQLStmt string   = `SELECT price FROM "subscibers" WHERE link=$1;`
		row     *sql.Row = db.QueryRow(SQLStmt, link)
	)

	if err = row.Scan(&price); err != nil {
		return -1, err
	}

	return price, nil
}

// CreateNewSubscription ...
func CreateNewSubscription(link string, id int, price int) (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	var (
		SQLStmt string = `INSERT INTO "subscribers" VALUES (DEFAULT, $1, $2, $3);`
	)

	if _, err = db.Exec(SQLStmt, link, id, price); err != nil {
		return err
	}

	return nil
}
