package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Destination merepresentasikan data tempat pariwisata
type Destination struct {
	ID          string   `json:"id" firestore:"id"`
	Name        string   `json:"name" firestore:"name"`
	Category    string   `json:"category" firestore:"category"`
	Cost        float64  `json:"cost" firestore:"cost"`
	Distance    float64  `json:"distance" firestore:"distance"`
	Location    string   `json:"location" firestore:"location"`
	Description string   `json:"description" firestore:"description"`
	Facilities  []string `json:"facilities" firestore:"facilities"`
	Rides       []string `json:"rides" firestore:"rides"`
	ImageURL    string   `json:"imageUrl" firestore:"imageUrl"`
}

var (
	// Local JSON Database Variables
	destinations []Destination
	dbMutex      sync.RWMutex
	dbFilePath   = "destinations.json"

	// Firebase Firestore Variables
	firestoreClient *firestore.Client
	isFirebaseMode  = false
	ctxTimeout      = 10 * time.Second
)

// InitDB menginisialisasi database
func InitDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	// 1. Periksa apakah file service account key Firebase ada
	credentialPath := "serviceAccountKey.json"
	if _, err := os.Stat(credentialPath); err == nil {
		fmt.Println("Menghubungkan ke Firebase Firestore Cloud Database...")
		
		opt := option.WithCredentialsFile(credentialPath)
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			return fmt.Errorf("gagal menginisialisasi firebase app: %w", err)
		}

		client, err := app.Firestore(ctx)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan client firestore: %w", err)
		}

		firestoreClient = client
		isFirebaseMode = true
		fmt.Println("SUKSES: NusantaraGo terhubung ke Cloud Firebase Firestore!")
		
		// Lakukan seeding otomatis di cloud jika koleksi destinations masih kosong
		err = seedFirestoreIfNeeded(ctx)
		if err != nil {
			fmt.Printf("Peringatan: Gagal memeriksa/seeding Firestore: %v\n", err)
		}
		return nil
	}

	// 2. Jika serviceAccountKey tidak ditemukan, jalankan Mode Fallback JSON Lokal
	fmt.Println("PERINGATAN: File 'serviceAccountKey.json' tidak ditemukan.")
	fmt.Println("NusantaraGo diaktifkan dalam MODE SIMULASI JSON LOKAL.")
	fmt.Println("Semua data tersimpan aman secara offline di file 'destinations.json'.")
	fmt.Println("Anda tetap dapat menguji CRUD, Filter, Search, dan Sorting secara utuh!")
	
	return initLocalJSONDB()
}

// seedFirestoreIfNeeded mengisi data awal di Firestore jika koleksi kosong
func seedFirestoreIfNeeded(ctx context.Context) error {
	// Cek apakah ada dokumen di koleksi "destinations"
	iter := firestoreClient.Collection("destinations").Limit(1).Documents(ctx)
	defer iter.Stop()
	_, err := iter.Next()
	
	if err == iterator.Done {
		// Koleksi kosong, lakukan seeding
		fmt.Println("Koleksi Firestore kosong. Melakukan seeding data awal ke cloud...")
		seedData := getSeedData()
		for _, dest := range seedData {
			_, err := firestoreClient.Collection("destinations").Doc(dest.ID).Set(ctx, dest)
			if err != nil {
				return fmt.Errorf("gagal menyemai destinasi %s: %w", dest.ID, err)
			}
		}
		fmt.Printf("Sukses mengunggah %d data awal pariwisata ke Firebase Firestore!\n", len(seedData))
	}
	return nil
}

// initLocalJSONDB menginisialisasi database file JSON offline
func initLocalJSONDB() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, err := os.Stat(dbFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Membuat file database JSON lokal awal...")
		destinations = getSeedData()
		return saveToFileLocked()
	}

	fileData, err := ioutil.ReadFile(dbFilePath)
	if err != nil {
		return fmt.Errorf("gagal membaca database file lokal: %w", err)
	}

	err = json.Unmarshal(fileData, &destinations)
	if err != nil {
		fmt.Println("File database lokal rusak, melakukan seeding ulang...")
		destinations = getSeedData()
		return saveToFileLocked()
	}

	return nil
}

func saveToFileLocked() error {
	dataBytes, err := json.MarshalIndent(destinations, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dbFilePath, dataBytes, 0644)
}

