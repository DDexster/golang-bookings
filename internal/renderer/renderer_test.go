package renderer

import (
	"github.com/DDexster/golang_bookings/internal/models"
	"net/http"
	"testing"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()

	if err != nil {
		t.Error(err)
	}
	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Failed to add Default Data")
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)
	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewTemplates(app)
}

func TestRenderTemplate(t *testing.T) {
	pathToTmpl = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
	app.TemplateCache = tc

	r, e := getSession()

	if e != nil {
		t.Error(e)
	}

	var ww myWriter

	err = RenderTemplate(&ww, r, "home.page.tmpl", &models.TemplateData{})

	if err != nil {
		t.Error(err)
	}
	err = RenderTemplate(&ww, r, "not-existing.page.tmpl", &models.TemplateData{})

	if err == nil {
		t.Error("got template that is not exist")
	}
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTmpl = "./../../templates"
	_, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}
}
