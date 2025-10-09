package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/services"
)

/**
 * FeedbackController handles HTTP requests for feedback operations
 * @description
 * - Implements RESTful API endpoints for feedback management
 * - Handles request validation and response formatting
 * - Integrates with FeedbackService for business logic
 */
type FeedbackController struct {
	feedbackService *services.FeedbackService
	log             *logrus.Logger
}

/**
 * NewFeedbackController creates a new FeedbackController instance
 * @param {logrus.Logger} log - Logger instance
 * @returns {*FeedbackController} New FeedbackController instance
 */
func NewFeedbackController(log *logrus.Logger) *FeedbackController {
	// Initialize DAOs and services here
	feedbackService := services.NewFeedbackService(nil, log) // Will be properly initialized later

	return &FeedbackController{
		feedbackService: feedbackService,
		log:             log,
	}
}

// PostCompletionFeedback handles POST /feedbacks/completion request
// @Summary Create completion feedback
// @Description Create a new completion feedback record
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedback body map[string]interface{} true "Feedback data"
// @Success 201 {object} map[string]interface{} "Created feedback"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/completion [post]
func (fc *FeedbackController) PostCompletionFeedback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create completion feedback
	feedback, err := fc.feedbackService.CreateCompletionFeedback(c.Request.Context(), data)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Completion feedback created successfully",
		"data":    feedback,
	})
}

// PostBatchCompletionFeedback handles POST /feedbacks/completions request
// @Summary Create batch completion feedbacks
// @Description Create multiple completion feedback records in batch
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedbacks body []map[string]interface{} true "List of feedback data"
// @Success 201 {object} map[string]interface{} "Batch creation result"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/completions [post]
func (fc *FeedbackController) PostBatchCompletionFeedback(c *gin.Context) {
	var dataList []map[string]interface{}
	if err := c.ShouldBindJSON(&dataList); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create batch completion feedbacks
	count, err := fc.feedbackService.CreateBatchCompletionFeedback(c.Request.Context(), dataList)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Batch completion feedbacks created successfully",
		"data": map[string]interface{}{
			"created_count": count,
		},
	})
}

// PostCopyCodeFeedback handles POST /feedbacks/copy_code request
// @Summary Create copy code feedback
// @Description Create a new copy code feedback record
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedback body map[string]interface{} true "Feedback data"
// @Success 201 {object} map[string]interface{} "Created feedback"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/copy_code [post]
func (fc *FeedbackController) PostCopyCodeFeedback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create copy code feedback
	feedback, err := fc.feedbackService.CreateCopyCodeFeedback(c.Request.Context(), data)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Copy code feedback created successfully",
		"data":    feedback,
	})
}

// PostEvaluateFeedback handles POST /feedbacks/evaluate request
// @Summary Create evaluation feedback
// @Description Create a new evaluation feedback record (like/dislike)
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedback body map[string]interface{} true "Feedback data"
// @Success 201 {object} map[string]interface{} "Created feedback"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/evaluate [post]
func (fc *FeedbackController) PostEvaluateFeedback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create evaluation feedback
	feedback, err := fc.feedbackService.CreateEvaluateFeedback(c.Request.Context(), data)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Evaluation feedback created successfully",
		"data":    feedback,
	})
}

// PostUseCodeFeedback handles POST /feedbacks/use_code request
// @Summary Create use code feedback
// @Description Create a new use code feedback record
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedback body map[string]interface{} true "Feedback data"
// @Success 201 {object} map[string]interface{} "Created feedback"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/use_code [post]
func (fc *FeedbackController) PostUseCodeFeedback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create use code feedback
	feedback, err := fc.feedbackService.CreateUseCodeFeedback(c.Request.Context(), data)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Use code feedback created successfully",
		"data":    feedback,
	})
}

// PostIssueFeedback handles POST /feedbacks/issue request
// @Summary Create issue feedback
// @Description Create a new issue feedback record
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedback body map[string]interface{} true "Feedback data"
// @Success 201 {object} map[string]interface{} "Created feedback"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/issue [post]
func (fc *FeedbackController) PostIssueFeedback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create issue feedback
	feedback, err := fc.feedbackService.CreateIssueFeedback(c.Request.Context(), data)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Issue feedback created successfully",
		"data":    feedback,
	})
}

// PostErrorFeedback handles POST /feedbacks/error request
// @Summary Create error feedback
// @Description Create a new error feedback record
// @Tags Feedback
// @Accept json
// @Produce json
// @Param feedback body map[string]interface{} true "Feedback data"
// @Success 201 {object} map[string]interface{} "Created feedback"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-manager/api/v1/feedbacks/error [post]
func (fc *FeedbackController) PostErrorFeedback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		fc.handleError(c, &services.ValidationError{Field: "body", Message: "Invalid request body"})
		return
	}

	// Create error feedback
	feedback, err := fc.feedbackService.CreateErrorFeedback(c.Request.Context(), data)
	if err != nil {
		fc.handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"code":    "success",
		"message": "Error feedback created successfully",
		"data":    feedback,
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
func (fc *FeedbackController) handleError(c *gin.Context, err error) {
	// Log error
	fc.log.WithError(err).Error("Request processing failed")

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
 * SetFeedbackService sets the feedback service (used for dependency injection)
 * @param {services.FeedbackService} feedbackService - Feedback service instance
 * @description
 * - Allows setting the feedback service after controller creation
 * - Used for proper dependency injection
 */
func (fc *FeedbackController) SetFeedbackService(feedbackService *services.FeedbackService) {
	fc.feedbackService = feedbackService
}
