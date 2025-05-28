package split

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Post merepresentasikan struktur data postingan yang ingin kita ekstrak
type Post struct {
	OwnerUsername string `json:"owner_username,omitempty"` // Akun yang posting
	Text          string `json:"text"`                     // Konten (caption)
	Comments      int    `json:"comments,omitempty"`       // Jumlah Komentar
	PostURL       string `json:"post_url,omitempty"`       // URL Langsung ke Postingan
	// CreatedAt dihapus dari sini sesuai permintaan untuk output, tapi tetap digunakan untuk filter.
	// Field lain seperti MediaType, Likes, Views, Share, ImageURL tidak diminta untuk output CSV akhir
	// jadi tidak disertakan di sini, meskipun datanya bisa diekstrak oleh TopLevelResponse.
}

// Data adalah struktur untuk menampung koleksi Post yang diekstrak
type Data struct {
	Posts []Post `json:"posts"`
}

// TopLevelResponse mewakili struktur respons JSON keseluruhan dari posts_surabaya.json Anda.
// Ini adalah hasil parse dari JSON yang Anda berikan.
type TopLevelResponse struct {
	Count int `json:"count"`
	Data  struct {
		AllowFollowing         bool        `json:"allow_following"`
		AllowMutingStory       bool        `json:"allow_muting_story"`
		ContentAdvisory        interface{} `json:"content_advisory"` // Bisa null
		FollowButtonText       interface{} `json:"follow_button_text"` // Bisa null
		FollowStatus           int         `json:"follow_status"`
		Following              int         `json:"following"`
		FormattedMediaCount    string      `json:"formatted_media_count"`
		HideUseHashtagButton   bool        `json:"hide_use_hashtag_button"`
		ID                     string      `json:"id"`
		IsTrending             bool        `json:"is_trending"`
		MediaCount             int         `json:"media_count"`
		Name                   string      `json:"name"`
		ProfilePicURL          string      `json:"profile_pic_url"`
		Recent                 struct {
			Sections []interface{} `json:"sections"` // "sections" kosong di sample, jadi interface{}
		} `json:"recent"`
		ShowFollowDropDown     bool        `json:"show_follow_drop_down"`
		SocialContext          string      `json:"social_context"`
		SocialContextProfileLinks []interface{} `json:"social_context_profile_links"`
		Subtitle               string      `json:"subtitle"`
		Top                    struct {
			MoreAvailable bool   `json:"more_available"`
			NextMaxID     string `json:"next_max_id"`
			NextMediaIDs  []string `json:"next_media_ids"`
			NextPage      int    `json:"next_page"`
			Sections      []struct {
				ExploreItemInfo struct {
					AspectRatio   float64 `json:"aspect_ratio"`
					Autoplay      bool    `json:"autoplay"`
					NumColumns    int     `json:"num_columns"`
					TotalNumColumns int     `json:"total_num_columns"`
				} `json:"explore_item_info"`
				FeedType    string `json:"feed_type"`
				LayoutContent struct {
					// Ini bisa "fill_items" untuk clips atau "medias" untuk grid media
					FillItems []struct { // Untuk type "clips" (carousel)
						Media struct {
							Caption struct {
								CreatedAt   int64 `json:"created_at"` // Waktu posting
								Text        string `json:"text"`       // Konten
								User        struct {
									Username string `json:"username"` // Akun yang posting
								} `json:"user"`
							} `json:"caption"`
							Code         string `json:"code"`          // Untuk URL postingan
							CommentCount int    `json:"comment_count"` // Jumlah komentar
							// Field lain yang ada di sample JSON Anda
							IsVideo bool `json:"is_video"` // Untuk menentukan tipe media
							LikeCount int `json:"like_count"`
							PlayCount int `json:"play_count"` // Untuk views video (sering ada untuk clips)
							ImageVersions2 struct {
								Candidates []struct {
									URL string `json:"url"`
								} `json:"candidates"`
							} `json:"image_versions2"`
							VideoDashManifest string `json:"video_dash_manifest"`
							VideoDuration float64 `json:"video_duration"`
							VideoVersions []struct {
								URL string `json:"url"`
							} `json:"video_versions"`
						} `json:"media"`
					} `json:"fill_items,omitempty"` // Opsional jika layout bukan "clips"
					Medias []struct { // Untuk type "media" (grid)
						Media struct {
							Caption struct {
								CreatedAt   int64 `json:"created_at"` // Waktu posting
								Text        string `json:"text"`       // Konten
								User        struct {
									Username string `json:"username"` // Akun yang posting
								} `json:"user"`
							} `json:"caption"`
							Code         string `json:"code"`          // Untuk URL postingan
							CommentCount int    `json:"comment_count"` // Jumlah komentar
							// Field lain yang ada di sample JSON Anda
							IsVideo bool `json:"is_video"`
							LikeCount int `json:"like_count"`
							PlayCount int `json:"play_count"` // Untuk views video (sering ada untuk feed)
							ImageVersions2 struct {
								Candidates []struct {
									URL string `json:"url"`
								} `json:"candidates"`
							} `json:"image_versions2"`
							VideoDashManifest string `json:"video_dash_manifest"`
							VideoDuration float64 `json:"video_duration"`
							VideoVersions []struct {
								URL string `json:"url"`
							} `json:"video_versions"`
						} `json:"media"`
					} `json:"medias,omitempty"` // Opsional jika layout bukan "media"
				} `json:"layout_content"`
				LayoutType string `json:"layout_type"`
			} `json:"sections"`
		} `json:"top"` // Jalur data sebenarnya ada di dalam "top.sections"
		WarningMessage interface{} `json:"warning_message"` // Bisa null
	} `json:"data"`
	Status string `json:"status"`
}

