package handlers

import (
	"net/http"

	"url-shortener/models"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

func NewAPIKeyHandler(db *gorm.DB) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: services.NewAPIKeyService(db),
	}
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req models.APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, apiKey)
}

func (h *APIKeyHandler) GetAPIKeys(c *gin.Context) {
	apiKeys, err := h.apiKeyService.GetAPIKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get API keys"})
		return
	}

	// Remove sensitive information
	var response []models.APIKeyResponse
	for _, key := range apiKeys {
		response = append(response, models.APIKeyResponse{
			KeyID:       key.KeyID,
			Name:        key.Name,
			Description: key.Description,
			IsActive:    key.IsActive,
			CreatedAt:   key.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *APIKeyHandler) DeactivateAPIKey(c *gin.Context) {
	keyID := c.Param("keyId")

	err := h.apiKeyService.DeactivateAPIKey(keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deactivated successfully"})
}
