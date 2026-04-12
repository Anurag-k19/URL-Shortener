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
- Go 1.16 or higher

### Installation

1. Navigate to the project directory:
```bash
cd "URL Shortner"
```

2. Build the module:
```bash
go build
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

The application uses **in-memory maps** for storage:

- **URLmap**: Maps short IDs to their original URLs and expiration times
- **Reversemap**: Maps (userID:originalURL) to short IDs for deduplication

**Note:** Data is not persisted. All URLs are lost when the server restarts.

## Implementation Details

### Short ID Generation
- **Length**: 6 characters
- **Charset**: a-z, A-Z, 0-9 (62 possible characters)
- **Uniqueness**: Collision detection ensures all IDs are unique
- **Probability**: 62^6 = ~56 billion possible combinations

### URL Expiration
- **Guest URLs**: Expire 5 minutes after creation
- **User URLs**: No expiration (valid indefinitely)
- **Cleanup**: Expired URLs are deleted on access attempts

### Deduplication
- Same user requesting the same URL receives the same short ID
- Prevents duplicate entries and serves as URL existence check

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
