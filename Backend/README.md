# URL Shortener

A lightweight Go-based URL shortening service that converts long links into short, shareable URLs and redirects them back to the original destination.

## What the API is for

The API is used to:

- Create a short URL from a long URL
- Redirect a short URL back to the original URL
- Keep track of guest and user-specific links
- Cache redirects in Redis for faster lookups

## How it works

1. A client sends a long URL to the shorten endpoint.
2. The server validates the URL and checks whether it is reachable.
3. A short ID is generated and stored in PostgreSQL.
4. The client receives a short URL that can be used later.
5. When the short URL is visited, the server redirects the user to the original URL.

## Project structure

```text
.
├── main.go
├── handlers/
│   ├── shorten.go
│   └── redirect.go
├── storage/
│   ├── db.go
│   └── redis.go
├── utils/
│   ├── generator.go
│   └── validateURL.go
└── go.mod
```

## API endpoints

### 1) Create a short URL

Endpoint:

```http
POST /shorten
```

Use this endpoint when you want to shorten a long URL.

Headers:

- `Content-Type: application/json`
- `X-User-ID` (optional): if provided, the URL is associated with that user; otherwise it is treated as a guest link

Request body:

```json
{
  "url": "https://example.com/very/long/url"
}
```

Response:

```json
{
  "short_url": "http://localhost:8000/abc123"
}
```

Behavior:

- Guest links expire after 2 minutes.
- User-linked links do not expire.
- If the same user submits the same long URL again, the service returns the existing short URL instead of creating a duplicate.

### 2) Redirect to the original URL

Endpoint:

```http
GET /{shortId}
```

Use this endpoint when a user opens a shortened link.

Behavior:

- If the short ID exists in Redis, the redirect is served quickly.
- Otherwise, the server looks it up in PostgreSQL.
- If the URL is not found, the API returns `404 Not Found`.
- If the link has expired, it returns `410 Gone`.
- If the link is valid, it returns `302 Found` and redirects the browser.

## Example usage

### Create a short URL

```bash
curl -X POST http://localhost:8000/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.example.com"}'
```

### Create a short URL for a specific user

```bash
curl -X POST http://localhost:8000/shorten \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{"url":"https://www.example.com"}'
```

### Open a shortened URL

```bash
curl -i http://localhost:8000/abc123
```

## Setup and running

### Prerequisites

- Go installed
- PostgreSQL running
- Redis running on `localhost:6379`

### Environment variables

Create a `.env` file in the project root with values similar to:

```env
DB_USER=your_postgres_user
PASSWORD=your_postgres_password
DBNAME=url_shortener
SSLMODE=disable
DB_HOST=localhost
DB_PORT=5432
```

### Run the server

```bash
go run main.go
```

The server listens on `http://localhost:8000`.

## Storage details

- PostgreSQL stores the mapping between short IDs and original URLs.
- Redis is used as a temporary cache for redirects.
- Guest URLs are marked with an expiry time; user URLs are stored without an expiry.

## Notes

The service validates the URL before storing it and rejects invalid or unreachable URLs with a `400 Bad Request` response.
