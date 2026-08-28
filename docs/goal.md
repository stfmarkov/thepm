# Goal

thepm is a small web app for keeping track of small projects in one place.

Each project should hold the information that is otherwise scattered across bookmarks, hosting dashboards, and chat messages:

- What the project is, and its current status
- The stack
- Important URLs (repo, live site, docs)
- Where it is hosted, and a link to the hosting dashboard
- Where the domain comes from (registrar, DNS)
- Notes
- User feedback (typed in here, or submitted from a live app)
- A simple roadmap

Live apps will embed a feedback form that posts **directly into that project's inbox**. That is a public JSON endpoint, not an HTMX page: the other site sends a project id plus form fields; thepm stores the row; you read it on the project page.

It is a personal tool first. It should also be safe to release later as one of the small products on the same list: multi-user from day one, email + password login via Supabase Auth, every row owned by a user. No OTP, no magic links, no SMTP: turn off “confirm email” so signup does not send mail. Email is only the login identifier.

## Who it is for

- The operator (you), managing several small projects
- Later: other people who sign in with email and a password and only see their own projects

## What success looks like

- Open a project and see stack, hosting, domain, and key links without hunting
- Add a note, a piece of feedback, or a roadmap item without a page reload
- A production app can POST feedback into a project without the submitter having a thepm account
- Sign in with email and password; sign out cleanly
- Two users cannot see each other's data

## What it is not

- Not a general project manager (no sprints, assignees, or time tracking)
- Not a SPA. The signed-in UI is server-rendered HTML; HTMX swaps fragments
- Not a second API in the **thepm** browser session. The page never talks to Supabase directly
- Exception: one unauthenticated JSON ingest route for feedback from other apps. That route is CORS + rate-limited, CSRF-exempt, and never uses the operator session cookie

## Stack (locked)

| Layer | Choice |
|---|---|
| HTTP | Go, Gin |
| HTML | templ |
| Behavior | HTMX 2 (no Alpine) |
| CSS | Pico.css |
| Queries | sqlc + pgx |
| Database | Supabase Postgres |
| Migrations | Official Supabase CLI (`supabase/migrations`) |
| Auth | Supabase Auth: email + password. No OTP, no magic link, confirm-email off |
| Session | httpOnly cookies set by Go after Auth returns tokens |
| CSRF | Required on the signed-in UI (the app may be public) |
| Public ingest | JSON `POST` for feedback from other apps (no session cookie) |

## Public feedback ingest

Other projects (including the one already in production) will host a form that writes into thepm.

Likely payload (fields not locked yet):

| Field | Required | Notes |
|---|---|---|
| Project id | yes | Which inbox to write to |
| Name | yes (probably) | Display name of the person giving feedback |
| Email | no | Optional contact |
| Message | yes | The feedback body |
| Rating | no | Integer 1–5 if present |

Because the form lives in public HTML, the project id is not a secret. Each project also gets a **rotateable ingest key**. The request must include both. Rotating the key kills a leaked form without changing the project id.

Do not use implicit-flow tricks, supabase-js, or the operator's cookies on this route.

## Rule

The thepm UI talks only to Go. Other apps may `POST` JSON to the ingest endpoint. Go talks to Supabase Auth over the Auth HTTP API (anon key), and to Postgres over `SUPABASE_CONNECTION_STRING` (falling back to `SUPABASE_POOLER_STRING`).
