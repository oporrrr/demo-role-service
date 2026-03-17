package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// UserProfile mirrors the full account information response from Auth Center.
// Role is populated by the role service after login — not from Auth Center.
type UserProfile struct {
	ID                string  `json:"id"`
	KeycloakUserID    string  `json:"keycloakUserId"`
	FirstName         string  `json:"firstName"`
	LastName          string  `json:"lastName"`
	PrefixName        string  `json:"prefixName"`
	Gender            string  `json:"gender"`
	DateOfBirth       string  `json:"dateOfBirth"`
	Email             string  `json:"email"`
	Username          string  `json:"username"`
	AccountStatus     string  `json:"accountStatus"`
	AkidID            int64   `json:"akidId"`
	CisNumber         *string `json:"cisNumber"`
	ProfilePictureURL *string `json:"profilePictureUrl"`
	CountryCode       string  `json:"countryCode"`
	PhoneNumber       string  `json:"phoneNumber"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
	Role              string  `json:"role"`
}

// accountInfoResponse is the raw response shape from GET /api/v1/account/information.
type accountInfoResponse struct {
	ID                string  `json:"id"`
	KeycloakUserID    string  `json:"keycloakUserId"`
	FirstName         string  `json:"firstName"`
	LastName          string  `json:"lastName"`
	PrefixName        string  `json:"prefixName"`
	Gender            string  `json:"gender"`
	DateOfBirth       string  `json:"dateOfBirth"`
	Email             string  `json:"email"`
	Username          string  `json:"username"`
	AccountStatus     string  `json:"accountStatus"`
	AkidID            int64   `json:"akidId"`
	CisNumber         *string `json:"cisNumber"`
	ProfilePictureURL *string `json:"profilePictureUrl"`
	CountryCode       string  `json:"countryCode"`
	PhoneNumber       string  `json:"phoneNumber"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// RegisterResult is the data block returned by Auth Center on successful registration.
type RegisterResult struct {
	AccountID   string `json:"accountId"`
	AccessToken string `json:"accessToken"`
}

type clientTokenResponse struct {
	Code string `json:"code"`
	Data struct {
		AccessToken string `json:"accessToken"`
	} `json:"data"`
}

type cachedToken struct {
	token   string
	expiry  time.Time
	mu      sync.RWMutex
}

type AuthCenterService struct {
	baseURL    string
	httpClient *http.Client
	tokenCache sync.Map // key: clientID → *cachedToken
}

func NewAuthCenterService(baseURL string) *AuthCenterService {
	return &AuthCenterService{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// doRequest performs an authenticated HTTP request to Auth Center.
// If userToken is provided it is used as Bearer; otherwise a client token is fetched
// using the provided clientID and clientSecret.
func (s *AuthCenterService) doRequest(method, path string, body interface{}, userToken, clientID, clientSecret string) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, s.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if userToken != "" {
		req.Header.Set("Authorization", "Bearer "+userToken)
	} else {
		clientToken, err := s.getClientToken(clientID, clientSecret)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Authorization", "Bearer "+clientToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, nil
}

// getClientToken returns a cached client token per clientID, refreshing when expired.
func (s *AuthCenterService) getClientToken(clientID, clientSecret string) (string, error) {
	entry, _ := s.tokenCache.LoadOrStore(clientID, &cachedToken{})
	ct := entry.(*cachedToken)

	ct.mu.RLock()
	if ct.token != "" && time.Now().Before(ct.expiry) {
		token := ct.token
		ct.mu.RUnlock()
		return token, nil
	}
	ct.mu.RUnlock()

	cred := fmt.Sprintf("%s:%s", clientID, clientSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(cred))

	req, err := http.NewRequest("GET", s.baseURL+"/api/v1/client/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+encoded)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get client token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result clientTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse client token response: %w", err)
	}
	if result.Code != "SUCCESS" || result.Data.AccessToken == "" {
		return "", fmt.Errorf("get client token failed: %s", string(body))
	}

	ct.mu.Lock()
	ct.token = result.Data.AccessToken
	ct.expiry = time.Now().Add(50 * time.Minute)
	ct.mu.Unlock()

	return ct.token, nil
}

// Logout forwards the logout request to Auth Center using the user's access token.
func (s *AuthCenterService) Logout(refreshToken, userToken string) ([]byte, int, error) {
	body := map[string]string{"refreshToken": refreshToken}
	return s.doRequest("POST", "/api/v1/auth/logout", body, userToken, "", "")
}

// Login forwards the login payload to Auth Center using the system's client credentials.
func (s *AuthCenterService) Login(reqBody []byte, clientID, clientSecret string) ([]byte, int, error) {
	var payload interface{}
	if err := json.Unmarshal(reqBody, &payload); err != nil {
		return nil, 0, fmt.Errorf("invalid login body: %w", err)
	}
	return s.doRequest("POST", "/api/v1/auth/login", payload, "", clientID, clientSecret)
}

// Register forwards the registration payload using the system's client credentials.
// isIncludeAKID is forced to false.
// Returns (result, rawBody, statusCode, error) — rawBody is always populated for pass-through.
func (s *AuthCenterService) Register(req map[string]interface{}, clientID, clientSecret string) (*RegisterResult, []byte, int, error) {
	req["isIncludeAKID"] = false

	body, statusCode, err := s.doRequest("POST", "/api/v1/auth/register", req, "", clientID, clientSecret)
	if err != nil {
		return nil, nil, 0, err
	}

	var result struct {
		StatusCode int            `json:"statusCode"`
		Code       string         `json:"code"`
		Data       RegisterResult `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, body, statusCode, err
	}
	return &result.Data, body, statusCode, nil
}

// ProfileFromToken decodes the JWT payload (without verification) and maps
// standard claims to UserProfile. At minimum populates AccountID from "sub".
func ProfileFromToken(tokenString string) *UserProfile {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return &UserProfile{}
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return &UserProfile{}
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return &UserProfile{}
	}
	str := func(key string) string {
		if v, ok := claims[key].(string); ok {
			return v
		}
		return ""
	}
	return &UserProfile{
		ID:          str("sub"),
		FirstName:   str("given_name"),
		LastName:    str("family_name"),
		Email:       str("email"),
		PhoneNumber: str("phone_number"),
	}
}

// GetAccountInformation fetches the full user profile using the user's access token.
func (s *AuthCenterService) GetAccountInformation(accessToken string) (*UserProfile, error) {
	body, _, err := s.doRequest("GET", "/api/v1/account/information", nil, accessToken, "", "")
	if err != nil {
		return nil, err
	}

	var result struct {
		Data accountInfoResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	d := result.Data
	return &UserProfile{
		ID:                d.ID,
		KeycloakUserID:    d.KeycloakUserID,
		FirstName:         d.FirstName,
		LastName:          d.LastName,
		PrefixName:        d.PrefixName,
		Gender:            d.Gender,
		DateOfBirth:       d.DateOfBirth,
		Email:             d.Email,
		Username:          d.Username,
		AccountStatus:     d.AccountStatus,
		AkidID:            d.AkidID,
		CisNumber:         d.CisNumber,
		ProfilePictureURL: d.ProfilePictureURL,
		CountryCode:       d.CountryCode,
		PhoneNumber:       d.PhoneNumber,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}, nil
}
