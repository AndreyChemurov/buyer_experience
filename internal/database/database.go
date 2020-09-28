package database

import (
	"buyer_experience/internal/net"
	"buyer_experience/internal/types"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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

// CreateTables создает таблицу БД "subscribers",
// если она еще не была создана.
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

// CheckSubscriptionExists проверяет, что подписка на объявление уже существует.
//	Если да, то возвращается цена, чтобы не парсить ее заново
func CheckSubscriptionExists(link string) (price int, err error) {
	if db, err = openDB(); err != nil {
		return -1, err
	}

	defer db.Close()

	var (
		SQLStmt string   = `SELECT price FROM "subscribers" WHERE link=$1;`
		row     *sql.Row = db.QueryRow(SQLStmt, link)
	)

	if err = row.Scan(&price); err != nil {
		return -1, err
	}

	return price, nil
}

// DoubleSubscription проверяет, что пользователь не пытается
//	оформить подписку на одно и то же объявление
func DoubleSubscription(user *types.SubscriberResponse) (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	var (
		SQLStmt string   = `SELECT id FROM "subscribers" WHERE user_email=$1 AND link=$2;`
		row     *sql.Row = db.QueryRow(SQLStmt, user.Mail, user.Link)
		id      int      = -1
	)

	err = row.Scan(&id)

	if id != -1 { // Пользователь пытается подписаться дважды на одно объявление
		return errors.New("User already has subscription for this offer")
	} else if err == sql.ErrNoRows { // Нет ни одного пользователя
		return nil // -> можно подписаться
	} else {
		return err
	}
}

// CreateNewSubscription создает новую подписку
func CreateNewSubscription(user *types.SubscriberResponse, price int) (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	var (
		SQLStmt string = `INSERT INTO "subscribers" VALUES (DEFAULT, $1, $2, $3);`
	)

	if _, err = db.Exec(SQLStmt, user.Link, user.Mail, price); err != nil {
		return err
	}

	return nil
}

// CheckPriceChanged селектит инфу из бд для дальнейшего сравнения с объявлением на Авито
func CheckPriceChanged() {
	var (
		SQLStmt string = `SELECT DISTINCT link, user_email, price FROM "subscribers";`
		links   []string
		prices  []int
		emails  []string

		link  string
		price int
		email string
	)

	for {
		if db, err = openDB(); err != nil {
			log.Println(err)
			log.Fatal("database.go, 143")
			os.Exit(1)
		}

		defer db.Close()

		rows, err := db.Query(SQLStmt)

		if err != nil {
			//
		}

		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&link, &email, &price)

			if err != nil {
				//

				continue
			}
			links = append(links, link)
			emails = append(emails, email)
			prices = append(prices, price)
		}

		go net.RequestPrices(links, emails, prices)

		time.Sleep(time.Minute * 2)

		links = nil
	}
}
