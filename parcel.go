package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {

	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) values (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))

	if err != nil {
		return 0, fmt.Errorf("failed to add parcel: %w", err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {

	row := s.db.QueryRow("SELECT client, address, created_at, status FROM parcel WHERE number = :number",
		sql.Named("number", number))

	p := Parcel{}

	err := row.Scan(
		&p.Client,
		&p.Address,
		&p.CreatedAt,
		&p.Status,
	)

	switch {
	case err == nil:
		return p, nil
	case errors.Is(err, sql.ErrNoRows):
		return Parcel{}, fmt.Errorf("no row founded: %w", err)
	default:
		return Parcel{}, fmt.Errorf("bad query: %w", err)
	}

}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {

	rows, err := s.db.Query("SELECT number, client, address, created_at, status FROM parcel WHERE client = :client",
		sql.Named("client", client))

	if err != nil {
		return nil, fmt.Errorf("bad query: %w", err)
	}

	defer rows.Close()

	var res []Parcel

	for rows.Next() {

		parsel := Parcel{}

		err = rows.Scan(
			&parsel.Number,
			&parsel.Client,
			&parsel.Address,
			&parsel.CreatedAt,
			&parsel.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		res = append(res, parsel)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {

	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))

	if err != nil {
		return fmt.Errorf("can't update string with number %d: %w", number, err)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {

	var currentStatus string

	res := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))

	err := res.Scan(&currentStatus)

	if err != nil {
		return fmt.Errorf("failed to fetch status to parcel %d: %w", number, err)
	}

	if currentStatus != ParcelStatusRegistered {
		return fmt.Errorf("status mismatch for parcel %d: got %s, want %s", number, currentStatus, ParcelStatusRegistered)
	}

	_, err = s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
		sql.Named("address", address),
		sql.Named("number", number))

	if err != nil {
		return fmt.Errorf("failed to update address for parcel %d: %w", number, err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {

	var currentStatus string

	res := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))

	err := res.Scan(&currentStatus)
	if err != nil {
		return fmt.Errorf("failed to fetch status to parcel %d: %w", number, err)
	}

	if currentStatus != ParcelStatusRegistered {
		return fmt.Errorf("status mismatch for parcel %d: got %s, want %s", number, currentStatus, ParcelStatusRegistered)
	}

	_, err = s.db.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("failed to delete row number %d: %w", number, err)
	}

	return nil
}
