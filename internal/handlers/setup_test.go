package handlers

import (
	"encoding/gob"
	"fmt"
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/models"
	"github.com/DDexster/golang_bookings/internal/renderer"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var functions = template.FuncMap{
	"humanizeDate": renderer.HumanizeDate,
	"formatDate":   renderer.FormatDate,
	"iterate":      renderer.Iterate,
}

var app config.AppConfig
var session *scs.SessionManager
var pathToTmpl = "./../../templates"

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	app.UseCache = true
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = !app.InProduction

	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)

	listenForMail()

	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	app.InfoLog = infoLog
	app.ErrorLog = errorLog

	tc, err := CreateTestTemplateCache()

	if err != nil {
		log.Fatal("Cannot create template cache")
	}

	app.TemplateCache = tc

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	renderer.NewRenderer(&app)

	os.Exit(m.Run())
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)

	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)

	mux.Get("/reservation", Repo.Reservation)
	mux.Post("/reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/search-availability", Repo.SearchAvailability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/contact", Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))

	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Get("/contact", Repo.Contact)

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Route("/admin", func(m chi.Router) {
		m.Get("/dashboard", Repo.AdminDashboard)
		m.Get("/reservations-new", Repo.AdminNewReservations)
		m.Get("/reservations-all", Repo.AdminAllReservations)
		m.Get("/reservations-calendar", Repo.AdminReservationCalendar)
		m.Post("/reservations-calendar", Repo.AdminPostReservationCalendar)

		m.Get("/process-reservation/{src}/{id}/do", Repo.AdminProcessReservation)
		m.Get("/remove-reservation/{src}/{id}/do", Repo.AdminRemoveReservation)
		m.Get("/reservations/{src}/{id}/show", Repo.AdminShowReservation)
		m.Post("/reservations/{src}/{id}", Repo.AdminUpdateReservation)
	})

	return mux
}

func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}

func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := make(map[string]*template.Template)

	tmplFiles, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTmpl))
	if err != nil {
		return myCache, err
	}

	for _, tmpl := range tmplFiles {
		name := filepath.Base(tmpl)
		ts, err := template.New(name).Funcs(functions).ParseFiles(tmpl)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTmpl))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTmpl))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil
}
