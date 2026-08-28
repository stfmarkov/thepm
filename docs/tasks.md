# Tasks

Work top to bottom. A later group assumes the earlier one is done. Checkboxes are the source of progress; update them as we go.

---

## 1. Setup Supabase

Stand up the hosted project, CLI, Auth (email + password only), and the first schema. Do this before writing app code that needs a database.

### 1.1 Create the cloud project

- [x] Create a Supabase project (region close to where the app will be hosted)
- [x] Wait until the project is healthy (API and database ready)
- [x] Copy these values somewhere local and gitignored (not into the repo):
  - Project URL and anon (public) key — Go uses these when calling Auth (email + password)
  - Service role key — keep off the request path; do not put it in the browser
  - `SUPABASE_CONNECTION_STRING` (direct) and `SUPABASE_POOLER_STRING` (fallback if the direct host cannot be reached)



### 1.2 Install and log in to the CLI

- [x] Install the official Supabase CLI
- [x] Run `supabase login` and confirm the CLI can see the new project
- [x] Confirm `supabase --version` is recent enough for `db push`



### 1.3 Init and link this repo

- [x] From the repo root, run `supabase init` so `supabase/` exists
- [x] Link the local folder to the cloud project (`supabase link --project-ref …`)
- [x] Confirm `supabase/config.toml` is present and committed
- [x] Confirm secrets stay out of git (`.gitignore` covers `.env`, CLI temp files)



### 1.4 Configure Auth (email + password, no mail)

Supabase Auth still creates rows in `auth.users`. We do not send email.

- [x] Enable the Email provider
- [x] Enable password sign-in
- [x] Disable magic link / OTP if the dashboard allows it
- [x] Disable **Confirm email** (signup must work with no SMTP)
- [x] Allow new users to sign up
- [x] Anon key is required: Go calls Auth with it. Never put it in the browser

### 1.5 Local vs cloud database

Pick one workflow and stay on it until the first deploy.

- [x] **Preferred while building:** cloud project as the database (`supabase db push` against the linked project)
- [ ] Optional later: `supabase start` for a local stack; not required for v1
- [x] Confirm you can connect with `psql` or the dashboard SQL editor using the copied connection string



### 1.6 First migration: app schema

Create `supabase/migrations/<timestamp>_init.sql`. Every table is multi-user from day one.

- [x] Enable `pgcrypto` (or `uuid-ossp`) if we generate UUIDs in Postgres
- [x] `projects` — `id`, `user_id` (uuid, matches `auth.users.id`), `name`, `slug`, `status`, `stack`, `summary`, `feedback_ingest_key` (random, unique, rotateable), timestamps
- [x] `links` — `user_id`, `project_id`, `kind` (repo, site, hosting, domain, docs, other), `url`, `label`, `notes`
- [x] `notes` — `user_id`, `project_id`, `body`, `created_at`
- [x] `feedback` — `user_id`, `project_id`, `author_name`, `author_email` (nullable), `message`, `rating` (nullable, 1–5), `source` (`ingest` or `manual`), `received_at`
- [x] `roadmap_items` — `user_id`, `project_id`, `title`, `status`, `sort_order`
- [x] Foreign keys: child rows cascade with the project; `user_id` is not a FK to `auth.users` (Auth owns that table; Go copies `sub` from the session)
- [x] Indexes: `(user_id)`, `(project_id)`, unique `(user_id, slug)` on projects, unique `feedback_ingest_key` on projects
- [x] Apply with `supabase db push`
- [x] Spot-check tables in the dashboard

### 1.7 Row ownership (v1 vs later)

- [x] Signed-in UI: every sqlc query takes `user_id` from the verified Auth session (`sub`). Never load a row by id alone
- [ ] Public ingest: look up the project by `id` + `feedback_ingest_key`, then insert feedback with **that project's** `user_id` (the owner). No session
- [ ] Later: RLS as a second fence, on a database role that does not bypass RLS. Do not treat the `postgres` / service connection as proof that RLS is working

---



## 2. Setup the Go app

Empty module, runnable server, HTML layout. No features yet.

### 2.1 Module and layout

- [x] `go mod init` for this repo
- [x] Folders: `cmd/server`, `internal/http`, `internal/auth`, `internal/db`, `internal/views`, `static`
- [x] `.env.example` with `SUPABASE_CONNECTION_STRING`, `SUPABASE_POOLER_STRING`, `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SESSION_SECRET`, `ADDR`
- [x] `.gitignore` for binaries, `.env`, templ/sqlc generated if we choose not to commit them (prefer committing generated Go)



### 2.2 Tooling

- [x] Install templ, sqlc, and a live-reload tool (air or wgo)
- [x] Makefile or small scripts: `templ generate`, `sqlc generate`, `run`
- [x] Confirm `templ generate && go run ./cmd/server` starts



### 2.3 Gin skeleton

