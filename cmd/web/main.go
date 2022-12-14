package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/driver"
	"github.com/DDexster/golang_bookings/internal/handlers"
	"github.com/DDexster/golang_bookings/internal/helpers"
	"github.com/DDexster/golang_bookings/internal/models"
	"github.com/DDexster/golang_bookings/internal/renderer"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal("Error starting app", err)
	}

	defer close(app.MailChan)
	defer db.SQL.Close()

	log.Println("Email server listening...")
	listenForMail()

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	isProduction := flag.Bool("production", true, "Application mode")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user name")
	dbPass := flag.String("dbpass", "", "Database user password")
	dbPort := flag.String("dbport", "5432", "DB port")
	dbSSL := flag.String("dbssl", "disable", "DB ssl settings (disable, prefer, require)")

	flag.Parse()
	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.UseCache = *useCache
	app.InProduction = *isProduction

	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	app.InfoLog = infoLog
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = !app.InProduction

	app.Session = session

	// connect to DB
	log.Println("Connecting to DB")
	dbString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		*dbHost,
		*dbPort,
		*dbName,
		*dbUser,
		*dbPass,
		*dbSSL,
	)
	db, err := driver.ConnectSQL(dbString)
	if err != nil {
		log.Fatal("Cannot connect to DB")
	}
	log.Println("Connected to DB")

	tc, err := renderer.CreateTemplateCache()

	if err != nil {
		return nil, err
	}

	app.TemplateCache = tc

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	helpers.NewHelpers(&app)
	renderer.NewRenderer(&app)

	return db, nil
}
