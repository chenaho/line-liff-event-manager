package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"

	"cloud.google.com/go/firestore"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	Repo *repository.FirestoreRepository
}

func NewAuthService(repo *repository.FirestoreRepository) *AuthService {
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
	// Verify with LINE API
	// POST https://api.line.me/oauth2/v2.1/verify
	// Content-Type: application/x-www-form-urlencoded
	// id_token=...&client_id=...

	clientID := os.Getenv("LINE_CHANNEL_ID")
	// The LINE verify endpoint requires the Channel ID (not LIFF ID)

	data := url.Values{}
	data.Set("id_token", idToken)
	if clientID != "" {
		data.Set("client_id", clientID)
	}

	resp, err := http.PostForm("https://api.line.me/oauth2/v2.1/verify", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid line token")
	}

	var tokenResp LineTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *AuthService) Login(ctx context.Context, idToken string) (string, *models.User, error) {
	// 1. Verify LINE Token
	lineProfile, err := s.VerifyLineToken(idToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to verify line token: %w", err)
	}

	// 2. Check/Update User in Firestore
	userRef := s.Repo.Client.Collection("users").Doc(lineProfile.Sub)
	doc, err := userRef.Get(ctx)

	var user models.User
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || !doc.Exists() {
			// Create new user
			user = models.User{
				LineUserID:      lineProfile.Sub,
				LineDisplayName: lineProfile.Name,
				PictureURL:      lineProfile.Picture,
				Role:            "user",
				CreatedAt:       time.Now(),
			}
			// Check Admin Whitelist
			adminList := os.Getenv("ADMIN_LIST")
			if strings.Contains(adminList, lineProfile.Sub) {
				user.Role = "admin"
			}

			_, err = userRef.Set(ctx, user)
			if err != nil {
				return "", nil, err
			}
		} else {
			return "", nil, err
		}
	} else {
		// Update existing user info (optional, e.g. if name/picture changed)
		doc.DataTo(&user)
		updates := []firestore.Update{
			{Path: "lineDisplayName", Value: lineProfile.Name},
			{Path: "pictureUrl", Value: lineProfile.Picture},
		}
		// Re-check admin role if needed, or keep existing.
		// Let's keep existing role unless we want to force update from env.

		_, err = userRef.Update(ctx, updates)
		if err != nil {
			return "", nil, err
		}
		// Update local struct
		user.LineDisplayName = lineProfile.Name
		user.PictureURL = lineProfile.Picture
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
