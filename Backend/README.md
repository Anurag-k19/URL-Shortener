# URL Shortener

A lightweight Go-based URL shortening service that converts long links into short, shareable URLs and redirects them back to the original destination.

## What the API is for

The API is used to:

- Create a short URL from a long URL
- Redirect a short URL back to the original URL
- Sign users in with Google, and keep track of guest and signed-in users' links
- Cache redirects in Redis for faster lookups

## How it works

1. A client sends a long URL to the shorten endpoint.
2. The server validates the URL (see [URL validation](#url-validation)).
3. A short ID is generated and stored in PostgreSQL.
4. The client receives a short URL that can be used later.
5. When the short URL is visited, the server redirects the user to the original URL.

## Project structure

```text
.
├── main.go
├── auth/
│   ├── google.go       # OAuth/OIDC setup
│   ├── login.go        # /auth/google/login
│   ├── callback.go     # /auth/google/callback
│   ├── session.go      # sessions in Redis + cookie
│   ├── middleware.go   # session cookie -> user ID in request context
│   ├── account.go      # /auth/me, /auth/logout
│   └── users.go        # users table queries
├── handlers/
│   ├── shorten.go
│   └── redirect.go
├── migrations/
│   ├── 001_create_short_urls.sql
│   ├── 002_migrate_urls_data.sql
│   └── 003_create_users.sql
├── storage/
│   ├── db.go
│   ├── partitions.go
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
- The `session` cookie (sent automatically by the browser after signing in). Without it, the request is treated as a guest.

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
- Signed-in users' links do not expire.
- If the same user submits the same long URL again and their earlier short URL hasn't expired, the service returns it instead of creating a duplicate.

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
- If the link has expired, it returns `410 Gone`. The row is kept in PostgreSQL for analysis.
- If the link is valid, it returns `302 Found` and redirects the browser.

### 3) Authentication (Google sign-in)

| Endpoint | Method | What it does |
|---|---|---|
| `/auth/google/login` | GET | Redirects to Google's sign-in page. Open it in the browser, not with `fetch`. |
| `/auth/google/callback` | GET | Google redirects here. Verifies the sign-in, saves the user, sets the `session` cookie and redirects to `FRONTEND_URL`. |
| `/auth/me` | GET | Returns `{"id", "email", "name"}` for the signed-in user, or `401`. |
| `/auth/logout` | POST | Deletes the session and clears the cookie. Returns `204`. |

How it works:

1. `/auth/google/login` stores a random `state` in a short-lived cookie and sends the browser to Google.
2. Google sends the browser back to `/auth/google/callback` with a `code`. The server checks `state` (CSRF protection), swaps the code for an ID token, verifies it, and upserts the user into `users` by their Google `sub`.
3. The server creates a session in Redis (`session:<random id>` → user ID, 7 days) and sets an HttpOnly `session` cookie.
4. On every request, `auth.Middleware` looks the cookie up in Redis and puts the user ID in the request context. Handlers read it with `auth.UserID(ctx)`. Requests without a valid session are guests.

The old `X-User-ID` header is no longer used: anyone could set it to any value.

## Example usage

### Create a short URL

```bash
curl -X POST http://localhost:8000/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.example.com"}'
```

### Create a short URL as a signed-in user

Sign in through the frontend, then the browser sends the `session` cookie automatically. From the terminal, copy the cookie value from DevTools (Application → Cookies):

```bash
curl -X POST http://localhost:8000/shorten \
  -H "Content-Type: application/json" \
  -b "session=<your session id>" \
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

Create a `.env` file in the `Backend/` folder with values similar to:

```env
DB_USER=your_postgres_user
PASSWORD=your_postgres_password
DBNAME=url_shortener
SSLMODE=disable
DB_HOST=localhost
DB_PORT=5432
BASE_URL=http://localhost:8000   # optional: the domain used in returned short links

GOOGLE_CLIENT_ID=your_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8000/auth/google/callback
FRONTEND_URL=http://localhost:5500   # where to send the browser after signing in
```

### Google sign-in setup

1. In [Google Cloud Console](https://console.cloud.google.com), create a project and open **Google Auth Platform**.
2. **Clients → Create client → Web application**. Add `http://localhost:8000/auth/google/callback` as an authorised redirect URI. Copy the client ID and secret into `.env`.
3. **Audience**: while the app is in Testing, add your Gmail under **Test users**. Only listed users can sign in.

Open the frontend as `http://localhost:5500`, not `http://127.0.0.1:5500`. The browser treats them as different sites, so the session cookie set by `localhost:8000` is only sent from `localhost` pages.

### Database setup

The schema lives in `migrations/`. Run the files in order with `psql`, once per database.

**Fresh database:**

```bash
psql -U postgres -d url_shortener -f migrations/001_create_short_urls.sql
psql -U postgres -d url_shortener -f migrations/003_create_users.sql
```

**Existing database with the old `urls` table:** run both files. `002` copies your rows into `short_urls` and renames the old table to `urls_old` as a backup.

```bash
psql -U postgres -d url_shortener -f migrations/001_create_short_urls.sql
psql -U postgres -d url_shortener -f migrations/002_migrate_urls_data.sql
psql -U postgres -d url_shortener -f migrations/003_create_users.sql
```

Once everything works, you can remove the backup with `DROP TABLE urls_old;`.

Never edit a migration after it has been run. To change the schema, add a new numbered file (for example `003_add_clicks.sql`).

### Run the server

From the `Backend/` folder:

```bash
go run .
```

The server listens on `http://localhost:8000`. On startup it creates this month's and next month's partitions if they don't exist, and it refuses to start if it can't.

### Run the tests

```bash
go test ./...
```

## Storage details

- PostgreSQL stores the mapping between short IDs and original URLs in the `short_urls` table.
- Redis is used as a temporary cache for redirects.
- Guest URLs are marked with an expiry time; user URLs are stored without an expiry.
- Expired rows are not deleted, so the data stays available for analysis.

### Monthly partitions

`short_urls` is partitioned by month of `created_at` (UTC). Postgres stores each month in its own table, such as `short_urls_2026_10`, while the code only ever queries `short_urls`.

- `storage/partitions.go` creates the current and next month's partitions on startup, and re-checks once a day, so the next month always exists before it starts.
- Inserts go to the right month automatically.
- To analyse one month: `SELECT COUNT(*) FROM short_urls_2026_10;`
- To see which partition each row is in: `SELECT tableoid::regclass, short_id, created_at FROM short_urls;`
- To remove an old month: `DROP TABLE short_urls_2025_10;`

## URL validation

The service validates the URL before storing it and rejects invalid URLs with a `400 Bad Request` response. A URL is accepted only if:

- it uses `http` or `https` (so `javascript:`, `data:`, `file:` and `mailto:` are rejected)
- it has a host (`http://` alone is rejected)
- it has no login info before the host (`https://google.com@evil.com` is rejected)
- it is at most 2048 characters (the request body is capped at 4 KB)
- it doesn't point at this service itself (`localhost`, `127.0.0.1`)

Leading and trailing spaces are trimmed first. The server never fetches the submitted URL.
