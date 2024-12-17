package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	err error
	// mc  *memcache.Client
)

type User struct {
	ID   int `gorm:"primaryKey"`
	Name string
}

func main() {
	username := os.Getenv("DB_USERNAME")
	userpass := os.Getenv("DB_USERPASS")
	dbName := os.Getenv("DB_DATABASENAME")
	DBInit(username, userpass, dbName)
	defer Close()
	insertData()

	// MemInit()
	// defer MemClose()

	mux := http.NewServeMux()
	mux.HandleFunc("/", index)

	sigctx, sigcancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigcancel()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("server start")
	done := make(chan error, 1)
	go func() {
		done <- srv.ListenAndServe()
		close(done)
	}()

	select {
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[Error] main server.ListenAndServe: %v", err)
		}
	case <-sigctx.Done():
		fmt.Println("Server stopping")
		timectx, timecancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer timecancel()
		err := srv.Shutdown(timectx)
		if err != nil {
			log.Printf("[Error] main server.Shutdown: %v", err)
		}
		fmt.Println("Server gracefully stopped")
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("view/index.html")
	if err != nil {
		log.Printf("template.ParseFiles error:%v\n", err)
		http.Error(w, "ページの表示に失敗しました。", http.StatusInternalServerError)
		return
	}

	var u []User
	db.Find(&u)

	if err := t.Execute(w, u); err != nil {
		log.Printf("template.Execute error:%v\n", err)
		http.Error(w, "ページの表示に失敗しました。", http.StatusInternalServerError)
		return
	}
}

// データベースと接続
func DBInit(user, password, dbName string) {
	dsn := user + ":" + password + "@tcp(db:3306)/" + dbName + "?charset=utf8&parseTime=true&loc=Asia%2FTokyo"
	fmt.Println("DB接続開始")
	// 接続できるまで一定回数リトライ
	count := 0
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		for {
			if err == nil {
				fmt.Println("")
				break
			}
			fmt.Print(".")
			time.Sleep(time.Second)
			count++
			if count > 180 { // countが180になるまでリトライ
				fmt.Println("")
				log.Printf("db Init error: %v\n", err)
				panic(err)
			}
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		}
	}
	autoMigration()

	fmt.Println("DB接続完了")
}

// modelでデータベースとやりとりする用
func GetDB() *gorm.DB {
	return db
}

// サーバ終了時にデータベースとの接続終了
func Close() {
	if sqlDB, err := db.DB(); err != nil {
		log.Printf("db Close error: %v\n", err)
		panic(err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("db Close error: %v\n", err)
			panic(err)
		}
	}
}

func autoMigration() {
	db.AutoMigrate(&User{})
}

func insertData() {
	var users []User
	for i := 0; i <= 100; i++ {
		users = append(users, User{Name: "User" + strconv.Itoa(i)})
	}
	db.Create(&users)
}

// func MemInit() {
// 	mc = memcache.New("127.0.0.1:11211")
// }

// func MemClose() {
// 	mc.Close()
// }
