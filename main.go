package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	err := InitDB()
	if err != nil {
		log.Fatalf("Fatal error: Gagal menginisialisasi basis data: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/destinations/sequential-search",         sequentialSearchHandler)
	mux.HandleFunc("GET /api/destinations/binary-search",             binarySearchHandler)
	mux.HandleFunc("GET /api/destinations/selection-sort-cost",       selectionSortCostHandler)
	mux.HandleFunc("GET /api/destinations/insertion-sort-distance",   insertionSortDistanceHandler)
	mux.HandleFunc("GET /api/destinations/insertion-sort-facilities", insertionSortFacilitiesHandler)

// {id} HARUS di bawah semua route spesifik
	mux.HandleFunc("GET /api/destinations/{id}",    getDestinationByIDHandler)
	mux.HandleFunc("PUT /api/destinations/{id}",    updateDestinationHandler)
	mux.HandleFunc("DELETE /api/destinations/{id}", deleteDestinationHandler)

	mux.HandleFunc("GET /api/destinations",  getDestinationsHandler)
	mux.HandleFunc("POST /api/destinations", createDestinationHandler)

	mux.HandleFunc("OPTIONS /api/destinations",      func(w http.ResponseWriter, r *http.Request) { enableCORS(w, r) })
	mux.HandleFunc("OPTIONS /api/destinations/{id}", func(w http.ResponseWriter, r *http.Request) { enableCORS(w, r) })

	port := "8081"
	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)

	fmt.Println("APLIKASI PARIWISATA BACKEND (GOLANG) RUNNING")
	fmt.Printf("Server aktif di: http://localhost:%s\n", port)
	fmt.Println("API Endpoints:")
	fmt.Println("   - [GET]    /api/destinations")
	fmt.Println("   - [GET]    /api/destinations/{id}")
	fmt.Println("   - [POST]   /api/destinations")
	fmt.Println("   - [PUT]    /api/destinations/{id}")
	fmt.Println("   - [DELETE] /api/destinations/{id}")
	fmt.Println("   - [GET]    /api/destinations/sequential-search?q=keyword")
	fmt.Println("   - [GET]    /api/destinations/sequential-search?id=dest-1")
	fmt.Println("   - [GET]    /api/destinations/binary-search?id=dest-1")
	fmt.Println("   - [GET]    /api/destinations/selection-sort-cost?order=asc|desc")
	fmt.Println("   - [GET]    /api/destinations/insertion-sort-distance?order=asc|desc")
	fmt.Println("   - [GET]    /api/destinations/insertion-sort-facilities?order=desc|asc")

	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Fatal error: Gagal menjalankan server HTTP: %v", err)
	}
}
