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

func (repo *postgresDBRepo) ListAllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	stmt := `select id, room_name, created_at, updated_at from rooms order by id asc`

	rows, err := repo.DB.QueryContext(ctx, stmt)
	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Room
		err = rows.Scan(&i.ID, &i.RoomName, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, i)
	}
	if err != nil {
		return rooms, err
	}
	return rooms, nil
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

func (repo *postgresDBRepo) ListAllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	stmt := `select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		order by r.start_date desc
	`

	rows, err := repo.DB.QueryContext(ctx, stmt)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err = rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

func (repo *postgresDBRepo) ListNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	stmt := `select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name, r.processed
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.processed = 0
		order by r.created_at desc
	`

	rows, err := repo.DB.QueryContext(ctx, stmt)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err = rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
			&i.Processed,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

func (repo *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservation models.Reservation

	stmt := `select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name, r.processed
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.id = $1
    `

	row := repo.DB.QueryRowContext(ctx, stmt, id)
	err := row.Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
		&reservation.Processed,
	)

	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

func (repo *postgresDBRepo) UpdateReservation(res models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update reservations set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5 where id = $6`

	_, err := repo.DB.ExecContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		time.Now(),
		res.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (repo *postgresDBRepo) RemoveReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `delete from reservations where id = $1`

	_, err := repo.DB.ExecContext(ctx, stmt, id)

	if err != nil {
		return err
	}

	return nil
}

func (repo *postgresDBRepo) UpdateProcessedForReservation(id int, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update reservations set processed = $1 where id = $2`

	_, err := repo.DB.ExecContext(ctx, stmt, processed, id)

	if err != nil {
		return err
	}

	return nil
}

func (repo *postgresDBRepo) GetRestrictionsForRoomByDates(roomId int, startDate, endDate time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	stmt := `select id, coalesce(reservation_id, 0), restriction_id, start_date, end_date 
		from room_restrictions
		where $1 < end_date and $2 >= start_date and room_id = $3
		`

	rows, err := repo.DB.QueryContext(ctx, stmt, startDate, endDate, roomId)
	if err != nil {
		return restrictions, err
	}
	defer rows.Close()

	for rows.Next() {
		var rr models.RoomRestriction
		err = rows.Scan(
			&rr.ID,
			&rr.ReservationID,
			&rr.RestrictionID,
			&rr.StartDate,
			&rr.EndDate,
		)
		if err != nil {
			return restrictions, err
		}
		restrictions = append(restrictions, rr)
	}
	if err != nil {
		return restrictions, err
	}

	return restrictions, nil
}
