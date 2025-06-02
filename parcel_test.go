package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

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
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)

	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}

	if id == 0 {
		t.Fatal("expected non-zero ID after insert")
	}

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

	saved, err := store.Get(id)

	if err != nil {
		t.Fatalf("failed to get parcel: %v", err)
	}

	require.Equal(t, parcel.Client, saved.Client)
	require.Equal(t, parcel.Status, saved.Status)
	require.Equal(t, parcel.Address, saved.Address)
	require.Equal(t, parcel.CreatedAt, saved.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД

	err = store.Delete(id)

	if err != nil {
		t.Fatalf("failed to delete parcel: %v", err)
	}

	_, err = store.Get(id)
	if err == nil {
		t.Fatal("expected error when getting deleted parcel, got nil")
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)

	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}

	if id == 0 {
		t.Fatal("expected non-zero ID after insert")
	}

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)

	if err != nil {
		t.Fatalf("failed to set new address: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	updated, err := store.Get(id)

	if err != nil {
		t.Fatalf("failed to get parcel: %v", err)
	}

	require.Equal(t, newAddress, updated.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)

	if err != nil {
		t.Fatalf("failed to add parcel: %v", err)
	}

	if id == 0 {
		t.Fatal("expected non-zero ID after insert")
	}

	// set status
	newStatus := ParcelStatusSent

	err = store.SetStatus(id, ParcelStatusSent)

	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	updated, err := store.Get(id)

	if err != nil {
		t.Fatalf("failed to get parcel: %v", err)
	}

	require.Equal(t, newStatus, updated.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatalf("failed to add new parcel: %v", err)
		}
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	if err != nil {
		t.Fatalf("failed to get parcel for client %d: %v", client, err)
	}
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно

		// Проверяем, что parcel с таким number есть в карте
		expected, ok := parcelMap[parcel.Number]

		if !ok {
			t.Errorf("unexpected parcel with number %d found in storedParcels", parcel.Number)
			continue
		}

		require.Equal(t, expected.Client, parcel.Client)
		require.Equal(t, expected.Status, parcel.Status)
		require.Equal(t, expected.Address, parcel.Address)
		require.Equal(t, expected.CreatedAt, parcel.CreatedAt)
	}
}
