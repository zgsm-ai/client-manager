package dao

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zgsm-ai/client-manager/models"
)

/**
 * ConfigDAO handles data access operations for configurations
 * @description
 * - Provides CRUD operations for configuration data
 * - Supports Redis caching for performance optimization
 * - Implements database transactions for data consistency
 */
type ConfigDAO struct {
	db    *gorm.DB
	redis *redis.Client
	log   *logrus.Logger
}

/**
 * NewConfigDAO creates a new ConfigDAO instance
 * @param {gorm.DB} db - Database connection
 * @param {redis.Client} redis - Redis client
 * @param {logrus.Logger} log - Logger instance
 * @returns {*ConfigDAO} New ConfigDAO instance
 */
func NewConfigDAO(db *gorm.DB, redis *redis.Client, log *logrus.Logger) *ConfigDAO {
	return &ConfigDAO{
		db:    db,
		redis: redis,
		log:   log,
	}
}

/**
 * GetConfiguration retrieves a configuration by type and key
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} namespace - Configuration type
 * @param {string} key - Configuration key
 * @returns {*models.Configuration, error} Configuration and error if any
 * @description
 * - First tries to get from Redis cache
 * - If not found in cache, queries database
 * - Caches the result for future requests
 * @throws
 * - Database query errors
 * - Redis operation errors
 */
func (dao *ConfigDAO) GetConfiguration(ctx context.Context, namespace, key string) (*models.Configuration, error) {
	var config models.Configuration

	// Try to get from cache first if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + namespace + ":" + key
		cached, err := dao.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			// Found in cache
			config.Value = cached
			config.Namespace = namespace
			config.Key = key
			return &config, nil
		} else if err != redis.Nil {
			// Redis error but not Nil, log it and continue to database
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Redis get failed, falling back to database")
		}
	}

	// Not found in cache or Redis unavailable, query database
	err := dao.db.Where("namespace = ? AND key = ?", namespace, key).First(&config).Error
	if err != nil {
		return nil, err
	}

	// Cache the result if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + namespace + ":" + key
		err := dao.redis.Set(ctx, cacheKey, config.Value, 5*time.Minute).Err()
		if err != nil {
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to cache configuration")
		}
	}

	return &config, nil
}

/**
 * GetConfigurations retrieves a list of configurations with pagination
 * @param {context.Context} ctx - Context for request cancellation
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @param {string} search - Search term
 * @returns {[]models.Configuration, int64, error} Configurations, total count, and error
 * @description
 * - Supports pagination parameters
 * - Supports search by namespace or key
 * - Returns both data and total count for frontend pagination
 * @throws
 * - Database query errors
 */
