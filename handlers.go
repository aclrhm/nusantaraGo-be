package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"strings"
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

// getDestinationsHandler mengembalikan semua destinasi tanpa filter pencarian dan sorting (diproses di frontend)
func getDestinationsHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	allDestinations, count := GetDestinations()
	sendJSON(w, http.StatusOK, allDestinations[:count])
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

	// Generate ID 
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

func binarySearchHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	id       := r.URL.Query().Get("id")
	category := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))

	if id != "" {
		dest, found := BinarySearchById(id)
		if !found {
			sendJSON(w, http.StatusNotFound, map[string]string{"error": "destination not found"})
			return
		}
		sendJSON(w, http.StatusOK, []Destination{dest})
		return
	}

	if category != "" {
		if category == "semua" {
			all, count := GetDestinations()
			sendJSON(w, http.StatusOK, all[:count])
			return
		}
		all, count := GetDestinations()
		var results [MAX_DATA]Destination
		var resCount int
		for i := 0; i < count; i++ {
			d := all[i]
			if strings.ToLower(d.Category) == category {
				results[resCount] = d
				resCount++
			}
		}
		sendJSON(w, http.StatusOK, results[:resCount])
		return
	}

	sendJSON(w, http.StatusBadRequest, map[string]string{"error": "parameter 'id' atau 'category' diperlukan"})
}

// sequentialSearchHandler menangani pencarian berurutan (Sequential Search)
// Mendukung pencarian berdasarkan ID (?id=) atau kata kunci multi-kolom (?q=)
func sequentialSearchHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	id := r.URL.Query().Get("id")
	keyword := r.URL.Query().Get("q")

	if id != "" {
		dest, found := SequentialSearchByID(id)
		if !found {
			sendJSON(w, http.StatusNotFound,
				map[string]string{
					"error": "destination not found",
				})
			return
		}
		
		sendJSON(w, http.StatusOK, []Destination{dest})
		return
	}

	if keyword != "" {
		results, count := SequentialSearchByKeyword(keyword)
		sendJSON(w, http.StatusOK, results[:count])
		return
	}

	sendJSON(w, http.StatusBadRequest,
		map[string]string{
			"error": "parameter 'id' atau 'q' (keyword) diperlukan",
		})
}

// binarySearchHandler menangani pencarian biner (Binary Search)
// Data diurutkan dulu berdasarkan ID menggunakan Insertion Sort
// func binarySearchHandler(w http.ResponseWriter, r *http.Request) {

// 	if enableCORS(w, r) {
// 		return
// 	}

// 	id := r.URL.Query().Get("id")

// 	if id == "" {
// 		sendJSON(w, http.StatusBadRequest,
// 			map[string]string{
// 				"error": "id is required",
// 			})
// 		return
// 	}

// 	dest, found := BinarySearchById(id)

// 	if !found {
// 		sendJSON(w, http.StatusNotFound,
// 			map[string]string{
// 				"error": "destination not found",
// 			})
// 		return
// 	}

// 	sendJSON(w, http.StatusOK, []Destination{dest})
// }

// selectionSortCostHandler menangani pengurutan berdasarkan biaya (Selection Sort)
// Mendukung parameter ?order=asc (default) atau ?order=desc
func selectionSortCostHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "asc"
	}

	data, count := GetDestinations()
	SelectionSortCostSlice(&data, count, order)

	sendJSON(w, http.StatusOK, data[:count])
}

// insertionSortDistanceHandler menangani pengurutan berdasarkan jarak (Insertion Sort)
// Mendukung parameter ?order=asc (default) atau ?order=desc
func insertionSortDistanceHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "asc"
	}

	data, count := GetDestinations()
	InsertionSortDistanceSlice(&data, count, order)

	sendJSON(w, http.StatusOK, data[:count])
}

// insertionSortFacilitiesHandler menangani pengurutan berdasarkan jumlah fasilitas (Insertion Sort)
// Mendukung parameter ?order=desc (default) atau ?order=asc
func insertionSortFacilitiesHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	data, count := GetDestinations()
	InsertionSortFacilitiesSlice(&data, count, order)

	sendJSON(w, http.StatusOK, data[:count])
}
