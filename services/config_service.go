package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zgsm-ai/client-manager/dao"
	"github.com/zgsm-ai/client-manager/models"
)

/**
 * ConfigService handles business logic for configuration operations
 * @description
 * - Implements configuration management business rules
 * - Handles validation and authorization
 */
type ConfigService struct {
	configDAO *dao.ConfigDAO
	log       *logrus.Logger
}

/**
 * NewConfigService creates a new ConfigService instance
 * @param {dao.ConfigDAO} configDAO - Configuration data access object
 * @param {logrus.Logger} log - Logger instance
 * @returns {*ConfigService} New ConfigService instance
 */
func NewConfigService(configDAO *dao.ConfigDAO, log *logrus.Logger) *ConfigService {
	return &ConfigService{
		configDAO: configDAO,
		log:       log,
	}
}

/**
 * GetConfiguration retrieves a configuration by type and key
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} namespace - Configuration type
 * @param {string} key - Configuration key
 * @returns {*models.Configuration, error} Configuration and error if any
 * @description
 * - Validates input parameters
 * - Retrieves configuration from cache or database
 * - Logs access for audit purposes
 * @throws
 * - Validation errors for missing parameters
 * - Database access errors
 * - Cache operation errors
 */
func (s *ConfigService) GetConfiguration(ctx context.Context, namespace, key string) (*models.Configuration, error) {
	// Validate input parameters
	if namespace == "" {
		return nil, &ValidationError{Field: "namespace", Message: "namespace is required"}
	}
	if key == "" {
		return nil, &ValidationError{Field: "key", Message: "key is required"}
	}

	// Get configuration
	config, err := s.configDAO.GetConfiguration(ctx, namespace, key)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"namespace": namespace,
			"key":       key,
		}).Error("Failed to get configuration")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"namespace": namespace,
		"key":       key,
	}).Info("Configuration retrieved successfully")

	return config, nil
}

/**
 * GetConfigurations retrieves a list of configurations with pagination and search
 * @param {context.Context} ctx - Context for request cancellation
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @param {string} search - Search term
 * @returns {map[string]interface{}, error} Response containing configurations and pagination info
 * @description
 * - Validates pagination parameters
 * - Performs search if provided
 * - Returns structured response with pagination metadata
 * @throws
 * - Validation errors for invalid pagination parameters
 * - Database query errors
 */
func (s *ConfigService) GetConfigurations(ctx context.Context, page, pageSize int, search string) (map[string]interface{}, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get configurations
	configs, total, err := s.configDAO.GetConfigurations(ctx, page, pageSize, search)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"page":      page,
			"page_size": pageSize,
			"search":    search,
		}).Error("Failed to get configurations")
		return nil, err
	}

	// Prepare response
	response := map[string]interface{}{
		"data": configs,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	}

	s.log.WithFields(logrus.Fields{
		"page":      page,
		"page_size": pageSize,
		"search":    search,
		"total":     total,
	}).Info("Configurations retrieved successfully")

	return response, nil
}

/**
 * GetNamespaceConfigurations retrieves all configurations for a specific namespace
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} namespace - Namespace name
 * @returns {[]models.Configuration, error} List of configurations and error if any
 * @description
 * - Validates namespace parameter
 * - Retrieves all configurations in namespace
 * - Logs access for audit purposes
 * @throws
 * - Validation errors for missing namespace
 * - Database query errors
 */
func (s *ConfigService) GetNamespaceConfigurations(ctx context.Context, namespace string) ([]models.Configuration, error) {
	// Validate namespace parameter
	if namespace == "" {
		return nil, &ValidationError{Field: "namespace", Message: "namespace is required"}
	}

	// Get namespace configurations
	configs, err := s.configDAO.GetNamespaceConfigurations(ctx, namespace)
	if err != nil {
		s.log.WithError(err).WithField("namespace", namespace).Error("Failed to get namespace configurations")
		return nil, err
	}

	s.log.WithField("namespace", namespace).Info("Namespace configurations retrieved successfully")

	return configs, nil
}

/**
 * GetSpecificConfiguration retrieves a specific configuration by namespace and key
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} namespace - Namespace name
 * @param {string} key - Configuration key
 * @returns {*models.Configuration, error} Configuration and error if any
 * @description
 * - Validates input parameters
 * - Retrieves configuration from cache or database
 * - Logs access for audit purposes
 * @throws
 * - Validation errors for missing parameters
 * - Database access errors
 * - Cache operation errors
 */
func (s *ConfigService) GetSpecificConfiguration(ctx context.Context, namespace, key string) (*models.Configuration, error) {
	// Validate input parameters
	if namespace == "" {
		return nil, &ValidationError{Field: "namespace", Message: "namespace is required"}
	}
	if key == "" {
		return nil, &ValidationError{Field: "key", Message: "key is required"}
	}

	// Get specific configuration
	config, err := s.configDAO.GetSpecificConfiguration(ctx, namespace, key)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"namespace": namespace,
			"key":       key,
		}).Error("Failed to get specific configuration")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"namespace": namespace,
		"key":       key,
	}).Info("Specific configuration retrieved successfully")

	return config, nil
}

