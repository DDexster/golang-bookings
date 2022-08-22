package dbrepo

import (
	"errors"
	"github.com/DDexster/golang_bookings/internal/models"
	"log"
	"time"
)

func (repo *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts new reservation in database
func (repo *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 1, nil
}

func (repo *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	return nil
}

// SearchAvailabilityByDatesByRoomId checks for existing reservations
func (repo *testDBRepo) SearchAvailabilityByDatesByRoomId(start time.Time, end time.Time, roomID int) (bool, error) {
	layout := "2006-01-02"
	str := "2049-12-31"
	t, err := time.Parse(layout, str)
	if err != nil {
		log.Println(err)
	}

	// this is our test to fail the query -- specify 2060-01-01 as start
	testDateToFail, err := time.Parse(layout, "2060-01-01")
	if err != nil {
		log.Println(err)
	}

	if start == testDateToFail {
		return false, errors.New("some error")
	}

	// if the start date is after 2049-12-31, then return false,
	// indicating no availability;
	if start.After(t) {
		return false, nil
	}

	// otherwise, we have availability
	return true, nil
}

func (repo *testDBRepo) SearchAvailabilityByDatesForAllRooms(start time.Time, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

func (repo *testDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("no room with ID provided exist")
	}
	return room, nil
}

func (repo *testDBRepo) GetUserById(id int) (models.User, error) {
	return models.User{}, nil
}

func (repo *testDBRepo) GetUserByEmail(email string) (models.User, error) {
	return models.User{}, nil
}

func (repo *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

func (repo *testDBRepo) Authenticate(email string, password string) (int, string, error) {
	return 1, "sdasdasdasdasda", nil
}
