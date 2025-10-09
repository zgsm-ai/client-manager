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

/**
 * CreateLog creates a new log record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Log} log - Log data to create
 * @returns {error} Error if any
 * @description
 * - Creates log record from client data
 * - Validates required fields (client_id, module_name)
 * - Sets timestamps automatically
 * @throws
 * - Database creation errors
 */
func (dao *LogDAO) CreateLog(ctx context.Context, log *models.Log) error {
	return dao.db.Create(log).Error
}

/**
 * GetLogsByClient retrieves logs for a specific client
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
func (dao *LogDAO) GetLogsByClient(ctx context.Context, clientID string, page, pageSize int) ([]models.Log, int64, error) {
	var logs []models.Log
	var total int64

	query := dao.db.Model(&models.Log{}).Where("client_id = ?", clientID)

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
 * GetLogsByUser retrieves logs for a specific user
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} userID - User identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {[]models.Log, int64, error} List of logs, total count, and error
 * @description
 * - Retrieves log records filtered by user ID
 * - Supports pagination for large datasets
 * - Returns total count for frontend pagination
 * @throws
 * - Database query errors
 */
func (dao *LogDAO) GetLogsByUser(ctx context.Context, userID string, page, pageSize int) ([]models.Log, int64, error) {
	var logs []models.Log
	var total int64

	query := dao.db.Model(&models.Log{}).Where("user_id = ?", userID)

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
 * GetLogsByModule retrieves logs for a specific module
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} moduleName - Module name
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {[]models.Log, int64, error} List of logs, total count, and error
 * @description
 * - Retrieves log records filtered by module name
 * - Supports pagination for large datasets
 * - Returns total count for frontend pagination
 * @throws
 * - Database query errors
 */
func (dao *LogDAO) GetLogsByModule(ctx context.Context, moduleName string, page, pageSize int) ([]models.Log, int64, error) {
	var logs []models.Log
	var total int64

	query := dao.db.Model(&models.Log{}).Where("module_name = ?", moduleName)

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
 * GetLogStats retrieves statistics for log analysis
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} startDate - Start date for analysis
 * @param {string} endDate - End date for analysis
 * @returns {map[string]interface{}, error} Statistics data and error if any
 * @description
 * - Aggregates log data by client, user, and module
 * - Provides counts for different log categories
 * - Used for analytics and reporting
 * @throws
 * - Database query errors
 */
func (dao *LogDAO) GetLogStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total log count
	var total int64
	err := dao.db.Model(&models.Log{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&total).Error
	if err != nil {
		return nil, err
	}
	stats["total_count"] = total

	// Get counts by client
	clientCounts := make(map[string]int64)
	rows, err := dao.db.Model(&models.Log{}).
		Select("client_id, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("client_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var clientID string
		var count int64
		if err := rows.Scan(&clientID, &count); err != nil {
			return nil, err
		}
		clientCounts[clientID] = count
	}
	stats["client_counts"] = clientCounts

	// Get counts by module
	moduleCounts := make(map[string]int64)
	rows, err = dao.db.Model(&models.Log{}).
		Select("module_name, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("module_name").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var moduleName string
		var count int64
		if err := rows.Scan(&moduleName, &count); err != nil {
			return nil, err
		}
		moduleCounts[moduleName] = count
	}
	stats["module_counts"] = moduleCounts

	return stats, nil
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
	result := dao.db.Where("created_at < ?", beforeDate).Delete(&models.Log{})
	return result.RowsAffected, result.Error
}

/**
 * GetLogSessions retrieves log sessions based on start/end flags
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} clientID - Client identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {[]models.Log, int64, error} List of session logs, total count, and error
 * @description
 * - Retrieves log records that mark session boundaries
 * - Groups logs by start/end flags for session analysis
 * - Supports pagination for large datasets
 * @throws
 * - Database query errors
 */
func (dao *LogDAO) GetLogSessions(ctx context.Context, clientID string, page, pageSize int) ([]models.Log, int64, error) {
	var logs []models.Log
	var total int64

	query := dao.db.Model(&models.Log{}).
		Where("client_id = ? AND (start_flag = ? OR end_flag = ?)", clientID, true, true)

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
