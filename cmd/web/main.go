package main

import (
	"encoding/gob"
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/handlers"
	"github.com/DDexster/golang_bookings/internal/models"
	"github.com/DDexster/golang_bookings/internal/renderer"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	gob.Register(models.Reservation{})

	app.UseCache = false
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = !app.InProduction

	app.Session = session

	tc, err := renderer.CreateTemplateCache()

	if err != nil {
		log.Fatal("Error creating templates", err)
	}

	app.TemplateCache = tc

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	renderer.NewTemplates(&app)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
