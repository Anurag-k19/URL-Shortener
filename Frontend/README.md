# Snip frontend

This frontend is a framework-free, API-first web app that talks to the Go backend for URL shortening and redirecting. It is intentionally simple: plain HTML, CSS, and JavaScript are used, and the app is loaded as an ES module from the browser.

## How the frontend is structured

The frontend is organized as a small layered application:

- Pages contain user-facing behavior and event handlers
- Services handle API requests and response normalization
- The client layer is the only place that calls the browser fetch API
- Shared state stores runtime UI values
- UI components render reusable elements such as navbar, toast, and modal

The main files are:

- [index.html](index.html) – the entry HTML page
- [js/app.js](js/app.js) – application bootstrap
- [js/pages/home.js](js/pages/home.js) – the main page behavior
- [js/api/url.service.js](js/api/url.service.js) – URL shortening API logic
- [js/api/client.js](js/api/client.js) – centralized HTTP client
- [js/config.js](js/config.js) – app and API configuration
- [js/state/store.js](js/state/store.js) – shared UI state
- [js/utils/validators.js](js/utils/validators.js) – input validation
- [js/utils/clipboard.js](js/utils/clipboard.js) – copy-to-clipboard helper

## How the app starts

When the browser opens [index.html](index.html), it loads [js/app.js](js/app.js) as an ES module.

That bootstrap file performs three main tasks:

1. It finds the root containers in the HTML document
2. It creates the toast system and navigation UI
3. It renders the home page

The startup flow looks like this:

```text
index.html
  -> app.js
    -> navbar component
    -> toast component
    -> home page
```

No backend request is made during initial page load. The page is rendered first, and the user interaction begins only after that.

## How the user flow works

### 1. User opens the page

The browser loads [index.html](index.html), which contains placeholders such as:

- a navbar root
- an app root
- a toast root

These are then populated by the JavaScript modules.

### 2. User submits a URL

When the user enters a long URL and submits the form, the logic in [js/pages/home.js](js/pages/home.js) runs.

That file:

- prevents the browser from doing a normal form submission
- reads the input value
- sends it to the validator in [js/utils/validators.js](js/utils/validators.js)

The validator checks that the URL:

- is not empty
- is not too long
- has a valid URL format
- uses http or https

If validation fails, the page shows an inline error and does not call the backend.

### 3. The page calls the URL service

When validation succeeds, [js/pages/home.js](js/pages/home.js) calls a service function from [js/api/url.service.js](js/api/url.service.js):

```js
const shortened = await createShortUrl(validation.value);
```

This service is responsible for converting the page input into the shape expected by the backend. It does not render UI, and it does not directly interact with the browser fetch API.

### 4. The service calls the centralized HTTP client

The URL service then calls [js/api/client.js](js/api/client.js), which is the single place in the frontend that uses fetch.

That file:

- builds the final request URL from the base URL and endpoint
- adds JSON headers
- serializes the request body
- sets a timeout using AbortController
- handles retry rules for safe requests
- converts network failures into an API error object

In other words, the frontend uses a strict layering rule:

```text
Page -> Service -> HTTP client -> Backend
```

### 5. The backend responds

The backend returns a JSON response containing a short URL. The service then normalizes that response into a frontend-friendly object.

That normalization step is important because the frontend should not depend on the exact field names used by the backend. The service ensures the UI always receives a consistent structure.

### 6. The page updates the UI

After the service returns, [js/pages/home.js](js/pages/home.js) updates the shared state in [js/state/store.js](js/state/store.js) and renders the result in the page.

The UI flow is:

1. store the short link result
2. update the link text and href
3. show the success card
4. reset the button from the loading state
5. show a success toast

## How the frontend calls the backend

The call chain is:

```text
index.html
  -> app.js
    -> home page
      -> validators.js
      -> url.service.js
        -> client.js
          -> fetch() to backend
```

The backend endpoint used for creating a short URL is:

```http
POST /shorten
```

The frontend sends JSON like this:

```json
{
  "url": "https://example.com/very/long/url"
}
```

The response is then used to create the short link shown to the user.

## How the redirect flow works

After the user receives a short link, opening it in the browser sends a request directly to the backend.

That request is not handled by the frontend anymore. The browser simply visits the URL and the backend responds with an HTTP redirect to the original destination.

The redirect path is:

```text
Browser opens short URL
  -> backend /redirect route
    -> checks database and Redis
    -> redirects to original URL
```

## How the frontend is configured

The app gets its API base URL from the meta tag in [index.html](index.html). The configuration is loaded from [js/config.js](js/config.js).

That config file defines:

- the API base URL
- endpoint paths such as /shorten and /redirect
- timeout values
- retry behavior

This makes it easy to change the backend address without editing every page or component.

## Backend integration summary

The frontend is intentionally simple and follows a clean separation of concerns:

- UI pages handle interaction and display
- Services handle API contracts
- The client handles networking
- Shared state stores temporary UI state

This design is useful because if the backend response shape changes, only the service layer needs to be updated instead of the entire UI.

