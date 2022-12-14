package main

import (
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)
	mux.Use(NoSurve)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)

	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)

	mux.Get("/reservation", handlers.Repo.Reservation)
	mux.Post("/reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/search-availability", handlers.Repo.SearchAvailability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	fileServer := http.FileServer(http.Dir("./static/"))

	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(m chi.Router) {
		m.Use(Auth)

		m.Get("/dashboard", handlers.Repo.AdminDashboard)
		m.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		m.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		m.Get("/reservations-calendar", handlers.Repo.AdminReservationCalendar)
		m.Post("/reservations-calendar", handlers.Repo.AdminPostReservationCalendar)

		m.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		m.Get("/remove-reservation/{src}/{id}/do", handlers.Repo.AdminRemoveReservation)
		m.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminShowReservation)
		m.Post("/reservations/{src}/{id}", handlers.Repo.AdminUpdateReservation)
	})

	return mux
}
