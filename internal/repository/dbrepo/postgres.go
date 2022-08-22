package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"github.com/DDexster/golang_bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

// SearchAvailabilityByDatesByRoomId checks for existing reservations
func (repo *postgresDBRepo) SearchAvailabilityByDatesByRoomId(start time.Time, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	stmt := `
		select count(id)
		from room_restrictions
		where
		    room_id = $1 and
		    $2 < end_date and $3 > start_date
	`

	row := repo.DB.QueryRowContext(ctx, stmt, roomID, start, end)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (repo *postgresDBRepo) SearchAvailabilityByDatesForAllRooms(start time.Time, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	stmt := `
		select
				r.id, r.room_name
		from
				rooms r
		where
				r.id not in (select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)
	`

	rows, err := repo.DB.QueryContext(ctx, stmt, start, end)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var room models.Room
		err = rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (repo *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	stmt := `select id, room_name from rooms where id = $1`

	row := repo.DB.QueryRowContext(ctx, stmt, id)
	err := row.Scan(&room.ID, &room.RoomName)
	if err != nil {
		return room, err
	}
	return room, nil
}

func (repo *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	stmt := `select id, first_name, last_name, email, password, access_level
		from users
		where id = $1
	`
	row := repo.DB.QueryRowContext(ctx, stmt, id)
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.AccessLevel)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (repo *postgresDBRepo) GetUserByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	stmt := `select id, first_name, last_name, email, password, access_level
		from users
		where email = $1
	`
	row := repo.DB.QueryRowContext(ctx, stmt, email)
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.AccessLevel)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (repo *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update users set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
		where id = $6
	`
	_, err := repo.DB.ExecContext(ctx, stmt,
		u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now(), u.ID)
	if err != nil {
		return err
	}
	return nil
}

func (repo *postgresDBRepo) Authenticate(email string, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPass string

	row := repo.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPass)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPass, nil
}
