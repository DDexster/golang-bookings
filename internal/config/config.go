package config

import (
	"github.com/DDexster/golang_bookings/internal/models"
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
}
