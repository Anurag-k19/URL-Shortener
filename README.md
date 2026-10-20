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

Users can also sign in with Google. Signed-in users' links never expire; guest links expire after 2 minutes.

The backend validates the URL before saving it and rejects invalid or unsafe URLs.

## Main Features

- URL shortening and redirection
- Google sign-in (OAuth 2.0 / OpenID Connect) with sessions stored in Redis
- Validation of submitted URLs
- Monthly partitioned link storage for analysis
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
│   ├── auth/          # Google sign-in, sessions, middleware
│   ├── handlers/
│   │   ├── shorten.go
│   │   └── redirect.go
│   ├── migrations/    # SQL schema, run in order
│   ├── storage/
│   │   ├── db.go
│   │   ├── partitions.go
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
2. The backend checks the `session` cookie: signed-in user or guest
3. The backend validates the URL
4. A unique short ID is generated
5. The mapping is stored in PostgreSQL
6. The backend returns a short URL such as:

```text
http://localhost:8000/ABC123
```

### 2. Redirecting the short URL

When the user opens the short URL:

1. The browser requests `/ABC123`
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
- a "Sign in with Google" button, and the signed-in user's name with a sign-out button
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

Then create the tables by running the migrations:

```bash
cd Backend
psql -U postgres -d url_shortener -f migrations/001_create_short_urls.sql
psql -U postgres -d url_shortener -f migrations/003_create_users.sql
```

`002_migrate_urls_data.sql` is only for databases that still have the old `urls` table. See `Backend/README.md`.

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

GOOGLE_CLIENT_ID=your_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8000/auth/google/callback
FRONTEND_URL=http://localhost:5500
```

To get the Google values, see "Google sign-in setup" in `Backend/README.md`.

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

Use `localhost`, not `127.0.0.1`, or the browser won't send the sign-in cookie to the backend.

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
  "short_url": "http://localhost:8000/ABC123"
}
```

### GET /ABC123

Used to redirect the browser to the original URL. `/redirect?code=ABC123` also works.

### Authentication

- `GET /auth/google/login`: start Google sign-in
- `GET /auth/google/callback`: Google redirects here after sign-in
- `GET /auth/me`: the signed-in user, or `401`
- `POST /auth/logout`: sign out

See `Backend/README.md` for details.

## Notes

- Guest-created links expire after 2 minutes.
- Signed-in users' links do not have an expiry.
- The app uses Redis as a cache layer and PostgreSQL as the source of truth.
- CORS is enabled for local development so the frontend can communicate with the backend in the browser.
