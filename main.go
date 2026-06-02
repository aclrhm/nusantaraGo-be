package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 1. Inisialisasi basis data lokal JSON
	err := InitDB()
	if err != nil {
		log.Fatalf("Fatal error: Gagal menginisialisasi basis data: %v", err)
	}

	// 2. Buat multiplexer router baru
	mux := http.NewServeMux()

	// 3. Registrasi route REST API (Menggunakan sintaks routing baru Go 1.22+)
	mux.HandleFunc("GET /api/destinations", getDestinationsHandler)
	mux.HandleFunc("GET /api/destinations/{id}", getDestinationByIDHandler)
	mux.HandleFunc("POST /api/destinations", createDestinationHandler)
	mux.HandleFunc("PUT /api/destinations/{id}", updateDestinationHandler)
	mux.HandleFunc("DELETE /api/destinations/{id}", deleteDestinationHandler)

	// Route OPTIONS global sebagai penunjang preflight CORS jika diperlukan
	mux.HandleFunc("OPTIONS /api/destinations", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
	})
	mux.HandleFunc("OPTIONS /api/destinations/{id}", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
	})

	// 4. Konfigurasi alamat server
	port := "8081"
	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)

	fmt.Println("==================================================")
	fmt.Println("APLIKASI PARIWISATA BACKEND (GOLANG) RUNNING")
	fmt.Println("==================================================")
	fmt.Printf("Server aktif di: http://localhost:%s\n", port)
	fmt.Println("API Endpoints:")
	fmt.Println("   - [GET]    /api/destinations (Pencarian, Filter, Sorting)")
	fmt.Println("   - [GET]    /api/destinations/{id} (Detail Destinasi)")
	fmt.Println("   - [POST]   /api/destinations (Admin: Tambah Destinasi)")
	fmt.Println("   - [PUT]    /api/destinations/{id} (Admin: Edit Destinasi)")
	fmt.Println("   - [DELETE] /api/destinations/{id} (Admin: Hapus Destinasi)")
	fmt.Println("==================================================")

	// 5. Jalankan server
	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Fatal error: Gagal menjalankan server HTTP: %v", err)
	}
}