## How the backend works

The backend is implemented in the Go files under [../Backend](../Backend).

### Main entry point

The server is started in [../Backend/main.go](../Backend/main.go). It:

- initializes PostgreSQL and Redis
- creates the router
- registers the shorten and redirect handlers
- starts the HTTP server on port 8000

### Request routing

The router in [../Backend/main.go](../Backend/main.go) routes requests to two main handlers:

- /shorten → handled by [../Backend/handlers/shorten.go](../Backend/handlers/shorten.go)
- /redirect → handled by [../Backend/handlers/redirect.go](../Backend/handlers/redirect.go)

### Shorten handler

The shorten handler in [../Backend/handlers/shorten.go](../Backend/handlers/shorten.go) receives the POST request from the frontend.

It performs these steps:

1. checks that the request method is POST
2. reads the URL from the JSON body
3. validates the URL
4. verifies that the URL is reachable
5. generates a short ID
6. stores the short URL mapping in PostgreSQL
7. returns the generated short URL to the frontend

### Redirect handler

The redirect handler in [../Backend/handlers/redirect.go](../Backend/handlers/redirect.go) handles browser requests for the short URL.

It:

1. reads the short code from the query parameter or path
2. checks the database for the mapping
3. validates whether the link is still active based on the expiry logic
4. checks Redis for a cached redirect
5. redirects the browser to the original URL if the link is valid

### Storage layer

The backend uses two storage systems:

- [../Backend/storage/db.go](../Backend/storage/db.go) for PostgreSQL connection and initialization
- [../Backend/storage/redis.go](../Backend/storage/redis.go) for Redis connection and cache access

PostgreSQL is the source of truth for the URL mapping, while Redis is used as a temporary cache layer for faster access.

## End-to-end architecture summary

Here is the full call chain:

```text
User types URL in browser
  -> home page validates input
  -> url service builds request
  -> HTTP client sends POST /shorten
  -> Go backend stores the URL in PostgreSQL
  -> backend returns short URL to frontend
  -> frontend displays the short link

User opens short link
  -> browser requests /redirect?code=...
  -> backend checks database and Redis
  -> backend redirects to original URL
```

## Important notes

- The frontend must be served from a static server such as Python http.server or Live Server.
- The backend runs on port 8000.
- The frontend runs on port 5500 locally.
- CORS must be enabled in the backend so the browser can make requests from the static frontend.

## Where to add new features

Keep this request flow for every backend feature:

```text
Page or component -> API service -> HTTP client -> Go backend
```

| Change | Location |
| --- | --- |
| API base URL, endpoint path, or timeout | `js/config.js` |
| Backend request/response handling | `js/api/*.service.js` |
| Reusable UI element | `js/components/` |
| Page behavior and event handling | `js/pages/` |
| Shared application state | `js/state/store.js` |
| Validation, clipboard, storage, and helpers | `js/utils/` |
| Styling | `css/styles.css` |
| Application startup | `js/app.js` |

Do not call `fetch()` from a page or component. Do not place endpoint URLs in UI
files. If a backend response changes, normalize it inside its service so the UI
continues receiving the same frontend-friendly object.

### Example: add custom aliases

First, extend the existing function in `js/api/url.service.js`:

```js
export async function createShortUrl(longUrl, options = {}) {
  return apiClient.post(CONFIG.api.endpoints.shorten, {
    url: longUrl,
    custom_alias: options.customAlias,
  });
}
```

Next, add an alias input to `js/pages/home.js` and pass its value to the service:

```js
const shortened = await createShortUrl(validation.value, {
  customAlias: aliasInput.value.trim(),
});
```

The HTTP client, application startup, and unrelated components do not need to
change.

### Example: add a completely new backend feature

For URL folders, for example:

1. Add folder endpoint paths to `js/config.js`.
2. Create `js/api/folder.service.js` with functions such as `getFolders()`,
   `createFolder()`, and `addUrlToFolder()`. These functions call `apiClient`.
3. Export those functions from `js/api/index.js`.
4. Create reusable folder UI in `js/components/folder-card.js`.
5. Create page behavior in `js/pages/folders.js`.
6. Add navigation or routing only if the feature needs a separate page.

Use the same pattern for analytics, history, QR codes, bulk shortening, API keys,
uploads, and admin features. Existing placeholder services and pages can be filled
in instead of recreated.

## Adding authentication later

Authentication remains backend-owned. When the Go endpoints are available:

1. Add endpoint paths such as `/login`, `/logout`, and `/refresh` to `js/config.js`.
2. Implement their calls in `js/api/auth.service.js`.
3. Add login UI in a page or modal component.
4. Register a token provider with `setAuthTokenProvider()` so `client.js` adds the
   `Authorization: Bearer <token>` header centrally.
5. Handle `401` responses in the API/auth layer instead of duplicating auth logic
   across pages.

Never place secrets, private API keys, OAuth client secrets, or signing keys in
this frontend. Those belong in the Go backend.
