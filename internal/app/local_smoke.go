package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func RunLocalSmoke(baseURL string) error {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if err := expectStatus(client, http.MethodGet, baseURL+"/healthz", nil, nil, http.StatusOK); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	userToken, err := loginAndCaptureToken(client, baseURL, userEmail, []string{userRequestedPassword, userFallbackPassword})
	if err != nil {
		return fmt.Errorf("user login failed: %w", err)
	}
	adminToken, err := superuserLoginAndCaptureToken(client, baseURL, adminEmail, []string{adminRequestedPassword, adminFallbackPassword})
	if err != nil {
		return fmt.Errorf("admin login failed: %w", err)
	}

	if err := expectStatus(client, http.MethodGet, baseURL+"/admin/users", nil, authCookie(adminToken), http.StatusOK); err != nil {
		return fmt.Errorf("admin access failed: %w", err)
	}
	if err := expectStatus(client, http.MethodGet, baseURL+"/admin/users", nil, authCookie(userToken), http.StatusForbidden); err != nil {
		return fmt.Errorf("regular user should be forbidden on admin route: %w", err)
	}

	fmt.Println("Local smoke passed.")
	return nil
}

func superuserLoginAndCaptureToken(client *http.Client, baseURL string, email string, candidates []string) (string, error) {
	var lastErr error
	for _, password := range candidates {
		token, err := superuserLoginWithPassword(client, baseURL, email, password)
		if err == nil && strings.TrimSpace(token) != "" {
			return token, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no valid credentials")
	}
	return "", lastErr
}

func loginAndCaptureToken(client *http.Client, baseURL string, email string, candidates []string) (string, error) {
	var lastErr error
	for _, password := range candidates {
		token, err := loginWithPassword(client, baseURL, email, password)
		if err == nil && token != "" {
			return token, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no valid credentials")
	}
	return "", lastErr
}

func loginWithPassword(client *http.Client, baseURL string, email string, password string) (string, error) {
	form := url.Values{}
	form.Set("email", email)
	form.Set("password", password)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/login", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusSeeOther {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "pb_auth" && strings.TrimSpace(cookie.Value) != "" {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("missing pb_auth cookie")
}

func superuserLoginWithPassword(client *http.Client, baseURL string, email string, password string) (string, error) {
	payload := fmt.Sprintf(`{"identity":%q,"password":%q}`, email, password)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/collections/_superusers/auth-with-password", bytes.NewBufferString(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var authResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", err
	}
	if strings.TrimSpace(authResp.Token) == "" {
		return "", fmt.Errorf("missing token")
	}
	return authResp.Token, nil
}

func authCookie(token string) []*http.Cookie {
	return []*http.Cookie{{Name: "pb_auth", Value: token, Path: "/"}}
}

func expectStatus(client *http.Client, method, target string, body io.Reader, cookies []*http.Cookie, expected int) error {
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != expected {
		return fmt.Errorf("expected %d got %d", expected, resp.StatusCode)
	}
	return nil
}
