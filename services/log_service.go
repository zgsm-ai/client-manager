package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/dao"
	"github.com/zgsm-ai/client-manager/models"
)

/**
 * LogService handles business logic for log operations
 * @description
 * - Implements log processing business rules
 * - Validates log data
 * - Handles different log types
 */
type LogService struct {
	logDAO *dao.LogDAO
	log    *logrus.Logger
}

/**
 * NewLogService creates a new LogService instance
 * @param {dao.LogDAO} logDAO - Log data access object
 * @param {logrus.Logger} log - Logger instance
 * @returns {*LogService} New LogService instance
 */
func NewLogService(logDAO *dao.LogDAO, log *logrus.Logger) *LogService {
	return &LogService{
		logDAO: logDAO,
		log:    log,
	}
}

/**
 * CreateLog creates a new log record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Log data
 * @returns {*models.Log, error} Created log and error if any
 * @description
 * - Validates log data
 * - Creates log record
 * - Logs creation operation
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 */
func (s *LogService) CreateLog(ctx context.Context, data map[string]interface{}) (*models.Log, error) {
	// Validate and extract log data
	log, err := s.validateAndExtractLog(data)
	if err != nil {
		return nil, err
	}

	// Create log
	err = s.logDAO.CreateLog(ctx, log)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id":   log.ClientID,
			"user_id":     log.UserID,
			"module_name": log.ModuleName,
		}).Error("Failed to create log")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"client_id":   log.ClientID,
		"user_id":     log.UserID,
		"module_name": log.ModuleName,
	}).Info("Log created successfully")

	return log, nil
}

/**
 * GetLogsByClient retrieves logs for a specific client
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} clientID - Client identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {map[string]interface{}, error} Response containing logs and pagination info
 * @description
 * - Validates client ID and pagination parameters
 * - Retrieves logs from database
 * - Returns structured response with pagination metadata
 * @throws
 * - Validation errors for invalid parameters
 * - Database query errors
 */
func (s *LogService) GetLogsByClient(ctx context.Context, clientID string, page, pageSize int) (map[string]interface{}, error) {
	// Validate client ID
	if clientID == "" {
		return nil, &ValidationError{Field: "client_id", Message: "client_id is required"}
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get logs
	logs, total, err := s.logDAO.GetLogsByClient(ctx, clientID, page, pageSize)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id": clientID,
			"page":      page,
			"page_size": pageSize,
		}).Error("Failed to get logs by client")
		return nil, err
	}

	// Prepare response
	response := map[string]interface{}{
		"data": logs,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	}

	s.log.WithFields(logrus.Fields{
		"client_id": clientID,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}).Info("Logs retrieved successfully by client")

	return response, nil
}

/**
 * GetLogsByUser retrieves logs for a specific user
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} userID - User identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {map[string]interface{}, error} Response containing logs and pagination info
 * @description
 * - Validates user ID and pagination parameters
 * - Retrieves logs from database
 * - Returns structured response with pagination metadata
 * @throws
 * - Validation errors for invalid parameters
 * - Database query errors
 */
func (s *LogService) GetLogsByUser(ctx context.Context, userID string, page, pageSize int) (map[string]interface{}, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get logs
	logs, total, err := s.logDAO.GetLogsByUser(ctx, userID, page, pageSize)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"user_id":   userID,
			"page":      page,
			"page_size": pageSize,
		}).Error("Failed to get logs by user")
		return nil, err
	}

	// Prepare response
	response := map[string]interface{}{
		"data": logs,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	}

	s.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}).Info("Logs retrieved successfully by user")

	return response, nil
}

/**
 * GetLogStats retrieves log statistics
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} startDate - Start date for analysis
 * @param {string} endDate - End date for analysis
 * @returns {map[string]interface{}, error} Statistics data and error if any
 * @description
 * - Validates date parameters
 * - Retrieves log statistics
 * - Returns aggregated data
 * @throws
 * - Validation errors for invalid dates
 * - Database query errors
 */
