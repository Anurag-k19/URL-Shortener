# URL Shortener

A lightweight HTTP-based URL shortening service written in Go. This service allows users to convert long URLs into short, easy-to-share links with optional expiration for guests.

## Features

- **URL Shortening**: Convert long URLs into 6-character short codes
- **User Management**: Support for registered users and guest users
- **Expiration Handling**: Guest URLs expire after 5 minutes; registered user URLs never expire
- **Deduplication**: Prevents duplicate short URLs for the same user and original URL
- **Zero-Collision Generation**: Automatically handles short ID collision prevention
- **Redirect Service**: Retrieve original URLs from short codes

## Project Structure

```
.
├── main.go                 # Server setup and route registration
├── go.mod                  # Go module definition
├── handlers/
│   ├── shorten.go         # POST /shorten endpoint for creating short URLs
│   └── redirect.go        # GET / endpoint for redirecting to original URLs
├── storage/
│   └── memory.go          # In-memory storage for URLs using maps
└── utils/
    └── generator.go       # Short ID generation utility
```

## API Endpoints

### Create Short URL
**POST** `/shorten`

**Headers:**
- `X-User-ID` (optional): User identifier for URL persistence. If not provided, defaults to "guest"

**Request Body:**
```json
{
  "url": "https://example.com/very/long/url"
}
```

**Response:**
```json
{
  "short_url": "http://localhost:8000/abc123"
}
```

**Guest URLs:** Expire after 5 minutes  
**Registered User URLs:** Never expire

### Redirect to Original URL
**GET** `/{shortId}`

Redirects to the original URL if found and not expired.

**Success:** HTTP 302 (Found)  
**Not Found:** HTTP 404  
**Expired:** HTTP 410 (Gone)

## Getting Started

### Prerequisites
- Go 1.26.1 or higher
- PostgreSQL 12 or higher

### Installation

1. Navigate to the project directory:
```bash
cd "URL Shortner"
```

2. Download Go dependencies:
```bash
go mod download
```

3. Build the module:
```bash
go build
```

4. Set up PostgreSQL database:

   a. Create the database:
   ```bash
   createdb url_shortener
   ```

   b. Initialize the table:
   ```bash
   psql url_shortener < init.sql
   ```

   c. Set up environment variables by copying `.env.example` to `.env`:
   ```env
   DB_USER=your_postgres_username
   PASSWORD=your_postgres_password
   DBNAME=url_shortener
   SSLMODE=disable
   ```

### Running the Server

```bash
go run main.go
```

The server will start on `http://localhost:8000`

## Usage Examples

### Create a Short URL (Guest)
```bash
curl -X POST http://localhost:8000/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com"}'
```

### Create a Short URL (Registered User)
```bash
curl -X POST http://localhost:8000/shorten \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{"url": "https://www.example.com"}'
```

### Access Short URL
```bash
curl -i http://localhost:8000/abc123
```

## Data Storage

The application uses **PostgreSQL** for persistent storage with the following schema:

- **short_id**: Unique 6-character identifier for the shortened URL
- **long_url**: The original long URL
- **user_id**: User identifier (guest or registered)
- **expires_at**: Expiration timestamp (NULL for permanent URLs)
- **created_at**: URL creation timestamp

## Implementation Details

### Short ID Generation
- **Length**: 6 characters
- **Charset**: a-z, A-Z, 0-9 (62 possible characters)
- **Uniqueness**: Collision detection queries database
- **Probability**: 62^6 = ~56 billion possible combinations

### URL Expiration
- **Guest URLs**: Expire 5 seconds after creation (configurable in code)
- **User URLs**: No expiration (valid indefinitely)
- **Cleanup**: Expired URLs are deleted on redirect attempts

### Deduplication
- Database query prevents duplicate short IDs for the same user and URL
- Returns existing short ID if user tries to shorten the same URL again
- Prevents duplicate entries and serves as URL existence check

### Database Indexes
- `idx_short_id`: For fast lookups during redirects
- `idx_user_id`: For querying user's shortened URLs

## Troubleshooting

### "URL not Found" Error
1. **Database not initialized**: Run `psql url_shortener < init.sql`
2. **Database connection failed**: Check `.env` file credentials and PostgreSQL service is running
3. **Table doesn't exist**: Verify table creation with `psql url_shortener -c "\dt"`

### "Failed to store URL" Error
1. **Database permissions**: Ensure database user has INSERT permissions
2. **Connection pool exhausted**: Check if application is closing database connections properly

### Connection Refused
1. Make sure PostgreSQL is running: `psql --version` and try connecting manually
2. Verify connection string in `.env` matches your PostgreSQL setup

## Future Improvements

- Add persistent storage (database)
- Implement user authentication
- Add URL analytics (click tracking)
- Implement custom short aliases
- Add API rate limiting
- Create web UI for URL management
- Support batch URL shortening
- Add URL preview before redirect

## License

This project is open source and available under the MIT License.