// getSeedData menyediakan data pariwisata inisial Indonesia
func getSeedData() []Destination {
	return []Destination{
		{
			ID:          "dest-1",
			Name:        "Candi Borobudur",
			Category:    "Budaya",
			Cost:        50000,
			Distance:    40.2,
			Location:    "Magelang, Jawa Tengah",
			Description: "Candi Buddha terbesar di dunia yang megah, diakui sebagai warisan budaya dunia oleh UNESCO. Menyajikan pemandangan matahari terbit yang sangat memukau di atas stupa kuno.",
			Facilities:  []string{"Area Parkir Luas", "Toilet Bersih", "Mushola", "Pemandu Wisata", "Toko Souvenir", "Pusat Informasi"},
			Rides:       []string{"Museum Borobudur", "Kereta Wisata Keliling", "Sewa Sepeda"},
			ImageURL:    "https://images.unsplash.com/photo-1584810359583-96fc3448beaa?auto=format&fit=crop&w=800&q=80",
		},
		{
			ID:          "dest-2",
			Name:        "Pantai Kuta Bali",
			Category:    "Alam",
			Cost:        15000,
			Distance:    12.5,
			Location:    "Badung, Bali",
			Description: "Pantai pasir putih legendaris yang terkenal di seluruh dunia. Merupakan pusat selancar, bersantai menikmati matahari terbenam yang romantis, dan memiliki garis pantai yang sangat panjang.",
			Facilities:  []string{"Shower Umum", "Penyewaan Payung Pantai", "Toilet", "Life Guard", "Area Parkir"},
			Rides:       []string{"Pusat Surfing (Selancar)", "Banana Boat", "Sewa Jet Ski"},
			ImageURL:    "https://images.unsplash.com/photo-1537996194471-e657df975ab4?auto=format&fit=crop&w=800&q=80",
		},
		{
			ID:          "dest-3",
			Name:        "Gunung Bromo",
			Category:    "Petualangan",
			Cost:        35000,
			Distance:    110.0,
			Location:    "Probolinggo, Jawa Timur",
			Description: "Gunung berapi aktif yang menawarkan pemandangan magis lautan pasir luas, kawah aktif yang menakjubkan, dan spot sunrise legendaris dari puncak Penanjakan.",
			Facilities:  []string{"Toilet Umum", "Mushola Penanjakan", "Warung Kopi", "Penyewaan Jaket Hangat", "Area Parkir Jeep"},
			Rides:       []string{"Wisata Jeep 4x4", "Berkuda di Pasir Berbisik", "Pendakian Kawah Bromo"},
			ImageURL:    "https://images.unsplash.com/photo-1602002418082-a4443e081dd1?auto=format&fit=crop&w=800&q=80",
		},
		{
			ID:          "dest-4",
			Name:        "Dunia Fantasi (Dufan)",
			Category:    "Rekreasi",
			Cost:        250000,
			Distance:    8.0,
			Location:    "Jakarta Utara, DKI Jakarta",
			Description: "Taman hiburan terbesar dan terlengkap di Indonesia yang terletak di kompleks Ancol. Menawarkan puluhan wahana memacu adrenalin dan rekreasi keluarga yang seru.",
			Facilities:  []string{"Ruang P3K", "Toilet AC & Ramah Anak", "Mushola Besar", "Food Court", "Ruang Menyusui", "Loker Penitipan Barang"},
			Rides:       []string{"Halilintar (Roller Coaster)", "Kora-Kora", "Bianglala Raksasa", "Istana Boneka"},
			ImageURL:    "https://images.unsplash.com/photo-1513885045260-6b15d6604b7a?auto=format&fit=crop&w=800&q=80",
		},
		{
			ID:          "dest-5",
			Name:        "Taman Wisata Alam Raja Ampat",
			Category:    "Alam",
			Cost:        500000,
			Distance:    32.0,
			Location:    "Raja Ampat, Papua Barat",
			Description: "Surga bawah laut terindah di dunia dengan keanekaragaman terumbu karang tertinggi. Menawarkan pemandangan gugusan pulau karang eksotis yang memanjakan mata.",
			Facilities:  []string{"Pusat Diving & Snorkeling", "Homestay Terapung", "Pemandu Lokal Berlisensi", "Kapal Speedboat"},
			Rides:       []string{"Diving (Menyelam)", "Snorkeling Bersama Penyu", "Island Hopping (Jelajah Pulau)"},
			ImageURL:    "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=800&q=80",
		},
	}
}

