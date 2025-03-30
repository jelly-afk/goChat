package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"go.chat/internal/jwt"
	"go.chat/internal/models"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	users    *models.UserModel
	jwt      *jwt.Manager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:beans@/gochat?parseTime=true", "MySql dsn")
	secretKey := flag.String("secret-key", "your-secret-key", "JWT secret key")
	flag.Parse()
	infolog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorlog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorlog.Fatal(err)
	}
	defer db.Close()
	app := &application{
		errorLog: errorlog,
		infoLog:  infolog,
		users:    &models.UserModel{DB: db},
		jwt:      jwt.NewManager(*secretKey),
	}
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorlog,
		Handler:  app.routes(),
	}
	infolog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorlog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
