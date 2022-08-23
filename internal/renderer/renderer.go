package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/models"
	"github.com/justinas/nosurf"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var functions = template.FuncMap{
	"humanizeDate": HumanizeDate,
}

var app *config.AppConfig
var pathToTmpl = "./templates"

func NewRenderer(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("error getting template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	err := t.Execute(buf, td)

	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)

	if err != nil {
		return err
	}

	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
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

func HumanizeDate(t time.Time) string {
	dateFormat := "2006-01-02"
	return t.Format(dateFormat)
}
