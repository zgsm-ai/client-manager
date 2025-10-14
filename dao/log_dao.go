package dao

import (
	"context"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zgsm-ai/client-manager/models"
)

/**
 * LogDAO handles data access operations for log data
 * @description
 * - Provides CRUD operations for log data
 * - Supports client and user based log filtering
 * - Implements batch operations for performance optimization
 */
type LogDAO struct {
	db  *gorm.DB
	log *logrus.Logger
}

/**
 * NewLogDAO creates a new LogDAO instance
 * @param {gorm.DB} db - Database connection
 * @param {logrus.Logger} log - Logger instance
 * @returns {*LogDAO} New LogDAO instance
 */
func NewLogDAO(db *gorm.DB, log *logrus.Logger) *LogDAO {
	return &LogDAO{
		db:  db,
		log: log,
	}
}

func (dao *LogDAO) Upsert(ctx context.Context, log *models.Log) error {
	// 使用 ClientID 和 FileName 作为唯一标识
	var existingLog models.Log
	result := dao.db.Where("client_id = ? AND file_name = ?", log.ClientID, log.FileName).FirstOrInit(&existingLog)

	// 如果记录存在，更新记录
	if result.RowsAffected > 0 {
		return dao.db.Model(&existingLog).Updates(log).Error
	}

	// 如果记录不存在，创建新记录
	return dao.db.Create(log).Error
}

/**
 * ListLogs retrieves logs for a specific client
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} clientID - Client identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {[]models.Log, int64, error} List of logs, total count, and error
 * @description
 * - Retrieves log records filtered by client ID
 * - Supports pagination for large datasets
 * - Returns total count for frontend pagination
 * @throws
 * - Database query errors
 */
func (dao *LogDAO) ListLogs(ctx context.Context, clientID, userID, fileName string, page, pageSize int) ([]models.Log, int64, error) {
	var logs []models.Log
	var total int64

	query := dao.db.Model(&models.Log{})

	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if fileName != "" {
		query = query.Where("file_name = ?", fileName)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

/**
 * DeleteOldLogs deletes logs older than specified date
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} beforeDate - Delete logs before this date
 * @returns {int64, error} Number of deleted records and error if any
 * @description
 * - Performs cleanup of old log records
 * - Uses soft delete for data safety
 * - Returns count of deleted records
 * @throws
 * - Database deletion errors
 */
func (dao *LogDAO) DeleteOldLogs(ctx context.Context, beforeDate string) (int64, error) {
	result := dao.db.Where("updated_at < ?", beforeDate).Delete(&models.Log{})
	return result.RowsAffected, result.Error
}
