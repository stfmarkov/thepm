-- App schema. user_id is auth.users.id (JWT sub). No FK onto auth.users:
-- Auth owns that table; Go copies the id from a verified session.
-- Deleting a project still cascades to links, notes, feedback, and roadmap items.

create extension if not exists pgcrypto;

create table public.projects (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  name text not null,
  slug text not null,
  status text not null default 'active',
  stack text not null default '',
  summary text not null default '',
  feedback_ingest_key text not null default replace(gen_random_uuid()::text, '-', ''),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint projects_user_slug_key unique (user_id, slug),
  constraint projects_feedback_ingest_key_key unique (feedback_ingest_key)
);

create index projects_user_id_idx on public.projects (user_id);

comment on column public.projects.user_id is 'auth.users.id';

create table public.links (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  project_id uuid not null references public.projects (id) on delete cascade,
  kind text not null,
  url text not null,
  label text not null default '',
  notes text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint links_kind_check check (
    kind in ('repo', 'site', 'hosting', 'domain', 'docs', 'other')
  )
);

create index links_user_id_idx on public.links (user_id);
create index links_project_id_idx on public.links (project_id);

comment on column public.links.user_id is 'auth.users.id';

create table public.notes (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  project_id uuid not null references public.projects (id) on delete cascade,
  body text not null,
  created_at timestamptz not null default now()
);

create index notes_user_id_idx on public.notes (user_id);
create index notes_project_id_idx on public.notes (project_id);

comment on column public.notes.user_id is 'auth.users.id';

create table public.feedback (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  project_id uuid not null references public.projects (id) on delete cascade,
  author_name text not null,
  author_email text,
  message text not null,
  rating integer,
  source text not null,
  received_at timestamptz not null default now(),
  constraint feedback_rating_check check (rating is null or rating between 1 and 5),
  constraint feedback_source_check check (source in ('ingest', 'manual'))
);

create index feedback_user_id_idx on public.feedback (user_id);
create index feedback_project_id_idx on public.feedback (project_id);

comment on column public.feedback.user_id is 'auth.users.id';

create table public.roadmap_items (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  project_id uuid not null references public.projects (id) on delete cascade,
  title text not null,
  status text not null default 'later',
  sort_order integer not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint roadmap_items_status_check check (
    status in ('later', 'next', 'doing', 'done')
  )
);

create index roadmap_items_user_id_idx on public.roadmap_items (user_id);
create index roadmap_items_project_id_idx on public.roadmap_items (project_id);

comment on column public.roadmap_items.user_id is 'auth.users.id';
