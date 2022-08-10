package handlers

import (
	"net/http"

	"github.com/DDexster/golang_bookings/pkg/config"
	"github.com/DDexster/golang_bookings/pkg/models"
	"github.com/DDexster/golang_bookings/pkg/renderer"
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

	remoteIP := r.RemoteAddr
	repo.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	renderer.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["test"] = "This is only a test...."
	stringMap["pageTitle"] = "About Page"

	remoteIP := repo.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	renderer.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (repo *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "General's Quarters"

	renderer.RenderTemplate(w, "generals.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (repo *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Major's Suite"

	renderer.RenderTemplate(w, "majors.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (repo *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Reservation"

	renderer.RenderTemplate(w, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (repo *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Search Availability"

	renderer.RenderTemplate(w, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (repo *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Contact Us"

	renderer.RenderTemplate(w, "contact.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
