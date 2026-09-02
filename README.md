# thepm

A small hub for side projects: status, links, notes, and a feedback inbox. Sign in with email and password. Each account only sees its own projects.

Open a project to add links and notes inline. Feedback can be typed in by hand or posted from another app.

## Feedback from another app

Open a project and scroll to **Feedback inbox**. At the top of that section is **API ingest credentials**. Copy **Endpoint** and **Key** from there (Project ID is listed too; it is already part of the endpoint).

- **Endpoint** is the URL the other site POSTs to. It is unique per project.
- **Key** is a secret. Send it as the `X-Feedback-Key` header. Anyone with both the endpoint and the key can write into this inbox.

The other site POSTs JSON; no thepm login is required.

```javascript
await fetch("ENDPOINT", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-Feedback-Key": "KEY",
  },
  body: JSON.stringify({
    name: "Ada",
    email: "ada@example.com",
    message: "Loved the new screen.",
    rating: 5,
  }),
});
```

`name` and `message` are required. `email` and `rating` (1–5) are optional.

In the same panel, **Website** is the embedding origin (`https://myapp.com`). Leave it empty to allow any site. Rotate the key if it leaks — existing forms stop working.

Submissions show up in that project's inbox. Average rating sits next to the sort controls (newest, oldest, highest, lowest).

Field limits and error bodies: [docs/ingest.md](docs/ingest.md).
