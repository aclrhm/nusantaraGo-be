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

	mux.HandleFunc("GET /api/destinations", getDestinationsHandler)
	mux.HandleFunc("GET /api/destinations/{id}", getDestinationByIDHandler)
	mux.HandleFunc("POST /api/destinations", createDestinationHandler)
	mux.HandleFunc("PUT /api/destinations/{id}", updateDestinationHandler)
	mux.HandleFunc("DELETE /api/destinations/{id}", deleteDestinationHandler)
	mux.HandleFunc("GET /api/destinations/sequential-search", sequentialSearchHandler)
	mux.HandleFunc("GET /api/destinations/binary-search", binarySearchHandler)
	mux.HandleFunc("GET /api/destinations/selection-sort-cost", selectionSortCostHandler)
	mux.HandleFunc("GET /api/destinations/insertion-sort-distance", insertionSortDistanceHandler)
	mux.HandleFunc("GET /api/destinations/insertion-sort-facilities", insertionSortFacilitiesHandler)

	mux.HandleFunc("OPTIONS /api/destinations", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
	})
	mux.HandleFunc("OPTIONS /api/destinations/{id}", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
	})
	

	port := "8081"
	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)


	fmt.Println("APLIKASI PARIWISATA BACKEND (GOLANG) RUNNING")

	fmt.Printf("Server aktif di: http://localhost:%s\n", port)
	fmt.Println("API Endpoints:")
	fmt.Println("   - [GET]    /api/destinations (Semua Data Destinasi)")
	fmt.Println("   - [GET]    /api/destinations/{id} (Detail Destinasi)")
	fmt.Println("   - [POST]   /api/destinations (Admin: Tambah Destinasi)")
	fmt.Println("   - [PUT]    /api/destinations/{id} (Admin: Edit Destinasi)")
	fmt.Println("   - [DELETE] /api/destinations/{id} (Admin: Hapus Destinasi)")
	fmt.Println("   - [GET]    /api/destinations/sequential-search?q=keyword (Sequential Search - Kata Kunci)")
	fmt.Println("   - [GET]    /api/destinations/sequential-search?id=dest-1 (Sequential Search - ID)")
	fmt.Println("   - [GET]    /api/destinations/binary-search?id=dest-1 (Binary Search - ID)")
	fmt.Println("   - [GET]    /api/destinations/selection-sort-cost?order=asc|desc (Selection Sort - Biaya)")
	fmt.Println("   - [GET]    /api/destinations/insertion-sort-distance?order=asc|desc (Insertion Sort - Jarak)")
	fmt.Println("   - [GET]    /api/destinations/insertion-sort-facilities?order=desc|asc (Insertion Sort - Fasilitas)")


	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Fatal error: Gagal menjalankan server HTTP: %v", err)
	}
}