func Split(inputFile string, outputFileBase string, limitTimestampStr string) {
	log.Printf("Starting data splitting and filtering from '%s'...", inputFile)

	bytes, err := os.ReadFile(inputFile)
	if err != nil {
		log.Printf("Error reading input file '%s': %v\n", inputFile, err)
		return
	}
	log.Printf("Input file '%s' read successfully. Size: %d bytes.", inputFile, len(bytes))

	limitTime, err := strconv.ParseInt(limitTimestampStr, 10, 64)
	if err != nil {
		log.Fatalf("FATAL: Error converting timestamp string '%s' to integer: %v", limitTimestampStr, err)
	}
	log.Printf("Filtering posts created after or at: %s (Unix: %d)", time.Unix(limitTime, 0).Format("2006-01-02 15:04:05 MST"), limitTime)

	var extractedData Data
	extractedData.Posts = make([]Post, 0)

	var instaResp TopLevelResponse
	if err := json.Unmarshal(bytes, &instaResp); err == nil {
		// Log ini sekarang akan mencerminkan total count yang ditemukan di JSON top-level
		log.Printf("Successfully unmarshalled JSON to TopLevelResponse. Top more_available status: %t", instaResp.Data.Top.MoreAvailable)// count di level top.sections.layout_content.medias atau fill_items

		// ITERASI MELALUI SEMUA SECTIONS DAN MEDIAS DI DALAMNYA
		filteredCount := 0
		for _, section := range instaResp.Data.Top.Sections {
			if len(section.LayoutContent.FillItems) > 0 { // Untuk tipe "clips" atau carousel
				for _, item := range section.LayoutContent.FillItems {
					media := item.Media
					postCreatedAt := media.Caption.CreatedAt

					if postCreatedAt >= limitTime {
						postURL := fmt.Sprintf("https://www.instagram.com/p/%s/", media.Code)
						extractedData.Posts = append(extractedData.Posts, Post{
							OwnerUsername: media.Caption.User.Username,
							Text:          media.Caption.Text,
							Comments:      media.CommentCount,
							PostURL:       postURL,
						})
						filteredCount++
					}
				}
			} else if len(section.LayoutContent.Medias) > 0 { // Untuk tipe "media_grid"
				for _, item := range section.LayoutContent.Medias {
					media := item.Media
					postCreatedAt := media.Caption.CreatedAt // Atau media.TakenAt jika ada

					if postCreatedAt >= limitTime {
						postURL := fmt.Sprintf("https://www.instagram.com/p/%s/", media.Code)
						extractedData.Posts = append(extractedData.Posts, Post{
							OwnerUsername: media.Caption.User.Username,
							Text:          media.Caption.Text,
							Comments:      media.CommentCount,
							PostURL:       postURL,
						})
						filteredCount++
					}
				}
			}
		}
		log.Printf("Finished filtering. %d posts extracted based on timestamp filter.", filteredCount)
	} else {
		log.Printf("WARNING: Direct Unmarshal to TopLevelResponse failed (%v). This indicates an outdated TopLevelResponse struct. Please update it based on your actual posts.json. Proceeding with recursive extraction as fallback.", err)
		var rawData interface{}
		if err := json.Unmarshal(bytes, &rawData); err != nil {
			log.Fatalf("FATAL: Error parsing JSON for recursive extraction: %v", err)
		}
		// Menggunakan fungsi rekursif dengan parameter yang sesuai dengan Post struct yang disederhanakan
		extractPostsRecursively(rawData, &extractedData, limitTime) // limitTime tetap dipassing untuk filter rekursif
		log.Printf("Recursive extraction finished. %d posts extracted.", len(extractedData.Posts))
	}

	log.Printf("Total %d posts extracted for output.", len(extractedData.Posts))

	// --- Output JSON (Opsional, bisa dihapus jika hanya ingin CSV) ---
	jsonOutputFileName := fmt.Sprintf("%s.json", outputFileBase)
	outputBytes, err := json.MarshalIndent(extractedData, "", "    ")
	if err != nil {
		log.Fatalf("Error marshalling extracted data to JSON: %v", err)
	}
	if err := os.WriteFile(jsonOutputFileName, outputBytes, 0644); err != nil {
		log.Fatalf("Error writing extracted data to JSON file '%s': %v", jsonOutputFileName, err)
	}
	log.Printf("Raw data saved to '%s'", jsonOutputFileName)

	// --- Bagian Output CSV ---
	csvOutputFileName := fmt.Sprintf("%s.csv", outputFileBase)
	csvFile, err := os.Create(csvOutputFileName)
	if err != nil {
		log.Fatalf("Error creating CSV file '%s': %v", csvOutputFileName, err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Tulis header CSV sesuai format baru (username, text, comment_count, url)
	header := []string{"akun yang posting", "konten", "jumlah komentar", "url postingan"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Error writing CSV header: %v", err)
	}
	log.Println("CSV header written.")

	// Tulis data posts ke CSV
	for i, post := range extractedData.Posts {
		record := []string{
			post.OwnerUsername,
			post.Text,
			strconv.Itoa(post.Comments),
			post.PostURL,
		}
		if err := writer.Write(record); err != nil {
			log.Fatalf("Error writing CSV record for post %d: %v", i+1, err)
		}
	}
	log.Printf("All %d posts written to CSV. Data also saved to '%s'", len(extractedData.Posts), csvOutputFileName)
}

// Fungsi rekursif untuk mencari dan mengekstrak posts jika struktur JSON tidak langsung cocok dengan TopLevelResponse.
func extractPostsRecursively(value interface{}, data *Data, limitTime int64) {
	switch v := value.(type) {
	case map[string]interface{}:
		// Coba ekstrak informasi post dari objek "media"
		if mediaMap, ok := v["media"].(map[string]interface{}); ok {
			if captionMap, ok := mediaMap["caption"].(map[string]interface{}); ok {
				if text, textOK := captionMap["text"].(string); textOK {
					if createdAtRaw, createdAtOK := captionMap["created_at"].(float64); createdAtOK {
						createdAt := int64(createdAtRaw)
						if createdAt >= limitTime { // Filter waktu tetap berjalan di sini
							// Default values
							ownerUsername := ""
							comments := 0
							postURL := ""

							if userMap, userOK := captionMap["user"].(map[string]interface{}); userOK {
								if un, unOK := userMap["username"].(string); unOK {
									ownerUsername = un
								}
							}
							if commentCountRaw, commentCountOK := mediaMap["comment_count"].(float64); commentCountOK {
								comments = int(commentCountRaw)
							}
							if code, codeOK := mediaMap["code"].(string); codeOK {
								postURL = fmt.Sprintf("https://www.instagram.com/p/%s/", code)
							}

							data.Posts = append(data.Posts, Post{
								OwnerUsername: ownerUsername,
								Text:          text,
								Comments:      comments,
								PostURL:       postURL,
								// CreatedAt tidak lagi diisi ke Post struct karena sudah dihapus
							})
						}
					}
				}
			}
		}
		// Terus eksplorasi nilai-nilai di dalam map secara rekursif.
		for _, item := range v {
			extractPostsRecursively(item, data, limitTime)
		}
	case []interface{}:
		// Jika ini adalah array, iterasi melalui elemen-elemennya dan panggil rekursif untuk setiap item.
		for _, item := range v {
			extractPostsRecursively(item, data, limitTime)
		}
	}
}
