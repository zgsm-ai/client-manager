package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/services"
)

/**
 * ConfigController handles HTTP requests for configuration operations
 * @description
 * - Implements RESTful API endpoints for configuration management
 * - Handles request validation and response formatting
 * - Integrates with ConfigService for business logic
 */
type ConfigController struct {
	configService *services.ConfigService
	log           *logrus.Logger
}

/**
 * NewConfigController creates a new ConfigController instance
 * @param {logrus.Logger} log - Logger instance
 * @returns {*ConfigController} New ConfigController instance
 */
func NewConfigController(log *logrus.Logger) *ConfigController {
	// Initialize DAOs and services here
	configService := services.NewConfigService(nil, log) // Will be properly initialized later

	return &ConfigController{
		configService: configService,
		log:           log,
	}
}

// GetConfigurations handles GET /configurations request
// @Summary Get configurations list
// @Description Retrieve a list of configurations with pagination and search
// @Tags Configuration
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Param search query string false "Search term"
// @Success 200 {object} map[string]interface{} "Configurations list with pagination"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/configurations [get]
func (cc *ConfigController) GetConfigurations(c *gin.Context) {
	// Get and validate pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get search parameter
	search := c.Query("search")

	// Get configurations
	response, err := cc.configService.GetConfigurations(c.Request.Context(), page, pageSize, search)
	if err != nil {
		cc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "Configurations retrieved successfully",
		"data":    response,
	})
}

// GetNamespaceConfigurations handles GET /configurations/{namespace} request
// @Summary Get configurations by namespace
// @Description Retrieve all configurations within a specific namespace
// @Tags Configuration
// @Accept json
// @Produce json
// @Param namespace path string true "Namespace name"
// @Success 200 {object} map[string]interface{} "Configurations list"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/configurations/{namespace} [get]
func (cc *ConfigController) GetNamespaceConfigurations(c *gin.Context) {
	// Get path parameter
	namespace := c.Param("namespace")

	// Get namespace configurations
	configs, err := cc.configService.GetNamespaceConfigurations(c.Request.Context(), namespace)
	if err != nil {
		cc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "Namespace configurations retrieved successfully",
		"data":    configs,
	})
}

// GetSpecificConfiguration handles GET /configurations/{namespace}/{key} request
// @Summary Get specific configuration
// @Description Retrieve a specific configuration by namespace and key
// @Tags Configuration
// @Accept json
// @Produce json
// @Param namespace path string true "Namespace name"
// @Param key path string true "Configuration key"
// @Success 200 {object} map[string]interface{} "Configuration data"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Configuration not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/configurations/{namespace}/{key} [get]
func (cc *ConfigController) GetSpecificConfiguration(c *gin.Context) {
	// Get path parameters
	namespace := c.Param("namespace")
	key := c.Param("key")

	// Get specific configuration
	config, err := cc.configService.GetSpecificConfiguration(c.Request.Context(), namespace, key)
	if err != nil {
		cc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "Specific configuration retrieved successfully",
		"data":    config,
	})
}

/**
 * handleError handles errors and returns appropriate HTTP responses
 * @param {gin.Context} c - Gin context
 * @param {error} err - Error to handle
 * @description
 * - Maps different error types to appropriate HTTP status codes
 * - Returns standardized error response format
 * - Logs errors for debugging
 */
func (cc *ConfigController) handleError(c *gin.Context, err error) {
	// Log error
	cc.log.WithError(err).Error("Request processing failed")

	// Handle different error types
	switch e := err.(type) {
	case *services.ValidationError:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "validation.error",
			"message": e.Message,
			"field":   e.Field,
		})
	case *services.ConflictError:
		c.JSON(http.StatusConflict, gin.H{
			"code":    "conflict.error",
			"message": e.Message,
		})
	case *services.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "notfound.error",
			"message": e.Message,
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal.error",
			"message": "Internal server error",
		})
	}
}

/**
 * SetConfigService sets the config service (used for dependency injection)
 * @param {services.ConfigService} configService - Config service instance
 * @description
 * - Allows setting the config service after controller creation
 * - Used for proper dependency injection
 */
func (cc *ConfigController) SetConfigService(configService *services.ConfigService) {
	cc.configService = configService
}
