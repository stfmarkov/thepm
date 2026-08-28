// Package auth talks to Supabase Auth (email + password).
// Go calls the Auth HTTP API with the anon key, then sets httpOnly cookies.
// No OTP, magic links, or SMTP. Confirm-email stays off in the dashboard.
package auth
