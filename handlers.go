package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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

// getDestinationsHandler mengembalikan semua destinasi tanpa filter pencarian dan sorting (diproses di frontend)
func getDestinationsHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	allDestinations := GetDestinations()
	sendJSON(w, http.StatusOK, allDestinations)
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

	dest, found := BinarySearchById(id)

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

	SelectionSortCostSlice(data, "asc")

	sendJSON(w, http.StatusOK, data)
}

func insertionSortDistanceHandler(w http.ResponseWriter, r *http.Request) {

	if enableCORS(w, r) {
		return
	}

	data := GetDestinations()

	InsertionSortDistanceSlice(data, "asc")

	sendJSON(w, http.StatusOK, data)
}
