package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoErrorf(t, err, "failed to open database: %v")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoErrorf(t, err, "failed to add parcel: %v")
	require.NotZero(t, id, "expected non-zero ID after insert")

	saved, err := store.Get(id)
	require.NoErrorf(t, err, "failed to get parcel: %v")

	assert.Equal(t, parcel.Number, saved.Number)
	assert.Equal(t, parcel.Client, saved.Client)
	assert.Equal(t, parcel.Status, saved.Status)
	assert.Equal(t, parcel.Address, saved.Address)
	assert.Equal(t, parcel.CreatedAt, saved.CreatedAt)

	err = store.Delete(id)
	assert.NoErrorf(t, err, "failed to delete parcel: %v")

	_, err = store.Get(id)
	require.Error(t, err, "expected error when getting deleted parcel, got nil")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoErrorf(t, err, "failed to open database: %v")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoErrorf(t, err, "failed to add parcel: %v")
	require.NotZero(t, id, "expected non-zero ID after insert")

	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	assert.NoErrorf(t, err, "failed to set new address: %v")

	updated, err := store.Get(id)
	require.NoErrorf(t, err, "failed to get parcel: %v")

	assert.Equal(t, newAddress, updated.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoErrorf(t, err, "failed to open database: %v")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoErrorf(t, err, "failed to add parcel: %v")
	require.NotZero(t, id, "expected non-zero ID after insert")

	newStatus := ParcelStatusSent

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoErrorf(t, err, "failed to update status: %v")

	updated, err := store.Get(id)
	require.NoErrorf(t, err, "failed to get parcel: %v")

	assert.Equal(t, newStatus, updated.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoErrorf(t, err, "failed to open database: %v")

	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		assert.NoErrorf(t, err, "failed to add new parcel: %v")

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoErrorf(t, err, "failed to get parcel for client %d: %v", client)

	assert.Equal(t, len(storedParcels), len(parcels))

	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		require.True(t, ok, "unexpected parcel with number %d found in storedParcels", parcel.Number)

		assert.Equal(t, expected.Number, parcel.Number)
		assert.Equal(t, expected.Client, parcel.Client)
		assert.Equal(t, expected.Status, parcel.Status)
		assert.Equal(t, expected.Address, parcel.Address)
		assert.Equal(t, expected.CreatedAt, parcel.CreatedAt)
	}
}
