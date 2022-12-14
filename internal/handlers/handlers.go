package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/DDexster/golang_bookings/internal/config"
	"github.com/DDexster/golang_bookings/internal/driver"
	"github.com/DDexster/golang_bookings/internal/forms"
	"github.com/DDexster/golang_bookings/internal/helpers"
	"github.com/DDexster/golang_bookings/internal/models"
	"github.com/DDexster/golang_bookings/internal/renderer"
	"github.com/DDexster/golang_bookings/internal/repository"
	"github.com/DDexster/golang_bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["pageTitle"] = "Home Page"

	err := renderer.Template(w, r, "home.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "About Page"

	remoteIP := repo.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	err := renderer.Template(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "General's Quarters"

	err := renderer.Template(w, r, "generals.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Major's Suite"

	err := renderer.Template(w, r, "majors.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Reservation"

	data := make(map[string]interface{})
	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get Reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	dateLayout := "2006-01-02"
	sd := res.StartDate.Format(dateLayout)
	ed := res.EndDate.Format(dateLayout)
	repo.App.Session.Put(r.Context(), "reservation", res)
	data["reservation"] = res
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	err := renderer.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Form:      forms.New(nil),
		Data:      data,
	})

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't find Room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
}

func (repo *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse Form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

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

		dateLayout := "2006-01-02"
		sd := reservation.StartDate.Format(dateLayout)
		ed := reservation.EndDate.Format(dateLayout)

		stringMap["start_date"] = sd
		stringMap["end_date"] = ed
		stringMap["pageTitle"] = "Reservation"

		err = renderer.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Form:      form,
			Data:      data,
		})
		if err != nil {
			helpers.ServerError(w, err)
		}
		return
	}

	reservationId, err := repo.DB.InsertReservation(reservation)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Failed to Create Reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	repo.App.InfoLog.Println("Reservation ID: ", reservationId)

	restriction := models.RoomRestriction{
		RoomID:        reservation.RoomID,
		ReservationID: reservationId,
		RestrictionID: 1,
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
	}

	err = repo.DB.InsertRoomRestriction(restriction)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Failed to Create Restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Send confirm mails
	dateLayout := "2006-01-02"
	userMessage := fmt.Sprintf(`
		<strong>Reservation Succeed!</strong><br>
		<p>Hi, %s %s!</p>
		<p>Your Reservation for room <strong>%s</strong>, from %s to %s, has been placed!</p>
	`, reservation.FirstName, reservation.LastName, reservation.Room.RoomName, reservation.StartDate.Format(dateLayout), reservation.EndDate.Format(dateLayout))

	adminMessage := fmt.Sprintf(`
		<strong>New Reservation!</strong><br>
		<p>A reservation has been made from %s %s (%s). Room %s, from %s to %s</p>
	`, reservation.FirstName, reservation.LastName, reservation.Email, reservation.Room.RoomName, reservation.StartDate.Format(dateLayout), reservation.EndDate.Format(dateLayout))

	userMail := models.MailData{
		To:       reservation.Email,
		From:     "oficial@boonkings.here",
		Subject:  "Reservation Success!",
		Content:  userMessage,
		Template: "reservation.html",
	}

	adminMail := models.MailData{
		To:       "admin@boonkings.here",
		From:     "oficial@boonkings.here",
		Subject:  "New Reservation!",
		Content:  adminMessage,
		Template: "reservation.html",
	}

	repo.App.MailChan <- userMail
	repo.App.MailChan <- adminMail

	repo.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (repo *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Search Availability"

	err := renderer.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	startDateString := r.Form.Get("start")
	endDateString := r.Form.Get("end")
	dateLayout := "2006-01-02"
	startDate, err := time.Parse(dateLayout, startDateString)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't Parse Date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(dateLayout, endDateString)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't Parse Date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := repo.DB.SearchAvailabilityByDatesForAllRooms(startDate, endDate)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Failed to get rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		repo.App.Session.Put(r.Context(), "error", "No Rooms Available for given Dates!")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	stringMap := make(map[string]string)
	stringMap["pageTitle"] = "Choose Room"

	data := make(map[string]interface{})
	data["rooms"] = rooms

	preReservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	repo.App.Session.Put(r.Context(), "reservation", preReservation)

	err = renderer.Template(w, r, "choose_room.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (repo *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// need to parse request body
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	dateLayout := "2006-01-02"
	startDate, _ := time.Parse(dateLayout, sd)
	endDate, _ := time.Parse(dateLayout, ed)

	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := repo.DB.SearchAvailabilityByDatesByRoomId(startDate, endDate, roomId)

	if err != nil {
		// got a database error, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error querying database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		RoomID:    strconv.Itoa(roomId),
		StartDate: sd,
		EndDate:   ed,
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

	err := renderer.Template(w, r, "contact.page.tmpl", &models.TemplateData{
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

	dateLayout := "2006-01-02"
	sd := reservation.StartDate.Format(dateLayout)
	ed := reservation.EndDate.Format(dateLayout)

	data := make(map[string]interface{})
	data["reservation"] = reservation

	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	err := renderer.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "No Room ID provided")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	room, err := repo.DB.GetRoomById(roomId)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't find Room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	preReservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't find Room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	preReservation.RoomID = roomId
	preReservation.Room = room

	repo.App.Session.Put(r.Context(), "reservation", preReservation)
	http.Redirect(w, r, "/reservation", http.StatusSeeOther)
}

func (repo *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	dateLayout := "2006-01-02"
	startDate, err := time.Parse(dateLayout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(dateLayout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := repo.DB.GetRoomById(roomId)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		RoomID:    roomId,
		StartDate: startDate,
		EndDate:   endDate,
		Room:      room,
	}
	repo.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation", http.StatusSeeOther)
}

func (repo *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)

	stringMap["pageTitle"] = "Login"

	err := renderer.Template(w, r, "login.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Form:      forms.New(nil),
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.RenewToken(r.Context())

	userEmail := r.Form.Get("email")
	userPass := r.Form.Get("password")

	form := forms.New(r.PostForm)

	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		stringMap := make(map[string]string)
		stringMap["pageTitle"] = "Login"

		err := renderer.Template(w, r, "login.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Form:      form,
			Error:     "Form Is Not Valid!",
		})
		if err != nil {
			helpers.ServerError(w, err)
		}
		return
	}

	userId, _, err := repo.DB.Authenticate(userEmail, userPass)

	if err != nil {
		repo.App.ErrorLog.Println("Failed, to authenticate user with email", userEmail, "error", err)
		repo.App.Session.Put(r.Context(), "error", "Invalid Credentials!")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	repo.App.Session.Put(r.Context(), "user_id", userId)
	repo.App.Session.Put(r.Context(), "flash", "Login Success")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (repo *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.Destroy(r.Context())
	_ = repo.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (repo *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	err := renderer.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.ListNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})

	data["reservations"] = reservations

	err = renderer.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.ListAllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})

	data["reservations"] = reservations

	err = renderer.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) AdminReservationCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	next := now.AddDate(0, 1, 0)
	prev := now.AddDate(0, -1, 0)

	stringMap := make(map[string]string)
	dataMap := make(map[string]interface{})

	dataMap["now"] = now

	stringMap["next_month"] = next.Format("01")
	stringMap["next_month_year"] = next.Format("2006")
	stringMap["prev_month"] = prev.Format("01")
	stringMap["prev_month_year"] = prev.Format("2006")
	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := repo.DB.ListAllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	dataMap["rooms"] = rooms

	for _, room := range rooms {
		// generate block maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)
		dLayout := "2006-01-2"

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format(dLayout)] = 0
			blockMap[d.Format(dLayout)] = 0
		}

		// get all restrictions for room
		restrictions, e := repo.DB.GetRestrictionsForRoomByDates(room.ID, firstOfMonth, lastOfMonth)
		if e != nil {
			helpers.ServerError(w, e)
			return
		}

		for _, rr := range restrictions {
			for d := rr.StartDate; d.After(rr.EndDate) == false; d = d.AddDate(0, 0, 1) {
				if rr.ReservationID > 0 {
					reservationMap[d.Format(dLayout)] = rr.ReservationID
				} else {
					blockMap[d.Format(dLayout)] = rr.ID
				}
			}
		}

		dataMap[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		dataMap[fmt.Sprintf("block_map_%d", room.ID)] = blockMap

		repo.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	err = renderer.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      dataMap,
		IntMap:    intMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) AdminPostReservationCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	// process blocks
	rooms, err := repo.DB.ListAllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	fmt.Sprintf("%+v", form)
	// removing blocks
	for _, room := range rooms {
		//	get blockMap from session
		initialBlockMap := repo.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", room.ID)).(map[string]int)

		for name, value := range initialBlockMap {
			if val, ok := initialBlockMap[name]; ok {
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", room.ID, name)) {
						//	Remove restriction by ID
						log.Println("deleting block", value)
						err = repo.DB.RemoveOwnerBlock(value)
						if err != nil {
							helpers.ServerError(w, err)
							return
						}
					}
				}
			}
		}
	}

	//handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			splitted := strings.Split(name, "_")
			roomId, _ := strconv.Atoi(splitted[len(splitted)-2])
			dateString := splitted[len(splitted)-1]
			dateFormat := "2006-01-2"
			date, err := time.Parse(dateFormat, dateString)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
			log.Printf("adding block for room %d, and date %s", roomId, dateString)
			err = repo.DB.CreateOwnerBlock(roomId, date)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
		}
	}

	repo.App.Session.Put(r.Context(), "flash", "Changes Applied!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

func (repo *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	src := exploded[len(exploded)-3]

	id, err := strconv.Atoi(exploded[len(exploded)-2])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	reservation, err := repo.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	stringMap := make(map[string]string)
	data := make(map[string]interface{})

	data["reservation"] = reservation
	stringMap["src"] = src
	stringMap["month"] = month
	stringMap["year"] = year

	err = renderer.Template(w, r, "admin-reservation.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	})

	if err != nil {
		helpers.ServerError(w, err)
	}
}

func (repo *Repository) AdminUpdateReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	src := exploded[len(exploded)-2]

	id, err := strconv.Atoi(exploded[len(exploded)-1])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, err := repo.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = repo.DB.UpdateReservation(res)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.Form.Get("year")
	month := r.Form.Get("month")

	log.Println(year)

	repo.App.Session.Put(r.Context(), "flash", "Reservation Updated")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (repo *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = repo.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	repo.App.Session.Put(r.Context(), "flash", "Reservation Processed!")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (repo *Repository) AdminRemoveReservation(w http.ResponseWriter, r *http.Request) {
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = repo.DB.RemoveReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	repo.App.Session.Put(r.Context(), "flash", "Reservation Removed!")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}
