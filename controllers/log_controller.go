package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/services"
)

/**
 * LogController handles HTTP requests for log operations
 * @description
 * - Implements RESTful API endpoints for log management
 * - Handles request validation and response formatting
 * - Integrates with LogService for business logic
 */
type LogController struct {
	logService *services.LogService
	log        *logrus.Logger
}

/**
 * NewLogController creates a new LogController instance
 * @param {logrus.Logger} log - Logger instance
 * @returns {*LogController} New LogController instance
 */
func NewLogController(log *logrus.Logger) *LogController {
	// Initialize DAOs and services here
	logService := services.NewLogService(nil, log) // Will be properly initialized later

	return &LogController{
		logService: logService,
		log:        log,
	}
}
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return ""
	}
}

func getUserId(header http.Header) string {
	// Get Authorization header
	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if the header has Bearer prefix
	tokenString := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = authHeader[7:] // Remove "Bearer " prefix
	}

	// Parse token without verification (for now)
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return ""
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Extract user_id from claims
		if userID, exists := claims["id"]; exists {
			// Set user_id in request header
			return toString(userID)
		}
	}
	return ""
}

// PostLog handles POST /logs request
// @Summary Create log
// @Description Create a new log record
// @Tags Log
// @Accept json
// @Produce json
// @Param log body map[string]interface{} true "Log data"
// @Success 201 {object} map[string]interface{} "Created log"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/logs [post]
func (lc *LogController) PostLog(c *gin.Context) {
	// 获取上传的文件
	fileHead, err := c.FormFile("logfile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := getUserId(c.Request.Header)
	// 创建目标文件路径
	destPath := filepath.Join("/data", userId, fileHead.Filename)
	file, err := fileHead.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	// 打开目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer destFile.Close()
	// 将上传的文件内容复制到目标文件
	if _, err := io.Copy(destFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": fmt.Sprintf("File uploaded successfully: %s", destPath),
	})
}

// GetLogsByClient handles GET /logs/client/{client_id} request
// @Summary Get logs by client
// @Description Retrieve logs for a specific client with pagination
// @Tags Log
// @Accept json
// @Produce json
// @Param client_id path string true "Client ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} map[string]interface{} "Logs list with pagination"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/logs/client/{client_id} [get]
func (lc *LogController) GetLogsByClient(c *gin.Context) {
	// Get path parameter
	clientID := c.Param("client_id")

	// Get and validate pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get logs by client
	response, err := lc.logService.GetLogsByClient(c.Request.Context(), clientID, page, pageSize)
	if err != nil {
		lc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "Logs retrieved successfully by client",
		"data":    response,
	})
}

// GetLogsByUser handles GET /logs/user/{user_id} request
// @Summary Get logs by user
// @Description Retrieve logs for a specific user with pagination
// @Tags Log
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} map[string]interface{} "Logs list with pagination"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/logs/user/{user_id} [get]
func (lc *LogController) GetLogsByUser(c *gin.Context) {
	// Get path parameter
	userID := c.Param("user_id")

	// Get and validate pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get logs by user
	response, err := lc.logService.GetLogsByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		lc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "Logs retrieved successfully by user",
		"data":    response,
	})
}

// GetLogStats handles GET /logs/stats request
// @Summary Get log statistics
// @Description Retrieve log statistics for a given time period
// @Tags Log
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Log statistics"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/logs/stats [get]
func (lc *LogController) GetLogStats(c *gin.Context) {
	// Get query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Get log statistics
	stats, err := lc.logService.GetLogStats(c.Request.Context(), startDate, endDate)
	if err != nil {
		lc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "Log statistics retrieved successfully",
		"data":    stats,
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
func (lc *LogController) handleError(c *gin.Context, err error) {
	// Log error
	lc.log.WithError(err).Error("Request processing failed")

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
 * SetLogService sets the log service (used for dependency injection)
 * @param {services.LogService} logService - Log service instance
 * @description
 * - Allows setting the log service after controller creation
 * - Used for proper dependency injection
 */
func (lc *LogController) SetLogService(logService *services.LogService) {
	lc.logService = logService
}
