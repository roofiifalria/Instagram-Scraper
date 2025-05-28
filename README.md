-----

# Instagram Hashtag Scraper (Go & Docker)

A real-time Instagram post scraper application built with Go, designed to efficiently extract post data based on specified hashtags and export it into a clean CSV format. This project is containerized with Docker for easy deployment and consistent execution across environments.

## Features

  * **Hashtag-Based Scraping**: Fetches Instagram post data for a given hashtag.
  * **Dynamic Date Filtering**: Extracts posts from a dynamically calculated period (e.g., last 30 days from now) or a user-defined timestamp.
  * **CSV Output**: Exports extracted data into a structured CSV file, ideal for data analysis.
  * **JSON Output (Raw)**: Also saves the raw JSON response from Instagram for detailed inspection and debugging.
  * **Containerized with Docker**: Ensures consistent execution by packaging the application and its dependencies.
  * **HTTP API Endpoint**: Provides a simple `/posts` endpoint to trigger the scraping process via HTTP requests.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

  * **Go (Golang)**: Version 1.20 or higher. [Download Go](https://go.dev/dl/)
  * **Git**: For cloning the repository. [Download Git](https://git-scm.com/downloads)
  * **Docker Desktop**: For building and running the Docker image. [Download Docker Desktop](https://www.docker.com/products/docker-desktop/)

## Project Structure

```
instagram-scraper/
├── go.mod
├── go.sum
├── Dockerfile
├── main.go               # Main HTTP server and API endpoint logic
├── posts/
│   └── posts.go          # Handles HTTP requests to Instagram API and saves raw JSON
└── split/
    └── split.go          # Processes raw JSON, filters data, and generates CSV/JSON output
└── output/               # Directory for scraped output files (created manually or by volume mount)
```

## Setup and Running

Follow these steps to get your Instagram Scraper up and running:

### 1\. Clone the Repository

Open your terminal or PowerShell and clone the project:

```bash
git clone https://github.com/Braquetes/instagram-scraper.git
cd instagram-scraper
```

### 2\. Verify Go Modules

Navigate into the cloned directory and ensure Go modules are tidy.

```bash
go mod tidy
```

### 3\. Create Output Directory

Create an `output` directory in your project root to store the scraped data. This folder will be mapped to the Docker container.

```bash
mkdir output
```

### 4\. Build the Docker Image

This step compiles your Go application and packages it into a Docker image.

```bash
docker build -t instagram-scraper-go .
```

You should see a successful build message at the end.

### 5\. Obtain Instagram Headers (Crucial Step\!)

The scraper relies on specific HTTP headers from a logged-in Instagram browser session to bypass anti-bot measures. These headers are dynamic and expire. **You MUST obtain the latest headers yourself.**

1.  Open your browser (Chrome/Firefox) and navigate to [https://www.instagram.com/](https://www.instagram.com/).
2.  **Log in** to your Instagram account.
3.  Open **Developer Tools** (press `F12`).
4.  Go to the **"Network"** tab.
5.  Click the **"Clear"** button (circle with a slash) to clear existing requests.
6.  In your browser's address bar (not Developer Tools), navigate to a hashtag page: `https://www.instagram.com/explore/tags/your_hashtag_here/` (e.g., `https://www.instagram.com/explore/tags/surabaya/`).
7.  Once the page loads, go back to the "Network" tab.
8.  In the filter box, type `graphql` (or `web_info`).
9.  Look for a **`POST`** request with a URL like `https://www.instagram.com/api/graphql/` (or `GET` to `api/v1/tags/web_info/` if it appears).
10. **Click on the name of that request**.
11. In the details panel that opens, click the **"Headers"** tab.
12. Scroll down to the **"Request Headers"** section.
13. **Copy the entire string/value for:**
      * `cookie:`
      * `x-asbd-id:`
      * `x-csrftoken:`
      * `x-ig-app-id:`
      * `x-ig-www-claim:` (If this header is present, copy its value. If not, you can leave it empty in the `docker run` command).
      * `user-agent:` (Copy the full User-Agent string).

### 6\. Run the Docker Container

Now, run the Docker container, mapping the output directory and providing the necessary Instagram headers as environment variables.

**Replace the placeholder values (`<YOUR_COPIED_VALUE>`) with the actual headers you obtained in the previous step.**

```bash
docker run -p 8000:8000 \
-v /mnt/c/Users/roofi/OneDrive\ -\ Institut\ Teknologi\ Sepuluh\ Nopember/magnag/instagram-scraper/output:/app/output \
-e COOKIE="<YOUR_COPIED_COOKIE_VALUE_HERE>" \
-e X_ASBD_ID="<YOUR_COPIED_X_ASBD_ID_HERE>" \
-e X_CSRFTOKEN="<YOUR_COPIED_X_CSRFTOKEN_HERE>" \
-e X_IG_APP_ID="<YOUR_COPIED_X_IG_APP_ID_HERE>" \
-e X_IG_WWW_CLAIM="<YOUR_COPIED_X_IG_WWW_CLAIM_HERE>" \
-e USER_AGENT="<YOUR_COPIED_USER_AGENT_HERE>" \
instagram-scraper-go
```

  * **`docker run`**: Starts a new container.
  * **`-p 8000:8000`**: Maps port 8000 on your host machine to port 8000 inside the container.
  * **`-v ...:/app/output`**: Mounts your local `output` directory to `/app/output` inside the container, so output files are saved on your host. **Double-check the host path for your system (e.g., `C:\Users\...\output` for Windows CMD/PowerShell, or `/mnt/c/.../output` for WSL/Bash).**
  * **`-e VARIABLE="value"`**: Sets environment variables. **Ensure all header values are enclosed in double quotes (`"`).**

### 7\. Test the API Endpoint

Once the container is running (you'll see "Server started on :8000" in your terminal), you can test the API.

Open your web browser or use `curl`:

```bash
# To get posts for 'surabaya' hashtag, default to last 30 days
http://localhost:8000/posts?hashtag=surabaya

# To get posts for 'kulonprogo' hashtag from a specific Unix timestamp (e.g., Jan 1, 2024 UTC)
http://localhost:8000/posts?hashtag=kulonprogo&limit=1704067200

# To get all posts for 'surabaya' hashtag (disables time filter)
http://localhost:8000/posts?hashtag=surabaya&limit=0
```

### 8\. Check the Output Files

After a successful API request, navigate to your local `output` folder (`instagram-scraper/output`). You should find:

  * `posts_YOURHASHTAG.json`: The raw JSON response from Instagram.
  * `extracted_posts_YOURHASHTAG.json`: The filtered and extracted data in JSON format.
  * `extracted_posts_YOURHASHTAG.csv`: The filtered and extracted data in CSV format.

## Troubleshooting

  * **`Extracted 0 posts.` or Empty CSV/JSON Output (despite `posts_YOURHASHTAG.json` being large):**

      * **Reason:** The `TopLevelResponse` struct in `split/split.go` does not precisely match the actual JSON structure of the `posts_YOURHASHTAG.json` file. Or, your `limit` timestamp is too recent for the available posts.
      * **Solution:**
        1.  **Open `posts_YOURHASHTAG.json`** (from your `output` folder) in a text editor.
        2.  **Copy its entire content.**
        3.  Go to [**https://json-to-go.appspot.com/**](https://www.google.com/search?q=https://json-to-go.appspot.com/).
        4.  Paste the JSON into the left panel. Copy the generated Go struct from the right panel.
        5.  **Replace the entire `type TopLevelResponse struct {...}` definition in `split/split.go`** with the newly generated struct.
        6.  Adjust the name of the main struct (e.g., `Root` or `AutoGenerated`) to `TopLevelResponse`.
        7.  You might need to adjust the extraction logic in the `Split` function (e.g., `edge.Node.Media.Caption.User.Username`) to match the new struct path.
        8.  **Rebuild the Docker image** (`docker build -t instagram-scraper-go .`) and **rerun the container**. Test with `limit=0` first.

  * **`Error reading input file 'posts_YOURHASHTAG.json': no such file or directory`:**

      * **Reason:** The `posts.Posts` function failed to create the file, or `split.Split` is looking in the wrong place.
      * **Solution:**
        1.  Check the Docker container logs for errors during the `posts.Posts` phase (e.g., HTTP status codes like 400, 403, 429). This often indicates invalid Instagram headers.
        2.  Ensure your Docker volume mount path (`-v /path/on/host:/app/output`) is correct for your OS.
        3.  Confirm `posts/posts.go` saves to `/app/output/` and `split/split.go` reads from `/app/output/`.

  * **`Error saving credentials: error storing credentials` during `docker login`:**

      * **Reason:** Docker's credential helper is misconfigured or inaccessible, especially in WSL.
      * **Solution:**
        1.  Ensure Docker Desktop is running and healthy on Windows.
        2.  Edit `~/.docker/config.json` in your WSL terminal and ensure it contains: `{"credsStore": "desktop"}`.
        3.  Try logging in again, preferably using a [Personal Access Token (PAT)](https://www.google.com/search?q=https://app.docker.com/settings/security/access-tokens) from Docker Hub as your password.

-----
