package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

// struct
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

//batas maximum data lokal json
const MAX_DATA = 100

// Alias
type DestList = [MAX_DATA]Destination

var (
	destinations      DestList
	destinationsCount int
	dbMutex           sync.RWMutex
	dbFilePath        = "destinations.json"
)

// InitDB menginisialisasi database lokal berbasis JSON (Array statis)
func InitDB() error {
	fmt.Println("NusantaraGo diaktifkan dalam MODE JSON LOKAL.")
	fmt.Println("Semua data tersimpan aman secara offline di file 'destinations.json'.")
	return initLocalJSONDB()
}

func initLocalJSONDB() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, err := os.Stat(dbFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Membuat file database JSON lokal awal...")
		seedData, count := getSeedData()
		for i := 0; i < count; i++ {
			destinations[i] = seedData[i]
		}
		destinationsCount = count
		return saveToFileLocked()
	}

	fileData, err := ioutil.ReadFile(dbFilePath)
	if err != nil {
		return fmt.Errorf("gagal membaca database file lokal: %w", err)
	}

	var temp []Destination
	err = json.Unmarshal(fileData, &temp)
	if err != nil {
		fmt.Println("File database lokal rusak, melakukan seeding ulang...")
		seedData, count := getSeedData()
		for i := 0; i < count; i++ {
			destinations[i] = seedData[i]
		}
		destinationsCount = count
		return saveToFileLocked()
	}

	destinationsCount = len(temp)
	if destinationsCount > MAX_DATA {
		destinationsCount = MAX_DATA
	}
	for i := 0; i < destinationsCount; i++ {
		destinations[i] = temp[i]
	}

	return nil
}

func saveToFileLocked() error {
	temp := make([]Destination, destinationsCount)
	for i := 0; i < destinationsCount; i++ {
		temp[i] = destinations[i]
	}
	dataBytes, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dbFilePath, dataBytes, 0644)
}

// getSeedData menyediakan data pariwisata inisial Indonesia
func getSeedData() (DestList, int) {
	var data DestList
	data[0] = Destination{
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
	}
	data[1] = Destination{
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
	}
	data[2] = Destination{
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
	}
	data[3] = Destination{
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
	}
	data[4] = Destination{
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
	}
	data[5] = Destination{
		ID:          "dest-6",
		Name:        "Kebun Raya Bogor",
		Category:    "Alam",
		Cost:        25000,
		Distance:    60.0,
		Location:    "Bogor, Jawa Barat",
		Description: "Salah satu kebun raya tertua dan terbesar di Asia Tenggara yang didirikan sejak tahun 1817. Menawarkan koleksi ribuan spesies tanaman tropis yang rimbun dan segar di tengah kota Bogor.",
		Facilities:  []string{"Area Parkir Luas", "Toilet Bersih", "Mushola", "Pusat Informasi", "Toko Souvenir", "Kafe & Resto"},
		Rides:       []string{"Wisata Sepeda Keliling", "Tram Kebun Raya", "Museum Zoologi"},
		ImageURL:    "https://images.unsplash.com/photo-1585320806297-9794b3e4aaae?auto=format&fit=crop&w=800&q=80",
	}
	return data, 6
}

// GetDestinations membaca semua destinasi
func GetDestinations() (DestList, int) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	return destinations, destinationsCount
}

// GetDestinationByID mencari destinasi berdasarkan ID
func GetDestinationByID(id string) (Destination, bool) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	for i := 0; i < destinationsCount; i++ {
		if destinations[i].ID == id {
			return destinations[i], true
		}
	}
	return Destination{}, false
}

// CreateDestination menambahkan destinasi baru
func CreateDestination(d Destination) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if d.Name == "" || d.Category == "" {
		return errors.New("name and category cannot be empty")
	}
	if destinationsCount >= MAX_DATA {
		return errors.New("database penuh")
	}
	destinations[destinationsCount] = d
	destinationsCount++
	return saveToFileLocked()
}

// UpdateDestination mengubah destinasi yang ada berdasarkan ID
func UpdateDestination(id string, updated Destination) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	for i := 0; i < destinationsCount; i++ {
		if destinations[i].ID == id {
			updated.ID = id
			destinations[i] = updated
			return saveToFileLocked()
		}
	}
	return fmt.Errorf("destinasi dengan ID %s tidak ditemukan", id)
}

// DeleteDestination menghapus destinasi berdasarkan ID
func DeleteDestination(id string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	for i := 0; i < destinationsCount; i++ {
		if destinations[i].ID == id {
			for j := i; j < destinationsCount-1; j++ {
				destinations[j] = destinations[j+1]
			}
			destinationsCount--
			return saveToFileLocked()
		}
	}
	return fmt.Errorf("destinasi dengan ID %s tidak ditemukan", id)
}

// SEQUENTIAL SEARCH
func SequentialSearchByID(targetID string) (Destination, bool) {
	data, count := GetDestinations()
	for i := 0; i < count; i++ {
		if data[i].ID == targetID {
			return data[i], true
		}
	}
	return Destination{}, false
}

// SEQUENTIAL SEARCH KEYWORD
func SequentialSearchByKeyword(query string) (DestList, int) {
	data, count := GetDestinations()
	var results DestList
	var resCount int
	lowerQuery := strings.ToLower(strings.TrimSpace(query))

	for i := 0; i < count; i++ {
		dest := data[i]
		matchName := strings.Contains(strings.ToLower(dest.Name), lowerQuery)
		matchDesc := strings.Contains(strings.ToLower(dest.Description), lowerQuery)
		matchLoc := strings.Contains(strings.ToLower(dest.Location), lowerQuery)

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
			if resCount < MAX_DATA {
				results[resCount] = dest
				resCount++
			}
		}
	}

	return results, resCount
}

// BINARY SEARCH
func BinarySearchById(targetID string) (Destination, bool) {
	data, count := InsertionSortByID()

	low := 0
	high := count - 1

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


// SelectionSortCostSlice mengurutkan destinasi berdasarkan biaya (in-place)
func SelectionSortCostSlice(data *DestList, count int, order string) {
	for i := 0; i < count-1; i++ {
		selectIdx := i
		for j := i + 1; j < count; j++ {
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


// INSERTION SORT BERDASARKAN ID
// Digunakan untuk Binary Search
func InsertionSortByID() (DestList, int) {
	data, count := GetDestinations()

	for i := 1; i < count; i++ {
		key := data[i]
		j := i - 1
		for j >= 0 && data[j].ID > key.ID {
			data[j+1] = data[j]
			j--
		}
		data[j+1] = key
	}

	return data, count
}

// InsertionSortDistanceSlice mengurutkan destinasi berdasarkan jarak (in-place)
// order: "asc" = terdekat dulu, "desc" = terjauh dulu
func InsertionSortDistanceSlice(data *DestList, count int, order string) {
	for i := 1; i < count; i++ {
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
func InsertionSortFacilitiesSlice(data *DestList, count int, order string) {
	for i := 1; i < count; i++ {
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