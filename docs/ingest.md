# Public feedback ingest

Unauthenticated JSON. No session cookie, no CSRF. The embedding site sends the project id (from the thepm project page) and the ingest key.

## Request

```
POST /api/v1/projects/<project_id>/feedback
Content-Type: application/json
X-Feedback-Key: <ingest key>
```

```json
{
  "name": "Ada",
  "email": "ada@example.com",
  "message": "Loved the new screen.",
  "rating": 5
}
```

| Field | Required | Notes |
|---|---|---|
| `name` | yes | Display name, max 200 characters |
| `email` | no | Optional contact, max 320 characters |
| `message` | yes | Body, max 8000 characters |
| `rating` | no | Integer 1–5 if present |

Wrong project id or key → `404` with the same body. Do not treat that as “id exists but key is wrong.”

## Website (origin)

On the project page, **Website** is the origin of the embedding site (`https://myapp.com`). Empty means any site.

If a website is saved and the request sends `Origin`, it must match. Browsers send `Origin` on cross-site `fetch`. `curl` can omit it.

CORS on this route allows any origin to call it (`Access-Control-Allow-Origin: *`). The website field is the extra check, not a second ingest URL.

## Rate limits

- 30 requests / 15 minutes / IP
- 60 requests / 15 minutes / project

Over limit → `429` `{"ok":false,"error":"too many requests"}`.

## Responses

```json
{"ok": true}
{"ok": false, "error": "not found"}
```

Other `error` values: `invalid request`, `name and message are required`, `field too long`, `invalid email`, `rating must be 1 to 5`, `origin not allowed`, `could not save`.

Rows are stored with `source = ingest` and the project owner's `user_id`.

## Snippet

```javascript
await fetch("https://THEPM_HOST/api/v1/projects/PROJECT_ID/feedback", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-Feedback-Key": "INGEST_KEY",
  },
  body: JSON.stringify({
    name: "Ada",
    email: "ada@example.com",
    message: "Loved the new screen.",
    rating: 5,
  }),
});
```