// GetDestinations membaca semua destinasi
func GetDestinations() []Destination {
	if isFirebaseMode {
		return getDestinationsFromFirestore()
	}

	dbMutex.RLock()
	defer dbMutex.RUnlock()
	dst := make([]Destination, len(destinations))
	copy(dst, destinations)
	return dst
}

func getDestinationsFromFirestore() []Destination {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	var results []Destination
	iter := firestoreClient.Collection("destinations").Documents(ctx)
	defer iter.Stop()

	done := false
	for !done {
		doc, err := iter.Next()
		if err == iterator.Done {
			done = true
		} else if err != nil {
			fmt.Printf("Gagal membaca dokumen Firestore: %v\n", err)
			done = true
		} else {
			var dest Destination
			if err := doc.DataTo(&dest); err != nil {
				fmt.Printf("Gagal mengurai dokumen Firestore: %v\n", err)
			} else {
				results = append(results, dest)
			}
		}
	}

	return results
}
// GetDestinationByID mencari destinasi berdasarkan ID
func GetDestinationByID(id string) (Destination, bool) {
	if isFirebaseMode {
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		defer cancel()

		doc, err := firestoreClient.Collection("destinations").Doc(id).Get(ctx)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return Destination{}, false
			}
			fmt.Printf("Gagal membaca dokumen Firestore %s: %v\n", id, err)
			return Destination{}, false
		}

		var dest Destination
		err = doc.DataTo(&dest)
		if err != nil {
			return Destination{}, false
		}
		return dest, true
	}

	// Local JSON fallback
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	for _, d := range destinations {
		if d.ID == id {
			return d, true
		}
	}
	return Destination{}, false
}

// CreateDestination menambahkan destinasi baru
func CreateDestination(d Destination) error {
	if isFirebaseMode {
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		defer cancel()

		_, err := firestoreClient.Collection("destinations").Doc(d.ID).Set(ctx, d)
		if err != nil {
			return fmt.Errorf("gagal menambahkan dokumen ke Firestore: %w", err)
		}
		return nil
	}

	// Local JSON fallback
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if d.Name == "" || d.Category == "" {
		return errors.New("name and category cannot be empty")
	}
	destinations = append(destinations, d)
	return saveToFileLocked()
}

// UpdateDestination mengubah destinasi yang ada berdasarkan ID
func UpdateDestination(id string, updated Destination) error {
	if isFirebaseMode {
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		defer cancel()

		updated.ID = id
		_, err := firestoreClient.Collection("destinations").Doc(id).Set(ctx, updated)
		if err != nil {
			return fmt.Errorf("gagal mengubah dokumen di Firestore: %w", err)
		}
		return nil
	}

	// Local JSON fallback
	dbMutex.Lock()
	defer dbMutex.Unlock()
	for i, d := range destinations {
		if d.ID == id {
			updated.ID = id
			destinations[i] = updated
			return saveToFileLocked()
		}
	}
	return fmt.Errorf("destinasi dengan ID %s tidak ditemukan", id)
}

// DeleteDestination menghapus destinasi berdasarkan ID
func DeleteDestination(id string) error {
	if isFirebaseMode {
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		defer cancel()

		_, err := firestoreClient.Collection("destinations").Doc(id).Delete(ctx)
		if err != nil {
			return fmt.Errorf("gagal menghapus dokumen dari Firestore: %w", err)
		}
		return nil
	}

	// Local JSON fallback
	dbMutex.Lock()
	defer dbMutex.Unlock()
	for i, d := range destinations {
		if d.ID == id {
			destinations = append(destinations[:i], destinations[i+1:]...)
			return saveToFileLocked()
		}
	}
	return fmt.Errorf("destinasi dengan ID %s tidak ditemukan", id)
}



// SEQUENTIAL SEARCH
func SequentialSearchByID(targetID string) (Destination, bool) {

	data := GetDestinations()

	for _, dest := range data {
		if dest.ID == targetID {
			return dest, true
		}
	}

	return Destination{}, false
}

