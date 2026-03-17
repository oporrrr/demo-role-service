package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"demo-role-service/entity"
	"demo-role-service/repository"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	authCenter *AuthCenterService
	userRepo   *repository.UserRepository
	redis      *redis.Client
}

func NewUserService(authCenter *AuthCenterService, userRepo *repository.UserRepository, rdb *redis.Client) *UserService {
	return &UserService{authCenter: authCenter, userRepo: userRepo, redis: rdb}
}

// LoginOutcome holds the parsed result from Auth Center login.
// On non-SUCCESS responses, only StatusCode + RawBody are populated (for pass-through).
type LoginOutcome struct {
	StatusCode       int
	Code             string
	RawBody          []byte // populated for non-SUCCESS — pass directly to client
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	Profile          *UserProfile
}

type loginResponseData struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

type loginResponse struct {
	StatusCode int               `json:"statusCode"`
	Code       string            `json:"code"`
	Data       loginResponseData `json:"data"`
}

// Login forwards reqBody to Auth Center using the provided client credentials.
// On success fetches the full profile, caches it, and returns a parsed LoginOutcome.
// On non-SUCCESS the raw upstream body is preserved in LoginOutcome.RawBody.
func (s *UserService) Login(reqBody []byte, clientID, clientSecret string) (*LoginOutcome, error) {
	body, statusCode, err := s.authCenter.Login(reqBody, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	var resp loginResponse
	if err := json.Unmarshal(body, &resp); err != nil || resp.Code != "SUCCESS" {
		return &LoginOutcome{StatusCode: statusCode, Code: resp.Code, RawBody: body}, nil
	}

	profile, err := s.authCenter.GetAccountInformation(resp.Data.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get account information: %w", err)
	}

	ttl := time.Duration(resp.Data.ExpiresIn) * time.Second
	if s.redis != nil {
		s.cacheUser(profile, ttl)
	}

	return &LoginOutcome{
		StatusCode:       statusCode,
		Code:             resp.Code,
		AccessToken:      resp.Data.AccessToken,
		RefreshToken:     resp.Data.RefreshToken,
		ExpiresIn:        resp.Data.ExpiresIn,
		RefreshExpiresIn: resp.Data.RefreshExpiresIn,
		Profile:          profile,
	}, nil
}

// Logout forwards the logout request to Auth Center (pass-through).
func (s *UserService) Logout(refreshToken, userToken string) ([]byte, int, error) {
	return s.authCenter.Logout(refreshToken, userToken)
}

// RegisterRequest mirrors the fields accepted by POST /api/v1/auth/register.
type RegisterRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PrefixName  string `json:"prefixName"`
	Gender      string `json:"gender"`
	DateOfBirth string `json:"dateOfBirth"`
	CountryCode string `json:"countryCode"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

// Register forwards the request to Auth Center using the provided client credentials,
// then upserts the user into the local DB and cache.
// Returns (rawBody, statusCode, error) for pass-through on non-201 responses.
func (s *UserService) Register(req RegisterRequest, clientID, clientSecret string) ([]byte, int, error) {
	payload := map[string]interface{}{
		"password": req.Password,
	}
	setIfNotEmpty := func(key, val string) {
		if val != "" {
			payload[key] = val
		}
	}
	setIfNotEmpty("firstName", req.FirstName)
	setIfNotEmpty("lastName", req.LastName)
	setIfNotEmpty("prefixName", req.PrefixName)
	setIfNotEmpty("gender", req.Gender)
	setIfNotEmpty("dateOfBirth", req.DateOfBirth)
	setIfNotEmpty("countryCode", req.CountryCode)
	setIfNotEmpty("phoneNumber", req.PhoneNumber)
	setIfNotEmpty("email", req.Email)

	result, rawBody, statusCode, err := s.authCenter.Register(payload, clientID, clientSecret)
	if err != nil {
		return nil, 500, fmt.Errorf("auth center error: %w", err)
	}
	if statusCode != 201 {
		return rawBody, statusCode, nil
	}

	if result.AccessToken != "" {
		profile, err := s.authCenter.GetAccountInformation(result.AccessToken)
		if err == nil {
			user := &entity.User{
				AccountID:         profile.ID,
				KeycloakUserID:    profile.KeycloakUserID,
				FirstName:         profile.FirstName,
				LastName:          profile.LastName,
				PrefixName:        profile.PrefixName,
				Gender:            profile.Gender,
				DateOfBirth:       profile.DateOfBirth,
				Email:             profile.Email,
				Username:          profile.Username,
				AccountStatus:     profile.AccountStatus,
				AkidID:            profile.AkidID,
				CisNumber:         profile.CisNumber,
				ProfilePictureURL: profile.ProfilePictureURL,
				CountryCode:       profile.CountryCode,
				PhoneNumber:       profile.PhoneNumber,
				Role:              "NEW_USER",
			}
			_ = s.userRepo.Upsert(user)
			if s.redis != nil {
				s.cacheUser(profile, 15*time.Minute)
			}
		}
	}

	return rawBody, statusCode, nil
}

func (s *UserService) cacheUser(profile *UserProfile, ttl time.Duration) {
	data, err := json.Marshal(profile)
	if err != nil {
		return
	}
	s.redis.Set(context.Background(), "user:"+profile.ID, data, ttl)
}
