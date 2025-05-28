package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv" // Pastikan ini diimpor
	"time"    // Pastikan ini diimpor

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"instagram-scraper/posts"
	"instagram-scraper/split"
)

// getPostsHandler adalah handler HTTP untuk endpoint /posts.
// Ia akan menerima request GET, memproses hashtag, melakukan scraping,
// memfilter data, dan mengembalikan hasilnya dalam format JSON.
func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Log setiap request yang masuk untuk keperluan debugging.
	log.Printf("Received GET request for /posts from %s", r.RemoteAddr)

	// Set header Content-Type untuk respons sebagai application/json.
	w.Header().Set("Content-Type", "application/json")

	// Ambil nilai hashtag dari query parameter URL (misalnya /posts?hashtag=KITTEN).
	hashtag := r.URL.Query().Get("hashtag")
	if hashtag == "" {
		// Jika parameter hashtag tidak disediakan, kembalikan error Bad Request (400).
		http.Error(w, `{"error": "Query parameter 'hashtag' is required."}`, http.StatusBadRequest)
		log.Println("Error: 'hashtag' query parameter is missing.")
		return // Hentikan eksekusi handler.
	}
	fmt.Printf("Processing hashtag: %s\n", hashtag)

	// --- LOGIKA UNTUK FILTER TANGGAL DINAMIS ---
	// Ambil nilai timestamp batas awal dari query parameter 'limit'.
	// Jika tidak disediakan, hitung secara dinamis (misalnya, 30 hari yang lalu dari sekarang).
	limitTimestampStr := r.URL.Query().Get("limit") // Coba ambil dari URL parameter

	if limitTimestampStr == "" {
		// Default: Hitung timestamp 30 hari yang lalu dari waktu sekarang (UTC)
		// Anda bisa mengubah angka 30 di sini untuk periode yang berbeda (misal 7 hari, 90 hari)
		defaultDaysAgo := 30
		time30DaysAgo := time.Now().UTC().AddDate(0, 0, -defaultDaysAgo) // Menghitung tanggal N hari yang lalu
		limitTimestampStr = strconv.FormatInt(time30DaysAgo.Unix(), 10) // Konversi ke string Unix timestamp
		log.Printf("No 'limit' timestamp provided. Defaulting to %d days ago: %s (Unix: %s)", defaultDaysAgo, time30DaysAgo.Format("2006-01-02 15:04:05 UTC"), limitTimestampStr)
	} else {
		// Jika parameter 'limit' disediakan, log nilai yang digunakan.
		parsedTime, err := strconv.ParseInt(limitTimestampStr, 10, 64)
		if err == nil {
			log.Printf("Using 'limit' timestamp from URL: %s (Parsed: %s)", limitTimestampStr, time.Unix(parsedTime, 0).Format("2006-01-02 15:04:05 UTC"))
		} else {
			log.Printf("Warning: Invalid 'limit' timestamp in URL: '%s'. Proceeding with provided string.", limitTimestampStr)
		}
	}

	// --- Langkah 1: Melakukan Scraping Data dari Instagram ---
	// Panggil fungsi posts.Posts yang akan mengambil data dari Instagram API.
	// Fungsi ini akan menyimpan hasil mentah ke file bernama 'posts_NAMAHASHTAG.json'
	posts.Posts(hashtag)

	// Tentukan nama file input dan output untuk langkah pemrosesan data (split).
	// inputFileName harus lengkap dengan path dan ekstensi.
	inputFileName := fmt.Sprintf("/app/output/posts_%s.json", hashtag)
	// outputBaseFileName adalah nama dasar tanpa ekstensi, karena split.Split akan membuat .json dan .csv
	outputBaseFileName := fmt.Sprintf("/app/output/extracted_posts_%s", hashtag)

	// --- Langkah 2: Memfilter dan Memisahkan Data ---
	// Panggil fungsi split.Split untuk membaca file input, memfilter postingan
	// berdasarkan timestamp batas, dan menulis hasilnya ke file output (JSON dan CSV).
	split.Split(inputFileName, outputBaseFileName, limitTimestampStr)

	// --- Langkah 3: Membaca Hasil Akhir untuk Dikembalikan via HTTP ---
	// Baca file JSON yang sudah difilter dan dipisahkan oleh split.Split.
	// Kita perlu tambahkan ekstensi .json secara manual di sini.
	data, err := os.ReadFile(fmt.Sprintf("%s.json", outputBaseFileName))
	if err != nil {
		// Jika file output tidak dapat dibaca (mungkin karena tidak ada atau error saat pemrosesan),
		// kembalikan error Internal Server Error (500).
		log.Printf("Error reading final output JSON file '%s.json': %v\n", outputBaseFileName, err)
		http.Error(w, fmt.Sprintf(`{"error": "Error reading extracted data file: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// --- Langkah 4: Mengirim Hasil ke Klien HTTP ---
	// Tulis konten file JSON ke response writer HTTP.
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response data: %v\n", err)
	}
	log.Printf("Successfully processed hashtag '%s' and sent response.", hashtag)
}

// main adalah fungsi entry point aplikasi server Go.
func main() {
	// Untuk logging, agar ada timestamp di setiap log
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	router := mux.NewRouter()
	router.HandleFunc("/posts", getPostsHandler).Methods("GET")

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),                               // Izinkan semua origin. Hati-hati di production, batasi ke domain yang spesifik.
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // Metode HTTP yang diizinkan.
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),   // Headers HTTP yang diizinkan.
	)(router)

	fmt.Println("Server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", corsHandler))
}