package repository

import (
	"github.com/DDexster/golang_bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomId(start time.Time, end time.Time, roomID int) (bool, error)
	SearchAvailabilityByDatesForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomById(id int) (models.Room, error)
	ListAllRooms() ([]models.Room, error)

	GetUserById(id int) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email string, password string) (int, string, error)

	ListAllReservations() ([]models.Reservation, error)
	ListNewReservations() ([]models.Reservation, error)
	GetReservationById(id int) (models.Reservation, error)
	UpdateReservation(res models.Reservation) error
	RemoveReservation(id int) error
	UpdateProcessedForReservation(id int, processed int) error

	GetRestrictionsForRoomByDates(roomId int, startDate, endDate time.Time) ([]models.RoomRestriction, error)
	CreateOwnerBlock(roomId int, date time.Time) error
	RemoveOwnerBlock(restrictionId int) error
}
