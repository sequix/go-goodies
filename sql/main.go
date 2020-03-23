package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
)

var (
	username = flag.String("username", "root", "mysql username")
	password = flag.String("password", "zc12345!", "mysql password")
	addr     = flag.String("addr", "127.0.0.1:3306", "mysql address")
	database = flag.String("database", "hello", "mysql database")
)

type User struct {
	// go底层使用int64表示整型，uint64可能会溢出
	ID         int64
	ResourceID string
	Name       string
	Age        int
	Desc       string
	DeletedAt  time.Time
	// mysql使用[]uint8存储timestamp
	DeletedAtString string
}

func main() {
	// 连接db
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s", *username, *password, *addr, *database)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("opening: %s", err)
	}
	defer dbClose(db)

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("pinging: %s", err)
	}

	// 设置参数
	db.SetConnMaxLifetime(5 * time.Minute) // if <= 0, always reuse connection.
	db.SetMaxIdleConns(2)                  // default 2. if <= 0, no idle connection retained.
	db.SetMaxOpenConns(0)                  // default 0, unlimited.

	// 执行sql
	execWithLog(db, `DROP TABLE IF EXISTS users`)
	statBytes, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("reading schema.sql: %s", err)
	}
	exec(db, string(statBytes))

	// sql 参数
	stmt, err := db.Prepare("INSERT INTO users (`resource_id`, `name`, `age`, `desc`) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("preparing statement: %s", err)
	}
	rst, err := stmt.Exec("u-1", "sequix", 24, "myself")
	if err != nil {
		log.Fatalf("inserting user1: %s", err)
	}
	logResult("insert-user1", rst)

	rst, err = stmt.Exec("u-2", "miyadi", 17, "someone")
	if err != nil {
		log.Fatalf("inserting user2: %s", err)
	}
	logResult("insert-user2", rst)

	// query rows
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatalf("selecting all: %s", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Fatalf("closing rows: %s", err)
		}
	}()
	u := User{}
	for rows.Next() {
		scanUser(&u, rows.Scan)
		spew.Dump(u)
	}

	// query & tx
	// tx, err := db.Begin()
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  true,
	})
	if err != nil {
		log.Fatalf("beginning tx: %s", err)
	}
	defer func() {
		if err != nil {
			if rtErr := tx.Rollback(); rtErr != nil {
				log.Fatalf("rolling back: %s", rtErr)
			}
		} else {
			if ctErr := tx.Commit(); ctErr != nil {
				log.Fatalf("commiting: %s", ctErr)
			}
		}
		cancel()
	}()

	// tx.QueryContext()
	row := tx.QueryRow("SELECT * FROM users WHERE resource_id = ?", "u-1")
	u2 := User{}
	scanUser(&u2, row.Scan)
	spew.Dump(u2)
}

func dbClose(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

func scanUser(u *User, scan func(...interface{}) error) {
	if err := scan(&u.ID, &u.ResourceID, &u.Name, &u.DeletedAtString, &u.Age, &u.Desc); err != nil {
		log.Fatalf("scaning user1: %s", err)
	}
	deletedAt, err := time.Parse("2006-01-02 15:04:05", u.DeletedAtString)
	if err != nil {
		log.Fatalf("parsing deletedAt: %s", err)
	}
	u.DeletedAt = deletedAt
}

func execWithLog(db *sql.DB, stat string) {
	rst, err := db.Exec(stat)
	if err != nil {
		log.Fatalf("%q: %s", stat, err)
	}
	logResult(stat, rst)
}

func exec(db *sql.DB, stat string) {
	if _, err := db.Exec(stat); err != nil {
		log.Fatalf("%q: %s", stat, err)
	}
}

func logResult(stat string, rst sql.Result) {
	lastInertId, err := rst.LastInsertId()
	if err != nil {
		log.Printf("%q getting LastInertId: %s", stat, err)
	} else {
		fmt.Printf("%q LastInertId: %v\n", stat, lastInertId)
	}
	rowsAffected, err := rst.RowsAffected()
	if err != nil {
		log.Printf("%q getting RowsAffected: %s", stat, err)
	} else {
		fmt.Printf("%q RowsAffected: %v\n", stat, rowsAffected)
	}
}
