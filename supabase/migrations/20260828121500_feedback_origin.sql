-- Website that may send ingest (browser Origin). Empty means any origin.
alter table public.projects
  add column feedback_origin text not null default '';

comment on column public.projects.feedback_origin is
  'Origin of the embedding site (scheme + host), e.g. https://myapp.com. Empty allows any origin.';
