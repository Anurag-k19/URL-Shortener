# Snip frontend

A framework-free, API-first frontend for a Go URL-shortening backend. Open it through a static web server; ES modules do not work reliably from `file://` URLs.

## Run locally

The frontend is plain HTML, CSS, and JavaScript. Python is not an application
dependency; the following command only serves the static files during development:

```sh
python3 -m http.server 5500
```

Then visit `http://localhost:5500`. The API defaults to `http://localhost:8080`.

You can instead use VS Code Live Server or any other static file server.

## Backend contract

The frontend currently sends:

```http
POST /shorten
Content-Type: application/json

{"url":"https://example.com/a/long/path"}
```

The response may use any of these common shapes:

```json
{"short_url":"http://localhost:8080/redirect?code=abc123","code":"abc123"}
```

```json
{"shortUrl":"http://localhost:8080/redirect?code=abc123","shortCode":"abc123"}
```

```json
{"code":"abc123"}
```

When only a code is returned, the frontend produces `/redirect?code=abc123`. If your Go endpoint expects a path parameter or a different query name, change only `getRedirectUrl()` in `js/api/url.service.js`.

The Go server must allow the frontend origin through CORS, including `POST`, `Content-Type`, and later `Authorization` headers. CORS permission is controlled by the server; browser JavaScript cannot grant it.

## Configuration

Change the `api-base-url` meta tag in `index.html` per environment. All endpoint paths, timeouts, and retry defaults live in `js/config.js`.

Only `js/api/client.js` calls `fetch()`. Pages and components call service functions, and services call the centralized client. Authentication can later register a JWT provider through `setAuthTokenProvider()` without changing feature components.

## How the application works end to end

### 1. The browser starts the application

The browser loads `index.html`, which provides three root elements:

- `#navbar-root` for the navigation bar.
- `#app` for the current page.
- `#toast-root` for success and error notifications.

At the bottom of the document, `index.html` loads `js/app.js` as an ES module:

```html
<script type="module" src="./js/app.js"></script>
```

`app.js` is the application entry point. It finds the root elements, creates the
toast manager, renders the navigation bar, and renders the home page. No backend
request is made during this initial rendering.

```text
index.html
    -> js/app.js
        -> components/navbar.js
        -> components/toast.js
        -> pages/home.js
```

### 2. The user submits a long URL

The form and its event handlers live in `js/pages/home.js`. When the form is
submitted, the page prevents the browser's normal form submission and passes the
input to `validateUrl()` from `js/utils/validators.js`.

The validator checks that:

- A value was entered.
- It is not longer than 2,048 characters.
- It is a complete URL.
- It uses `http://` or `https://`.

Invalid input stays in the browser. The page displays an accessible error and does
not contact the backend.

### 3. The page calls the URL service

For valid input, `home.js` calls:

```js
const shortened = await createShortUrl(validation.value);
```

`createShortUrl()` belongs to `js/api/url.service.js`. This service knows which
backend endpoint is needed and converts the page's value into the backend request
body:

```json
{
  "url": "https://example.com/a/long/path"
}
```

The page does not know that the endpoint is `/shorten`, and the service does not
render any HTML.

### 4. The centralized client contacts the Go backend

`url.service.js` passes the request to `apiClient.post()` in
`js/api/client.js`. This is the only file allowed to call `fetch()`.

Before sending the request, the client:

- Combines the API base URL with `/shorten`.
- Adds `Accept: application/json` and `Content-Type: application/json`.
- Serializes the request body as JSON.
- Starts a timeout using `AbortController`.
- Uses browser CORS mode.
- Adds a bearer token in the future when an auth token provider is registered.

With the default configuration, the resulting request is:

```http
POST http://localhost:8080/shorten
Accept: application/json
Content-Type: application/json

{"url":"https://example.com/a/long/path"}
```

POST requests are not automatically retried because repeating a creation request
can create duplicate records. The client supports safe automatic retries for
idempotent requests such as GET.

### 5. The Go backend creates the short URL

The Go backend validates and stores the URL, creates a code, and returns JSON. It
is the single source of truth; the frontend does not create or persist shortened
URLs itself.

The URL service accepts common Go response naming styles. For example:

```json
{
  "short_url": "http://localhost:8080/redirect?code=abc123",
  "code": "abc123"
}
```

`normalizeShortenResponse()` in `url.service.js` converts the backend response to
one stable frontend shape:

```js
{
  id: "abc123",
  code: "abc123",
  shortUrl: "http://localhost:8080/redirect?code=abc123",
  originalUrl: null,
  createdAt: null,
}
```

This normalization prevents backend field-name changes from spreading into UI
components.

If the backend returns only a code, `getRedirectUrl()` constructs the public URL
using the configured `/redirect` endpoint.

### 6. The page displays the result

After the service resolves, `home.js`:

1. Stores the normalized result in `js/state/store.js`.
2. Updates the result link's text and `href`.
3. Makes the result card visible.
4. Restores the submit button from its loading state.
5. Shows a success toast.

The state store contains runtime UI state only. It does not replace backend data
or treat browser storage as the source of truth.

### 7. Copying and opening the short URL

When the user selects **Copy link**, `home.js` reads the result from the store and
calls `copyText()` in `js/utils/clipboard.js`. That utility uses the browser
Clipboard API and includes a fallback for older browsers.

When somebody opens the copied link, their browser requests the Go backend:

```http
GET http://localhost:8080/redirect?code=abc123
```

The Go backend looks up `abc123` and returns an HTTP redirect to the original URL.
The frontend is not involved in this visit. If your backend instead uses a path
such as `/redirect/abc123`, update only `getRedirectUrl()` in
`js/api/url.service.js`.

### 8. Errors and timeouts

The HTTP client parses successful JSON responses and converts failures into a
consistent `ApiError`. It provides useful messages for:

- Network failures.
- Request timeouts.
- Backend validation errors.
- Rate limits.
- Server errors.

The error travels back through the same layers:

```text
Go backend or network
    -> api/client.js
        -> api/url.service.js
            -> pages/home.js
                -> inline error + error toast
```

Pages decide how an error should look. The client decides how an HTTP failure is
represented. This keeps networking and presentation concerns separate.

### Complete shortening flow

```text
User submits URL
    -> home.js validates it
        -> url.service.js creates the request
            -> client.js sends POST /shorten
                -> Go backend stores URL and returns code
            <- client.js parses the JSON response
        <- url.service.js normalizes the response
    <- home.js updates state and displays the short link
User copies or opens link
    -> browser requests GET /redirect?code=...
        -> Go backend redirects to the original URL
```

### Configuration and CORS

The API host comes from the `api-base-url` meta tag in `index.html`. `js/config.js`
reads it once and exposes the base URL, endpoint paths, timeout, and retry settings
to the API layer.

CORS must be configured in the Go backend. For local development, the backend
should allow the frontend origin (normally `http://localhost:5500`), the required
HTTP methods, and the `Content-Type` header. Add the `Authorization` header to the
backend's CORS configuration when authentication is introduced.

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