func (s *LogService) GetLogStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	// Validate date parameters
	if startDate == "" {
		return nil, &ValidationError{Field: "start_date", Message: "start_date is required"}
	}
	if endDate == "" {
		return nil, &ValidationError{Field: "end_date", Message: "end_date is required"}
	}

	// Get statistics
	stats, err := s.logDAO.GetLogStats(ctx, startDate, endDate)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"start_date": startDate,
			"end_date":   endDate,
		}).Error("Failed to get log statistics")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"start_date": startDate,
		"end_date":   endDate,
	}).Info("Log statistics retrieved successfully")

	return stats, nil
}

/**
 * DeleteOldLogs deletes logs older than specified date
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} beforeDate - Delete logs before this date
 * @returns {int64, error} Number of deleted records and error if any
 * @description
 * - Validates date parameter
 * - Performs cleanup of old log records
 * - Returns count of deleted records
 * @throws
 * - Validation errors for invalid date
 * - Database deletion errors
 */
func (s *LogService) DeleteOldLogs(ctx context.Context, beforeDate string) (int64, error) {
	// Validate date parameter
	if beforeDate == "" {
		return 0, &ValidationError{Field: "before_date", Message: "before_date is required"}
	}

	// Delete old logs
	count, err := s.logDAO.DeleteOldLogs(ctx, beforeDate)
	if err != nil {
		s.log.WithError(err).WithField("before_date", beforeDate).Error("Failed to delete old logs")
		return 0, err
	}

	s.log.WithFields(logrus.Fields{
		"before_date":   beforeDate,
		"deleted_count": count,
	}).Info("Old logs deleted successfully")

	return count, nil
}

/**
 * GetLogSessions retrieves log sessions based on start/end flags
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} clientID - Client identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {map[string]interface{}, error} Response containing session logs and pagination info
 * @description
 * - Validates client ID and pagination parameters
 * - Retrieves session logs from database
 * - Returns structured response with pagination metadata
 * @throws
 * - Validation errors for invalid parameters
 * - Database query errors
 */
func (s *LogService) GetLogSessions(ctx context.Context, clientID string, page, pageSize int) (map[string]interface{}, error) {
	// Validate client ID
	if clientID == "" {
		return nil, &ValidationError{Field: "client_id", Message: "client_id is required"}
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get session logs
	logs, total, err := s.logDAO.GetLogSessions(ctx, clientID, page, pageSize)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id": clientID,
			"page":      page,
			"page_size": pageSize,
		}).Error("Failed to get log sessions")
		return nil, err
	}

	// Prepare response
	response := map[string]interface{}{
		"data": logs,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	}

	s.log.WithFields(logrus.Fields{
		"client_id": clientID,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}).Info("Log sessions retrieved successfully")

	return response, nil
}

/**
 * validateAndExtractLog validates and extracts log data
 * @param {map[string]interface{}} data - Log data
 * @returns {*models.Log, error} Validated log and error if any
 * @description
 * - Validates required log fields
 * - Extracts log data
 * - Creates log object
 * @throws
 * - Validation errors for missing required fields
 */
func (s *LogService) validateAndExtractLog(data map[string]interface{}) (*models.Log, error) {
	// Validate required fields
	clientID, ok := data["client_id"].(string)
	if !ok || clientID == "" {
		return nil, &ValidationError{Field: "client_id", Message: "client_id is required and must be a string"}
	}

	moduleName, ok := data["module_name"].(string)
	if !ok || moduleName == "" {
		return nil, &ValidationError{Field: "module_name", Message: "module_name is required and must be a string"}
	}

	// Extract optional fields
	userID, _ := data["user_id"].(string)
	logContent, _ := data["log_content"].(string)
	startFlag, _ := data["start_flag"].(bool)
	endFlag, _ := data["end_flag"].(bool)

	// Create log
	log := &models.Log{
		ClientID:   clientID,
		UserID:     userID,
		ModuleName: moduleName,
		LogContent: logContent,
		StartFlag:  startFlag,
		EndFlag:    endFlag,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return log, nil
}