/**
 * CreateConfiguration creates a new configuration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Configuration data
 * @returns {*models.Configuration, error} Created configuration and error if any
 * @description
 * - Validates configuration data
 * - Checks for duplicates
 * - Creates configuration record
 * - Invalidates related cache
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 * - Cache operation errors
 */
func (s *ConfigService) CreateConfiguration(ctx context.Context, data map[string]interface{}) (*models.Configuration, error) {
	// Validate and extract configuration data
	namespace, ok := data["namespace"].(string)
	if !ok || namespace == "" {
		return nil, &ValidationError{Field: "namespace", Message: "namespace is required and must be a string"}
	}

	key, ok := data["key"].(string)
	if !ok || key == "" {
		return nil, &ValidationError{Field: "key", Message: "key is required and must be a string"}
	}

	value, _ := data["value"].(string)
	description, _ := data["description"].(string)
	namespace, _ = data["namespace"].(string)
	key, _ = data["key"].(string)

	// Check for duplicates
	existing, err := s.configDAO.GetSpecificConfiguration(ctx, namespace, key)
	if err == nil && existing != nil {
		return nil, &ConflictError{Message: "configuration already exists"}
	}

	// Create configuration
	config := &models.Configuration{
		Namespace:   namespace,
		Key:         key,
		Value:       value,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.configDAO.CreateConfiguration(ctx, config)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"namespace": namespace,
			"key":       key,
		}).Error("Failed to create configuration")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"namespace": namespace,
		"key":       key,
	}).Info("Configuration created successfully")

	return config, nil
}

/**
 * UpdateConfiguration updates an existing configuration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {uint} id - Configuration ID
 * @param {map[string]interface{}} data - Configuration data to update
 * @returns {*models.Configuration, error} Updated configuration and error if any
 * @description
 * - Validates configuration exists
 * - Validates update data
 * - Updates configuration record
 * - Invalidates and updates cache
 * @throws
 * - Validation errors for invalid data
 * - Database update errors
 * - Cache operation errors
 */
func (s *ConfigService) UpdateConfiguration(ctx context.Context, id uint, data map[string]interface{}) (*models.Configuration, error) {
	// Get existing configuration
	var config models.Configuration
	err := s.configDAO.GetConfigurationByID(ctx, id, &config)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &NotFoundError{Message: "configuration not found"}
		}
		return nil, err
	}

	// Update fields
	if value, ok := data["value"].(string); ok {
		config.Value = value
	}
	if description, ok := data["description"].(string); ok {
		config.Description = description
	}
	if namespace, ok := data["namespace"].(string); ok {
		config.Namespace = namespace
	}
	if key, ok := data["key"].(string); ok {
		config.Key = key
	}

	config.UpdatedAt = time.Now()

	// Update configuration
	err = s.configDAO.UpdateConfiguration(ctx, &config)
	if err != nil {
		s.log.WithError(err).WithField("id", id).Error("Failed to update configuration")
		return nil, err
	}

	s.log.WithField("id", id).Info("Configuration updated successfully")

	return &config, nil
}

/**
 * DeleteConfiguration deletes a configuration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {uint} id - Configuration ID
 * @returns {error} Error if any
 * @description
 * - Validates configuration exists
 * - Performs soft delete
 * - Invalidates related cache
 * @throws
 * - Validation errors for non-existent configuration
 * - Database deletion errors
 * - Cache operation errors
 */
func (s *ConfigService) DeleteConfiguration(ctx context.Context, id uint) error {
	// Check if configuration exists
	var config models.Configuration
	err := s.configDAO.GetConfigurationByID(ctx, id, &config)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &NotFoundError{Message: "configuration not found"}
		}
		return err
	}

	// Delete configuration
	err = s.configDAO.DeleteConfiguration(ctx, id)
	if err != nil {
		s.log.WithError(err).WithField("id", id).Error("Failed to delete configuration")
		return err
	}

	s.log.WithField("id", id).Info("Configuration deleted successfully")

	return nil
}

/**
 * ValidationError represents a validation error
 * @description
 * - Contains field name and error message
 * - Used for input validation failures
 */
type ValidationError struct {
	Field   string
	Message string
}

/**
 * Error returns the error message
 * @returns {string} Error message
 */
func (e *ValidationError) Error() string {
	return e.Message
}

/**
 * ConflictError represents a conflict error
 * @description
 * - Used for duplicate resource conflicts
 * - Contains error message
 */
type ConflictError struct {
	Message string
}

/**
 * Error returns the error message
 * @returns {string} Error message
 */
func (e *ConflictError) Error() string {
	return e.Message
}

/**
 * NotFoundError represents a not found error
 * @description
 * - Used for resource not found scenarios
 * - Contains error message
 */
type NotFoundError struct {
	Message string
}

/*
*
  - Error returns the error message
  - @returns {string} Error message
*/
func (e *NotFoundError) Error() string {
	return e.Message
}
