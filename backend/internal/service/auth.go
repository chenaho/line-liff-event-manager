package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	Repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{Repo: repo}
}

type LineTokenResponse struct {
	Iss     string `json:"iss"`
	Sub     string `json:"sub"`
	Aud     string `json:"aud"`
	Exp     int64  `json:"exp"`
	Iat     int64  `json:"iat"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
}

func (s *AuthService) VerifyLineToken(idToken string) (*LineTokenResponse, error) {
	clientID := os.Getenv("LINE_CHANNEL_ID")
	log.Printf("[LINE VERIFY] Using LINE_CHANNEL_ID: %s", clientID)

	data := url.Values{}
	data.Set("id_token", idToken)
	if clientID != "" {
		data.Set("client_id", clientID)
	}

	log.Printf("[LINE VERIFY] Sending request to LINE API...")
	resp, err := http.PostForm("https://api.line.me/oauth2/v2.1/verify", data)
	if err != nil {
		log.Printf("[LINE VERIFY] HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[LINE VERIFY] Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[LINE VERIFY] LINE API returned non-200 status. Body: %s", string(body))
		return nil, errors.New("invalid line token")
	}

	var tokenResp LineTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		log.Printf("[LINE VERIFY] Failed to decode response: %v", err)
		return nil, err
	}

	log.Printf("[LINE VERIFY] Successfully verified token for user: %s", tokenResp.Sub)
	return &tokenResp, nil
}

func (s *AuthService) Login(ctx context.Context, idToken string) (string, *models.User, error) {
	// 1. Verify LINE Token
	lineProfile, err := s.VerifyLineToken(idToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to verify line token: %w", err)
	}

	// 2. Check/Update User in database
	existingUser, err := s.Repo.GetByID(ctx, lineProfile.Sub)

	var user models.User
	if err != nil {
		// User doesn't exist, create new user
		user = models.User{
			LineUserID:      lineProfile.Sub,
			LineDisplayName: lineProfile.Name,
			PictureURL:      lineProfile.Picture,
			Role:            "user",
			CreatedAt:       time.Now(),
		}

		// Check Admin Whitelist
		adminList := os.Getenv("ADMIN_LIST")
		log.Printf("[NEW USER] Checking ADMIN_LIST for user: %s", lineProfile.Sub)
		log.Printf("[NEW USER] ADMIN_LIST: %s", adminList)
		if strings.Contains(adminList, lineProfile.Sub) {
			user.Role = "admin"
			log.Printf("[NEW USER] User %s assigned role: admin", lineProfile.Sub)
		} else {
			log.Printf("[NEW USER] User %s assigned role: user", lineProfile.Sub)
		}

		if err := s.Repo.Create(ctx, &user); err != nil {
			return "", nil, err
		}
	} else {
		// Update existing user info
		user = *existingUser

		// Re-check admin role from ADMIN_LIST on every login
		adminList := os.Getenv("ADMIN_LIST")
		log.Printf("[EXISTING USER] Checking ADMIN_LIST for user: %s", lineProfile.Sub)
		log.Printf("[EXISTING USER] ADMIN_LIST: %s", adminList)
		log.Printf("[EXISTING USER] Current role in DB: %s", user.Role)

		newRole := "user"
		if strings.Contains(adminList, lineProfile.Sub) {
			newRole = "admin"
		}
		log.Printf("[EXISTING USER] New role assigned: %s", newRole)

		updates := map[string]interface{}{
			"lineDisplayName": lineProfile.Name,
			"pictureUrl":      lineProfile.Picture,
			"role":            newRole,
		}

		if err := s.Repo.UpdateFields(ctx, lineProfile.Sub, updates); err != nil {
			return "", nil, err
		}

		// Update local struct
		user.LineDisplayName = lineProfile.Name
		user.PictureURL = lineProfile.Picture
		user.Role = newRole
	}

	// 3. Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":  user.LineUserID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-do-not-use-in-prod"
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}
