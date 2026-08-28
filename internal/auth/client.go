package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the four Auth calls we use: register, login, check token, logout.
type Client struct {
	baseURL    string
	anonKey    string
	httpClient *http.Client
}

type Token struct {
	Access    string
	Refresh   string
	ExpiresIn int
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewClient(baseURL, anonKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		anonKey: anonKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) SignIn(email, password string) (Token, error) {
	return c.token("/auth/v1/token?grant_type=password", map[string]string{
		"email":    email,
		"password": password,
	})
}

func (c *Client) SignUp(email, password string) (Token, error) {
	t, err := c.token("/auth/v1/signup", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return Token{}, err
	}
	if t.Access == "" {
		return Token{}, fmt.Errorf("auth: no session returned")
	}
	return t, nil
}

func (c *Client) User(accessToken string) (User, error) {
	var u User
	err := c.call(http.MethodGet, "/auth/v1/user", nil, accessToken, &u)
	return u, err
}

func (c *Client) Refresh(refreshToken string) (Token, error) {
	return c.token("/auth/v1/token?grant_type=refresh_token", map[string]string{
		"refresh_token": refreshToken,
	})
}

func (c *Client) Logout(accessToken string) error {
	return c.call(http.MethodPost, "/auth/v1/logout", nil, accessToken, nil)
}

func (c *Client) token(path string, body map[string]string) (Token, error) {
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Session      *struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		} `json:"session"`
	}
	if err := c.call(http.MethodPost, path, body, "", &out); err != nil {
		return Token{}, err
	}
	t := Token{
		Access:    out.AccessToken,
		Refresh:   out.RefreshToken,
		ExpiresIn: out.ExpiresIn,
	}
	if t.Access == "" && out.Session != nil {
		t.Access = out.Session.AccessToken
		t.Refresh = out.Session.RefreshToken
		t.ExpiresIn = out.Session.ExpiresIn
	}
	if t.ExpiresIn <= 0 {
		t.ExpiresIn = 3600
	}
	return t, nil
}

type Error struct {
	Status int
}

func (e *Error) Error() string {
	return fmt.Sprintf("auth: http %d", e.Status)
}

func (c *Client) call(method, path string, body any, accessToken string, dest any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return &Error{Status: res.StatusCode}
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}