- [x] Gin engine with logger and recovery
- [x] Health route
- [x] Static files (HTMX and Pico: vendored under `static/` or CDN in the layout — pick one)
- [x] templ layout: base HTML, Pico, HTMX, a content slot
- [x] Render helper: status + `component.Render(c.Request.Context(), c.Writer)` — do not use `c.HTML` for pages. `c.JSON` is allowed only on the public ingest route
- [x] A single “it works” page so we can see HTML before Auth

---



## 3. Database access from Go

- [x] `sqlc` config pointed at `supabase/migrations` (schema) and `internal/db/queries`
- [x] pgx pool from `SUPABASE_CONNECTION_STRING`, falling back to `SUPABASE_POOLER_STRING`
- [x] One smoke query (for example, count projects for a fake user) so a bad URL fails at boot
- [x] Commit generated sqlc code

---



## 4. Auth in Go

Go is the only Auth client the browser sees. Email + password against Supabase Auth. No OTP, no magic links.

### 4.1 Session cookies

- [x] After a successful sign-in, set httpOnly, Secure (in production), SameSite=Lax cookies for access + refresh tokens (or one opaque session id that maps to them)
- [x] Never put JWTs in `localStorage` or in a response body that HTMX displays
- [x] Logout clears cookies and, if cheap, revokes the refresh token

### 4.2 Register and login

- [x] `GET /login` — email + password form; link to register
- [x] `POST /login` — call Auth password grant with the anon key; on success set cookies and redirect home
- [x] `GET /register` — email + password (+ confirm)
- [x] `POST /register` — call Auth sign-up; on success set cookies and redirect home
- [x] Error states as HTML fragments (unknown failure, bad password, duplicate email), not JSON. Do not say which of email/password was wrong on login

### 4.3 Middleware

- [x] Load the user from the access token on each request
- [x] If expired, try refresh once; otherwise redirect to `/login`
- [x] Attach `user_id` (Auth `sub`) to the Gin context for handlers
- [x] Skip session auth on `POST /api/v1/projects/:project_id/feedback` (ingest key is the credential)
- [x] CSRF on signed-in state-changing routes (login and register included). Exempt `/api/…` ingest
- [x] Rate-limit `/login`, `/register`, and the public feedback ingest route before the first public URL

---



## 5. Core product

All writes scoped by `user_id`. HTMX swaps the section that changed.

### 5.1 Projects

- [x] List my projects
- [x] Create (name, slug, status, stack, summary)
- [x] Detail page
- [x] Edit metadata
- [x] Delete with `hx-confirm`



### 5.2 Links (stack, hosting, domain, URLs)

- [x] List links on the project page, grouped by kind
- [x] Add / edit / delete a link (url, label, notes)
- [x] Kinds include at least: repo, site, hosting dashboard, domain/registrar, docs, other



### 5.3 Notes

- [ ] Chronological notes on the project
- [ ] Add a note (HTMX prepend/append)
- [ ] Delete a note



### 5.4 Feedback inbox (signed-in UI)

- [ ] List feedback on the project page (name, email if present, message, rating, source, date)
- [ ] Manually add a row (same fields) for feedback you heard elsewhere
- [ ] Delete a row
- [ ] Show the ingest URL, project id, and ingest key on the project page (copyable)
- [ ] Rotate ingest key (`hx-confirm`); old forms stop working



### 5.5 Public feedback ingest API

Unauthenticated JSON. Called by forms on other apps. Field names can stay flexible until the form is designed; validate what we accept.

- [ ] `POST /api/v1/projects/:project_id/feedback`
- [ ] Require ingest key (header `X-Feedback-Key` or body field — pick one and document it)
- [ ] Lookup project by id **and** key; 404 if either is wrong (do not leak which)
- [ ] Accept JSON body, likely: `name`, `email` (optional), `message`, `rating` (optional 1–5)
- [ ] Reject empty message, overlong strings, rating outside 1–5
- [ ] Insert with that project's `user_id` (the owner), `source = ingest`
- [ ] CORS on this route so a browser form on another origin can POST (v1: allow all origins, or a per-project origin list later)
- [ ] No session cookie, no CSRF token, no supabase-js
- [ ] Rate-limit per IP and per project
- [ ] Return a small JSON success/error body the embedding form can display
- [ ] Document the contract in `docs/` once the field names are locked (snippet the other apps will copy)



### 5.6 Roadmap

- [ ] Ordered list of items with a status (for example: later, next, doing, done)
- [ ] Add, toggle status, delete
- [ ] Optional: reorder (can wait until the rest works)

---



## 6. Hardening and first deploy

- [ ] Secure cookies, HTTPS
- [ ] Confirm register, login, and logout on the production origin
- [ ] Deploy the Go binary (Fly, Railway, Render, or a VPS — decide at deploy time)
- [ ] `SUPABASE_POOLER_STRING` is set so boot can fall back on IPv4-only networks
- [ ] Smoke test: two browsers, two users, no cross-visible rows
- [ ] Smoke test: POST ingest with a real project id + key from another origin; row appears in the inbox
- [ ] Smoke test: wrong key does not create a row and does not reveal whether the project id exists