// SEQUENTIAL SEARCH KEYWORD
func SequentialSearchByKeyword(query string) []Destination {
	data := GetDestinations()
	results := []Destination{}
	lowerQuery := strings.ToLower(strings.TrimSpace(query))

	for _, dest := range data {
		matchName := strings.Contains(strings.ToLower(dest.Name), lowerQuery)
		matchDesc := strings.Contains(strings.ToLower(dest.Description), lowerQuery)
		matchLoc  := strings.Contains(strings.ToLower(dest.Location), lowerQuery)

		matchFac := false
		for _, fac := range dest.Facilities {
			if strings.Contains(strings.ToLower(fac), lowerQuery) {
				matchFac = true
			}
		}

		matchRide := false
		for _, ride := range dest.Rides {
			if strings.Contains(strings.ToLower(ride), lowerQuery) {
				matchRide = true
			}
		}

		if matchName || matchDesc || matchLoc || matchFac || matchRide {
			results = append(results, dest)
		}
	}

	return results
}

// SELECTION SORT
// Mengurutkan destinasi berdasarkan biaya tiket
func SelectionSortByCost() []Destination {

	data := GetDestinations()

	n := len(data)

	for i := 0; i < n-1; i++ {

		minIdx := i

		for j := i + 1; j < n; j++ {

			if data[j].Cost < data[minIdx].Cost {
				minIdx = j
			}
		}

		data[i], data[minIdx] = data[minIdx], data[i]
	}

	return data
}

// SelectionSortCostSlice mengurutkan destinasi berdasarkan biaya (in-place)
func SelectionSortCostSlice(data []Destination, order string) {

	n := len(data)

	for i := 0; i < n-1; i++ {

		selectIdx := i

		for j := i + 1; j < n; j++ {
			if order == "desc" {
				if data[j].Cost > data[selectIdx].Cost {
					selectIdx = j
				}
			} else {
				if data[j].Cost < data[selectIdx].Cost {
					selectIdx = j
				}
			}
		}

		data[i], data[selectIdx] = data[selectIdx], data[i]
	}
}

// INSERTION SORT
// Mengurutkan destinasi berdasarkan jarak
func InsertionSortByDistance() []Destination {

	data := GetDestinations()

	for i := 1; i < len(data); i++ {

		key := data[i]
		j := i - 1

		for j >= 0 && data[j].Distance > key.Distance {

			data[j+1] = data[j]
			j--
		}

		data[j+1] = key
	}

	return data
}


// INSERTION SORT BERDASARKAN ID
// Digunakan untuk Binary Search
func InsertionSortByID() []Destination {

	data := GetDestinations()

	for i := 1; i < len(data); i++ {

		key := data[i]
		j := i - 1

		for j >= 0 && data[j].ID > key.ID {

			data[j+1] = data[j]
			j--
		}

		data[j+1] = key
	}

	return data
}

// BINARY SEARCH
func BinarySearchById(targetID string) (Destination, bool) {

	data := InsertionSortByID()

	low := 0
	high := len(data) - 1

	for low <= high {
		mid := (low + high) / 2
		if data[mid].ID == targetID {
			return data[mid], true
		} else if data[mid].ID < targetID {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return Destination{}, false
}

// InsertionSortDistanceSlice mengurutkan destinasi berdasarkan jarak (in-place)
// order: "asc" = terdekat dulu, "desc" = terjauh dulu
func InsertionSortDistanceSlice(data []Destination, order string) {

	for i := 1; i < len(data); i++ {

		
		key := data[i]

		j := i - 1


		if order == "desc" {
			for j >= 0 && data[j].Distance < key.Distance {
				data[j+1] = data[j]
				j--
			}
		} else {
			for j >= 0 && data[j].Distance > key.Distance {
				data[j+1] = data[j]
				j--
			}
		}

		data[j+1] = key
	}
}

// InsertionSortFacilitiesSlice mengurutkan destinasi berdasarkan jumlah fasilitas 
// order: "desc" = terlengkap dulu, "asc" = tersedikit dulu
func InsertionSortFacilitiesSlice(data []Destination, order string) {

	for i := 1; i < len(data); i++ {

		key := data[i]
		keyLen := len(key.Facilities)
		j := i - 1

		if order == "asc" {
			for j >= 0 && len(data[j].Facilities) > keyLen {
				data[j+1] = data[j]
				j--
			}
		} else {
			for j >= 0 && len(data[j].Facilities) < keyLen {
				data[j+1] = data[j]
				j--
			}
		}

		data[j+1] = key
	}
}
