package dbrepo

import (
	"context"
	"fmt"
	"github.com/DDexster/golang_bookings/internal/models"
	"time"
)

func (repo *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts new reservation in database
func (repo *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newId int

	stmt := `insert into reservations 
    (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := repo.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}

	repo.App.InfoLog.Println(fmt.Sprintf("Inserted a new Reservation: %v", res))

	return newId, nil
}

func (repo *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions 
    (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
    values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.DB.ExecContext(ctx, stmt,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		res.ReservationID,
		res.RestrictionID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	repo.App.InfoLog.Println(fmt.Sprintf("Inserted a new Room Restriction: %v", res))

	return nil
}