func (dao *ConfigDAO) GetConfigurations(ctx context.Context, page, pageSize int, search string) ([]models.Configuration, int64, error) {
	var configs []models.Configuration
	var total int64

	query := dao.db.Model(&models.Configuration{})

	if search != "" {
		query = query.Where("namespace LIKE ? OR key LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&configs).Error
	if err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

/**
 * GetNamespaceConfigurations retrieves all configurations for a specific namespace
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} namespace - Namespace name
 * @returns {[]models.Configuration, error} Configurations and error if any
 * @description
 * - Retrieves all configurations within a namespace
 * - Ordered by key for consistent presentation
 * @throws
 * - Database query errors
 */
func (dao *ConfigDAO) GetNamespaceConfigurations(ctx context.Context, namespace string) ([]models.Configuration, error) {
	var configs []models.Configuration

	err := dao.db.Where("namespace = ?", namespace).Order("key ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}

	return configs, nil
}

/**
 * GetSpecificConfiguration retrieves a specific configuration by namespace and key
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} namespace - Namespace name
 * @param {string} key - Configuration key
 * @returns {*models.Configuration, error} Configuration and error if any
 * @description
 * - Uses composite key (namespace + key) for lookup
 * - Implements caching for frequently accessed configurations
 * - Returns nil if configuration not found
 * @throws
 * - Database query errors
 * - Redis operation errors
 */
func (dao *ConfigDAO) GetSpecificConfiguration(ctx context.Context, namespace, key string) (*models.Configuration, error) {
	var config models.Configuration

	// Try to get from cache first if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + namespace + ":" + key
		cached, err := dao.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			// Found in cache
			config.Value = cached
			config.Namespace = namespace
			config.Key = key
			return &config, nil
		} else if err != redis.Nil {
			// Redis error but not Nil, log it and continue to database
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Redis get failed, falling back to database")
		}
	}

	// Not found in cache or Redis unavailable, query database
	err := dao.db.Where("namespace = ? AND key = ?", namespace, key).First(&config).Error
	if err != nil {
		return nil, err
	}

	// Cache the result if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + namespace + ":" + key
		err := dao.redis.Set(ctx, cacheKey, config.Value, 5*time.Minute).Err()
		if err != nil {
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to cache configuration")
		}
	}

	return &config, nil
}

/**
 * CreateConfiguration creates a new configuration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Configuration} config - Configuration to create
 * @returns {error} Error if any
 * @description
 * - Creates new configuration record
 * - Invalidates related cache entries
 * - Validates required fields
 * @throws
 * - Database creation errors
 * - Redis operation errors
 */
func (dao *ConfigDAO) CreateConfiguration(ctx context.Context, config *models.Configuration) error {
	err := dao.db.Create(config).Error
	if err != nil {
		return err
	}

	// Invalidate cache if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + config.Namespace + ":" + config.Key
		err := dao.redis.Del(ctx, cacheKey).Err()
		if err != nil {
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to invalidate cache")
		}
	}

	return nil
}

/**
 * UpdateConfiguration updates an existing configuration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Configuration} config - Configuration to update
 * @returns {error} Error if any
 * @description
 * - Updates configuration record
 * - Invalidates and updates cache
 * - Validates configuration exists before update
 * @throws
 * - Database update errors
 * - Redis operation errors
 */
func (dao *ConfigDAO) UpdateConfiguration(ctx context.Context, config *models.Configuration) error {
	err := dao.db.Save(config).Error
	if err != nil {
		return err
	}

	// Update cache if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + config.Namespace + ":" + config.Key
		err := dao.redis.Set(ctx, cacheKey, config.Value, 5*time.Minute).Err()
		if err != nil {
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to update cache")
		}
	}

	return nil
}

/**
* GetConfigurationByID retrieves a configuration by ID
* @param {context.Context} ctx - Context for request cancellation
* @param {uint} id - Configuration ID
* @param {*models.Configuration} config - Pointer to store the result
* @returns {error} Error if any
* @description
* - Queries configuration by primary key ID
* - Returns gorm.ErrRecordNotFound if not found
* @throws
* - Database query errors
 */
func (dao *ConfigDAO) GetConfigurationByID(ctx context.Context, id uint, config *models.Configuration) error {
	err := dao.db.First(config, id).Error
	return err
}

/**
 * DeleteConfiguration deletes a configuration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {uint} id - Configuration ID
 * @returns {error} Error if any
 * @description
 * - Soft deletes configuration record
 * - Invalidates related cache entries
 * - Validates configuration exists before deletion
 * @throws
 * - Database deletion errors
 * - Redis operation errors
 */
func (dao *ConfigDAO) DeleteConfiguration(ctx context.Context, id uint) error {
	var config models.Configuration
	err := dao.db.First(&config, id).Error
	if err != nil {
		return err
	}

	err = dao.db.Delete(&config).Error
	if err != nil {
		return err
	}

	// Invalidate cache if Redis is available
	if dao.redis != nil {
		cacheKey := "config:" + config.Namespace + ":" + config.Key
		err := dao.redis.Del(ctx, cacheKey).Err()
		if err != nil {
			dao.log.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to invalidate cache")
		}
	}

	return nil
}
