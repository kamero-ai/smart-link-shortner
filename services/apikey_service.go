package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
	"url-shortener/models"
	"url-shortener/utils"

	"gorm.io/gorm"
)

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) CreateAPIKey(req models.APIKeyRequest) (*models.APIKeyResponse, error) {
	// Generate key ID and secret
	keyID, err := s.generateKeyID()
	if err != nil {
		return nil, err
	}

	keySecret, err := s.generateKeySecret()
	if err != nil {
		return nil, err
	}

	// Hash the secret for storage
	hashedSecret := utils.HashAPIKey(keySecret)

	apiKey := models.APIKey{
		KeyID:       keyID,
		KeySecret:   hashedSecret,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	result := s.db.Create(&apiKey)
	if result.Error != nil {
		return nil, result.Error
	}

	return &models.APIKeyResponse{
		KeyID:       apiKey.KeyID,
		KeySecret:   keySecret, // Return unhashed secret only on creation
		Name:        apiKey.Name,
		Description: apiKey.Description,
		IsActive:    apiKey.IsActive,
		CreatedAt:   apiKey.CreatedAt,
	}, nil
}

func (s *APIKeyService) ValidateAPIKey(keyID, keySecret string) (*models.APIKey, error) {
	var apiKey models.APIKey
	result := s.db.Where("key_id = ? AND is_active = ?", keyID, true).First(&apiKey)
	if result.Error != nil {
		return nil, errors.New("invalid API key")
	}

	hashedSecret := utils.HashAPIKey(keySecret)
	if apiKey.KeySecret != hashedSecret {
		return nil, errors.New("invalid API key")
	}

	// Update last used time
	now := time.Now()
	s.db.Model(&apiKey).Update("last_used_at", &now)

	return &apiKey, nil
}

func (s *APIKeyService) GetAPIKeys() ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	result := s.db.Where("is_active = ?", true).Find(&apiKeys)
	return apiKeys, result.Error
}

func (s *APIKeyService) DeactivateAPIKey(keyID string) error {
	return s.db.Model(&models.APIKey{}).Where("key_id = ?", keyID).
		Update("is_active", false).Error
}

func (s *APIKeyService) generateKeyID() (string, error) {
	for {
		bytes := make([]byte, 10)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		keyID := "ak_" + hex.EncodeToString(bytes)[:17] // ak_ + 17 chars = 20 total

		// Check if key ID already exists
		var existing models.APIKey
		result := s.db.Where("key_id = ?", keyID).First(&existing)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return keyID, nil
		}
	}
}

func (s *APIKeyService) generateKeySecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(bytes)[:62], nil // sk_ + 62 chars = 64 total
}
