package posts

import (
	"encoding/json"
	"fmt"
	"io"
	"log" // Pastikan ini diimpor
	"net/http"
	"os"
	"time"
)

// Posts mencoba mengambil data postingan Instagram berdasarkan hashtag
// dan menyimpannya ke file JSON.
func Posts(hashtag string) {
	url := "https://www.instagram.com/api/v1/tags/web_info/?tag_name=" + hashtag

	log.Printf("Attempting to fetch data for hashtag '%s' from Instagram API...", hashtag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating HTTP request for %s: %v\n", hashtag, err)
		return
	}

	cookie := os.Getenv("COOKIE")
	if cookie == "" {
		log.Println("WARNING: COOKIE environment variable is not set. Request might fail.")
	}
	req.Header.Set("cookie", cookie)

	if ua := os.Getenv("USER_AGENT"); ua != "" {
		req.Header.Set("User-Agent", ua)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36") // Fallback
	}

	if asbdID := os.Getenv("X_ASBD_ID"); asbdID != "" {
		req.Header.Set("X-ASBD_ID", asbdID)
	} else {
		log.Println("WARNING: X_ASBD_ID environment variable is not set.")
	}
	if csrfToken := os.Getenv("X_CSRFTOKEN"); csrfToken != "" {
		req.Header.Set("X-CSRFTOKEN", csrfToken)
	} else {
		log.Println("WARNING: X_CSRFTOKEN environment variable is not set.")
	}
	if igAppID := os.Getenv("X_IG_APP_ID"); igAppID != "" {
		req.Header.Set("X-IG_APP_ID", igAppID)
	} else {
		log.Println("WARNING: X_IG_APP_ID environment variable is not set.")
	}
	if igWWWClaim := os.Getenv("X_IG_WWW_CLAIM"); igWWWClaim != "" {
		req.Header.Set("X-IG_WWW_CLAIM", igWWWClaim)
	} else {
		log.Println("WARNING: X_IG_WWW_CLAIM environment variable is not set.")
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.instagram.com/explore/tags/"+hashtag+"/")
	req.Header.Set("Sec-Ch-Ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "Linux")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error performing HTTP request for %s: %v\n", hashtag, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		log.Printf("HTTP request for %s failed with status code %d: %s\n", hashtag, resp.StatusCode, string(errorBody))
		return
	}
	log.Printf("Successfully received HTTP response for hashtag '%s'. Status: %d", hashtag, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body for %s: %v\n", hashtag, err)
		return
	}
	log.Printf("Response body read. Size: %d bytes.", len(body))


	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error decoding JSON response for %s: %v\n", hashtag, err)
		log.Printf("Raw response body: %s\n", string(body))
		return
	}
	log.Println("JSON response successfully decoded.")


	fileName := fmt.Sprintf("/app/output/posts_%s.json", hashtag)
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Error creating JSON file %s: %v\n", fileName, err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(result); err != nil {
		log.Printf("Error writing to JSON file %s: %v\n", fileName, err)
		return
	}
	log.Printf("Raw data for hashtag '%s' saved to '%s'", hashtag, fileName)
}
