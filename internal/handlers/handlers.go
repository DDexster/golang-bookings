package handlers

import (
	"encoding/json"
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/forms"
	"github.com/DDexster/golang_bookings/internal/helpers"
	"github.com/DDexster/golang_bookings/internal/models"
	"github.com/DDexster/golang_bookings/internal/renderer"
	"net/http"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pageTitle"] = "Home Page"

	err := renderer.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["test"] = "This is only a test...."
	stringMap["pageTitle"] = "About Page"

	remoteIP := repo.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	err := renderer.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "General's Quarters"

	err := renderer.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Major's Suite"

	err := renderer.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Reservation"

	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	err := renderer.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Form:      forms.New(nil),
		Data:      data,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.MinLength("last_name", 3)
	form.MinLength("phone", 8)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		stringMap := make(map[string]string)

		stringMap["pageTitle"] = "Reservation"

		err = renderer.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Form:      form,
			Data:      data,
		})
		if err != nil {
			helpers.ServerError(w, err)
		}
		return
	}

	repo.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (repo *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Search Availability"

	err := renderer.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	startDate := r.Form.Get("start")
	endDate := r.Form.Get("end")
	w.Write([]byte("Posted to search availability from " + startDate + " to " + endDate))
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (repo *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Contact Us"

	err := renderer.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pageTitle"] = "Reservation Succeed!"

	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.ErrorLog.Println("Cannot get item from session")
		repo.App.Session.Put(r.Context(), "error", "Can't find a reservation for You!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	repo.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	err := renderer.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}
