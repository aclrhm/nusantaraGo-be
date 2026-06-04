package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

// enableCORS adalah helper untuk menambahkan header CORS ke respon
func enableCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	// Jika preflight request (OPTIONS), kembalikan langsung dengan status 200 OK
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

// sendJSON adalah helper untuk mengirim respon dalam format JSON
func sendJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// getDestinationsHandler mengembalikan semua destinasi dengan filter pencarian, kategori, dan sorting
func getDestinationsHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	allDestinations := GetDestinations()

	// 1. Ekstrak filter query params
	searchQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	categoryFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	sortBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))

	filteredDestinations := []Destination{}

	// 2. Lakukan pencarian dan penyaringan kategori
	for _, d := range allDestinations {
		matchSearch := true
		matchCategory := true

		// Filter Kategori
		if categoryFilter != "" && categoryFilter != "semua" {
			if strings.ToLower(d.Category) != categoryFilter {
				matchCategory = false
			}
		}

		// Pencarian Kata Kunci
		if searchQuery != "" {
			found := strings.Contains(strings.ToLower(d.Name), searchQuery) ||
				strings.Contains(strings.ToLower(d.Description), searchQuery) ||
				strings.Contains(strings.ToLower(d.Location), searchQuery)
			
			// Cek juga di fasilitas
			if !found {
				for _, fac := range d.Facilities {
					if strings.Contains(strings.ToLower(fac), searchQuery) {
						found = true
						break
					}
				}
			}
			// Cek di wahana
			if !found {
				for _, ride := range d.Rides {
					if strings.Contains(strings.ToLower(ride), searchQuery) {
						found = true
						break
					}
				}
			}

			if !found {
				matchSearch = false
			}
		}

		if matchSearch && matchCategory {
			filteredDestinations = append(filteredDestinations, d)
		}
	}

	// 3. Pengurutan (Sorting)
	switch sortBy {
	case "distance_asc", "distance": // Jarak terdekat
		sort.Slice(filteredDestinations, func(i, j int) bool {
			return filteredDestinations[i].Distance < filteredDestinations[j].Distance
		})
	case "distance_desc": // Jarak terjauh
		sort.Slice(filteredDestinations, func(i, j int) bool {
			return filteredDestinations[i].Distance > filteredDestinations[j].Distance
		})
	case "cost_asc", "cost": // Biaya termurah
		sort.Slice(filteredDestinations, func(i, j int) bool {
			return filteredDestinations[i].Cost < filteredDestinations[j].Cost
		})
	case "cost_desc": // Biaya termahal
		sort.Slice(filteredDestinations, func(i, j int) bool {
			return filteredDestinations[i].Cost > filteredDestinations[j].Cost
		})
	case "facilities_desc", "facilities": // Fasilitas terbanyak
		sort.Slice(filteredDestinations, func(i, j int) bool {
			return len(filteredDestinations[i].Facilities) > len(filteredDestinations[j].Facilities)
		})
	case "facilities_asc": // Fasilitas tersedikit
		sort.Slice(filteredDestinations, func(i, j int) bool {
			return len(filteredDestinations[i].Facilities) < len(filteredDestinations[j].Facilities)
		})
	}

	sendJSON(w, http.StatusOK, filteredDestinations)
}

// getDestinationByIDHandler mengembalikan satu destinasi berdasarkan ID
func getDestinationByIDHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	id := r.PathValue("id")
	if id == "" {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing destination ID"})
		return
	}

	d, found := GetDestinationByID(id)
	if !found {
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Destination not found"})
		return
	}

	sendJSON(w, http.StatusOK, d)
}

// createDestinationHandler menambahkan data destinasi baru (Admin)
func createDestinationHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	var d Destination
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Generate ID acak unik jika tidak dikirim dari frontend
	if d.ID == "" {
		rand.Seed(time.Now().UnixNano())
		d.ID = fmt.Sprintf("dest-%d", rand.Intn(1000000))
	}

	// Validasi input minimal
	if d.Name == "" || d.Category == "" {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Name and Category are required fields"})
		return
	}

	err = CreateDestination(d)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusCreated, d)
}

// updateDestinationHandler mengubah destinasi yang ada (Admin)
func updateDestinationHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	id := r.PathValue("id")
	if id == "" {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing destination ID"})
		return
	}

	// Cek apakah data ada
	_, found := GetDestinationByID(id)
	if !found {
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Destination not found"})
		return
	}

	var d Destination
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	err = UpdateDestination(id, d)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, d)
}

// deleteDestinationHandler menghapus destinasi berdasarkan ID (Admin)
func deleteDestinationHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	id := r.PathValue("id")
	if id == "" {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing destination ID"})
		return
	}

	// Cek apakah data ada
	_, found := GetDestinationByID(id)
	if !found {
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Destination not found"})
		return
	}

	err := DeleteDestination(id)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"message": "Destination deleted successfully", "id": id})
}

func sequentialSearchHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		sendJSON(w, http.StatusBadRequest,
			map[string]string{
				"error": "id is required",
			})
		return
	}

	dest, found := SequentialSearchByID(id)

	if !found {
		sendJSON(w, http.StatusNotFound,
			map[string]string{
				"error": "destination not found",
			})
		return
	}

	sendJSON(w, http.StatusOK, dest)
}

func binarySearchHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		sendJSON(w, http.StatusBadRequest,
			map[string]string{
				"error": "id is required",
			})
		return
	}

	dest, found := BinarySearchByID(id)

	if !found {
		sendJSON(w, http.StatusNotFound,
			map[string]string{
				"error": "destination not found",
			})
		return
	}

	sendJSON(w, http.StatusOK, dest)
}


func selectionSortCostHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	data := GetDestinations()

	SelectionSortCostSlice(data)

	sendJSON(w, http.StatusOK, data)
}

func insertionSortDistanceHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	data := GetDestinations()

	InsertionSortDistanceSlice(data)

	sendJSON(w, http.StatusOK, data)
}
