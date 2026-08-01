# URL Shortener

This project is a full-stack URL shortening application built with a Go backend and a framework-free frontend. It lets users shorten long URLs, store them securely in PostgreSQL, and redirect visitors to the original destination through a short link.

## Project Overview

The application is split into two main parts:

- Backend: a Go server that handles URL creation, validation, storage, and redirection
- Frontend: a lightweight static web app that provides the user interface for submitting URLs and viewing the generated short link

The app is designed to work locally with PostgreSQL for persistent storage and Redis for fast lookup and caching of redirects.

## What the app does

Users can:

- enter a long URL in the frontend
- submit it to the backend
- receive a short URL in return
- open that short URL in the browser
- get redirected to the original destination

The backend validates the URL before saving it and ensures that invalid or unreachable URLs are rejected.

## Main Features

- URL shortening and redirection
- Validation of submitted URLs
- PostgreSQL-backed persistence
- Redis caching for faster redirects
- Guest link expiry after 2 minutes
- CORS support for local frontend-to-backend communication
- Simple static frontend with no framework dependency

## Project Structure

```text
URL Shortner/
├── Backend/
│   ├── main.go
│   ├── go.mod
│   ├── handlers/
│   │   ├── shorten.go
│   │   └── redirect.go
│   ├── storage/
│   │   ├── db.go
│   │   └── redis.go
│   └── utils/
│       ├── generator.go
│       └── validateURL.go
└── Frontend/
    ├── index.html
    ├── css/
    └── js/
```

## Backend Flow

### 1. Shortening a URL

When a user submits a long URL from the frontend:

1. The request is sent to the backend endpoint `/shorten`
2. The backend validates the URL
3. It checks whether the URL is reachable
4. A unique short ID is generated
5. The mapping is stored in PostgreSQL
6. The backend returns a short URL such as:

```text
http://localhost:8000/redirect?code=ABC123
```

### 2. Redirecting the short URL

When the user opens the short URL:

1. The browser requests `/redirect?code=ABC123`
2. The backend checks the short code in the database
3. If the link is still valid, it redirects the browser to the original URL
4. If the link has expired, it returns an expired-link response
5. Redis is used to cache the redirect for quick repeated access

## Frontend Flow

The frontend is a static web app built with plain HTML, CSS, and JavaScript.

It provides:

- a text input for the long URL
- a submit button to create the short URL
- a result area to display the generated short link
- a simple user experience without needing a framework

The frontend communicates with the backend through the API layer in the JavaScript files under the Frontend/js folder.

## Technologies Used

- Go for the backend server and request handling
- PostgreSQL for persistent storage of URL mappings
- Redis for caching and faster redirect lookup
- HTML, CSS, and JavaScript for the frontend
- Docker (optional) for running Redis

## Setup Instructions

### 1. Install prerequisites

Make sure you have:

- Go installed
- PostgreSQL installed and running
- Redis installed or available via Docker
- Python 3 for serving the frontend locally

### 2. Create the PostgreSQL database

Run the following in PostgreSQL:

```sql
CREATE DATABASE url_shortener;
```

Then create the table:

```sql
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_id TEXT UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    user_id TEXT NOT NULL,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 3. Start Redis

If you are using Docker:

```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

### 4. Configure the backend environment

Create a `.env` file inside the Backend folder:

```env
DB_USER=postgres
PASSWORD=your_password
DBNAME=url_shortener
SSLMODE=disable
DB_HOST=localhost
DB_PORT=5432
```

### 5. Run the backend

```bash
cd Backend
go run .
```

The backend runs on:

```text
http://localhost:8000
```

### 6. Run the frontend

```bash
cd Frontend
python -m http.server 5500
```

Then open:

```text
http://localhost:5500
```

## API Endpoints

### POST /shorten

Used to shorten a long URL.

Example request:

```json
{
  "url": "https://example.com/very/long/url"
}
```

Example response:

```json
{
  "short_url": "http://localhost:8000/redirect?code=ABC123"
}
```

### GET /redirect?code=ABC123

Used to redirect the browser to the original URL.

## Notes

- Guest-created links expire after 2 minutes.
- Registered or user-specific links do not have an expiry.
- The app uses Redis as a cache layer and PostgreSQL as the source of truth.
- CORS is enabled for local development so the frontend can communicate with the backend in the browser.
