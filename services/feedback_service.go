package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/dao"
	"github.com/zgsm-ai/client-manager/models"
)

/**
 * FeedbackService handles business logic for feedback operations
 * @description
 * - Implements feedback processing business rules
 * - Validates feedback data
 * - Handles different feedback types
 */
type FeedbackService struct {
	feedbackDAO *dao.FeedbackDAO
	log         *logrus.Logger
}

/**
 * NewFeedbackService creates a new FeedbackService instance
 * @param {dao.FeedbackDAO} feedbackDAO - Feedback data access object
 * @param {logrus.Logger} log - Logger instance
 * @returns {*FeedbackService} New FeedbackService instance
 */
func NewFeedbackService(feedbackDAO *dao.FeedbackDAO, log *logrus.Logger) *FeedbackService {
	return &FeedbackService{
		feedbackDAO: feedbackDAO,
		log:         log,
	}
}

/**
 * CreateCompletionFeedback creates a completion feedback
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Feedback data
 * @returns {*models.Feedback, error} Created feedback and error if any
 * @description
 * - Validates completion feedback data
 * - Creates feedback record
 * - Logs feedback creation
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 */
func (s *FeedbackService) CreateCompletionFeedback(ctx context.Context, data map[string]interface{}) (*models.Feedback, error) {
	// Validate and extract feedback data
	feedback, err := s.validateAndExtractFeedback(data, "completion")
	if err != nil {
		return nil, err
	}

	// Create feedback
	err = s.feedbackDAO.CreateCompletionFeedback(ctx, feedback)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"type":            "completion",
			"conversation_id": feedback.ConversationID,
		}).Error("Failed to create completion feedback")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"type":            "completion",
		"conversation_id": feedback.ConversationID,
	}).Info("Completion feedback created successfully")

	return feedback, nil
}

/**
 * CreateBatchCompletionFeedback creates multiple completion feedbacks
 * @param {context.Context} ctx - Context for request cancellation
 * @param {[]map[string]interface{}} dataList - List of feedback data
 * @returns {int, error} Number of created feedbacks and error if any
 * @description
 * - Validates each feedback data
 * - Creates feedback records in batch
 * - Logs batch feedback creation
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 */
func (s *FeedbackService) CreateBatchCompletionFeedback(ctx context.Context, dataList []map[string]interface{}) (int, error) {
	feedbacks := make([]models.Feedback, 0, len(dataList))

	for _, data := range dataList {
		feedback, err := s.validateAndExtractFeedback(data, "completion")
		if err != nil {
			s.log.WithError(err).Error("Failed to validate feedback data in batch")
			continue
		}
		feedbacks = append(feedbacks, *feedback)
	}

	if len(feedbacks) == 0 {
		return 0, &ValidationError{Field: "data", Message: "no valid feedback data provided"}
	}

	// Create batch feedbacks
	err := s.feedbackDAO.CreateBatchCompletionFeedback(ctx, feedbacks)
	if err != nil {
		s.log.WithError(err).WithField("count", len(feedbacks)).Error("Failed to create batch completion feedbacks")
		return 0, err
	}

	s.log.WithField("count", len(feedbacks)).Info("Batch completion feedbacks created successfully")

	return len(feedbacks), nil
}

/**
 * CreateCopyCodeFeedback creates a copy code feedback
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Feedback data
 * @returns {*models.Feedback, error} Created feedback and error if any
 * @description
 * - Validates copy code feedback data
 * - Creates feedback record
 * - Logs feedback creation
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 */
func (s *FeedbackService) CreateCopyCodeFeedback(ctx context.Context, data map[string]interface{}) (*models.Feedback, error) {
	// Validate and extract feedback data
	feedback, err := s.validateAndExtractFeedback(data, "copy_code")
	if err != nil {
		return nil, err
	}

	// Create feedback
	err = s.feedbackDAO.CreateCopyCodeFeedback(ctx, feedback)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"type":            "copy_code",
			"conversation_id": feedback.ConversationID,
		}).Error("Failed to create copy code feedback")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"type":            "copy_code",
		"conversation_id": feedback.ConversationID,
	}).Info("Copy code feedback created successfully")

	return feedback, nil
}

/**
 * CreateEvaluateFeedback creates an evaluation feedback
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Feedback data
 * @returns {*models.Feedback, error} Created feedback and error if any
 * @description
 * - Validates evaluation feedback data
 * - Validates evaluation type (like/dislike)
 * - Creates feedback record
 * - Logs feedback creation
 * @throws
 * - Validation errors for invalid data or evaluation type
 * - Database creation errors
 */
func (s *FeedbackService) CreateEvaluateFeedback(ctx context.Context, data map[string]interface{}) (*models.Feedback, error) {
	// Validate conversation ID
	conversationID, ok := data["conversation_id"].(string)
	if !ok || conversationID == "" {
		return nil, &ValidationError{Field: "conversation_id", Message: "conversation_id is required and must be a string"}
	}

	// Validate evaluation type
	evaluationType, ok := data["evaluation_type"].(string)
	if !ok || (evaluationType != "like" && evaluationType != "dislike") {
		return nil, &ValidationError{Field: "evaluation_type", Message: "evaluation_type is required and must be 'like' or 'dislike'"}
	}

	// Extract other fields
	userID, _ := data["user_id"].(string)
	content := evaluationType // Use evaluation type as content
	metadata, _ := data["metadata"].(string)

	// Create feedback
	feedback := &models.Feedback{
		ConversationID: conversationID,
		UserID:         userID,
		Content:        content,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create feedback
	err := s.feedbackDAO.CreateEvaluateFeedback(ctx, feedback)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"type":            "evaluate",
			"conversation_id": feedback.ConversationID,
			"evaluation_type": evaluationType,
		}).Error("Failed to create evaluate feedback")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"type":            "evaluate",
		"conversation_id": feedback.ConversationID,
		"evaluation_type": evaluationType,
	}).Info("Evaluate feedback created successfully")

	return feedback, nil
}

/**
 * CreateUseCodeFeedback creates a use code feedback
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Feedback data
 * @returns {*models.Feedback, error} Created feedback and error if any
 * @description
 * - Validates use code feedback data
 * - Validates action type
 * - Creates feedback record
 * - Logs feedback creation
 * @throws
 * - Validation errors for invalid data or action type
 * - Database creation errors
 */
func (s *FeedbackService) CreateUseCodeFeedback(ctx context.Context, data map[string]interface{}) (*models.Feedback, error) {
	// Validate conversation ID
	conversationID, ok := data["conversation_id"].(string)
	if !ok || conversationID == "" {
		return nil, &ValidationError{Field: "conversation_id", Message: "conversation_id is required and must be a string"}
	}

	// Validate action type
	actionType, ok := data["action_type"].(string)
	if !ok {
		return nil, &ValidationError{Field: "action_type", Message: "action_type is required and must be a string"}
	}

	// Extract other fields
	userID, _ := data["user_id"].(string)
	content := actionType // Use action type as content
	metadata, _ := data["metadata"].(string)

	// Create feedback
	feedback := &models.Feedback{
		ConversationID: conversationID,
		UserID:         userID,
		Content:        content,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create feedback
	err := s.feedbackDAO.CreateUseCodeFeedback(ctx, feedback)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"type":            "use_code",
			"conversation_id": feedback.ConversationID,
			"action_type":     actionType,
		}).Error("Failed to create use code feedback")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"type":            "use_code",
		"conversation_id": feedback.ConversationID,
		"action_type":     actionType,
	}).Info("Use code feedback created successfully")

	return feedback, nil
}

/**
 * CreateIssueFeedback creates an issue feedback
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Feedback data
 * @returns {*models.Feedback, error} Created feedback and error if any
 * @description
 * - Validates issue feedback data
 * - Validates issue description
 * - Creates feedback record
 * - Logs feedback creation
 * @throws
 * - Validation errors for invalid data or missing description
 * - Database creation errors
 */
func (s *FeedbackService) CreateIssueFeedback(ctx context.Context, data map[string]interface{}) (*models.Feedback, error) {
	// Validate description
	description, ok := data["description"].(string)
	if !ok || description == "" {
		return nil, &ValidationError{Field: "description", Message: "description is required and must be a string"}
	}

	// Extract other fields
	userID, _ := data["user_id"].(string)
	issueType, _ := data["issue_type"].(string)
	metadata, _ := data["metadata"].(string)

	// Create feedback
	feedback := &models.Feedback{
		UserID:    userID,
		Content:   description,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add issue type to metadata if provided
	if issueType != "" {
		if metadata == "" {
			metadata = `{"issue_type":"` + issueType + `"}`
		} else {
			metadata = `{"issue_type":"` + issueType + `",` + metadata[1:]
		}
		feedback.Metadata = metadata
	}

	// Create feedback
	err := s.feedbackDAO.CreateIssueFeedback(ctx, feedback)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"type":       "issue",
			"issue_type": issueType,
			"user_id":    userID,
		}).Error("Failed to create issue feedback")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"type":       "issue",
		"issue_type": issueType,
		"user_id":    userID,
	}).Info("Issue feedback created successfully")

	return feedback, nil
}

/**
 * CreateErrorFeedback creates an error feedback
 * @param {context.Context} ctx - Context for request cancellation
 * @param {map[string]interface{}} data - Feedback data
 * @returns {*models.Feedback, error} Created feedback and error if any
 * @description
 * - Validates error feedback data
 * - Creates feedback record
 * - Logs feedback creation
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 */
func (s *FeedbackService) CreateErrorFeedback(ctx context.Context, data map[string]interface{}) (*models.Feedback, error) {
	// Validate and extract feedback data
	feedback, err := s.validateAndExtractFeedback(data, "error")
	if err != nil {
		return nil, err
	}

	// Create feedback
	err = s.feedbackDAO.CreateErrorFeedback(ctx, feedback)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"type":    "error",
			"user_id": feedback.UserID,
		}).Error("Failed to create error feedback")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"type":    "error",
		"user_id": feedback.UserID,
	}).Info("Error feedback created successfully")

	return feedback, nil
}

/**
 * GetFeedbackStats retrieves feedback statistics
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} startDate - Start date for analysis
 * @param {string} endDate - End date for analysis
 * @returns {map[string]interface{}, error} Statistics data and error if any
 * @description
 * - Validates date parameters
 * - Retrieves feedback statistics
 * - Returns aggregated data
 * @throws
 * - Validation errors for invalid dates
 * - Database query errors
 */
func (s *FeedbackService) GetFeedbackStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	// Validate date parameters
	if startDate == "" {
		return nil, &ValidationError{Field: "start_date", Message: "start_date is required"}
	}
	if endDate == "" {
		return nil, &ValidationError{Field: "end_date", Message: "end_date is required"}
	}

	// Get statistics
	stats, err := s.feedbackDAO.GetFeedbackStats(ctx, startDate, endDate)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"start_date": startDate,
			"end_date":   endDate,
		}).Error("Failed to get feedback statistics")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{
		"start_date": startDate,
		"end_date":   endDate,
	}).Info("Feedback statistics retrieved successfully")

	return stats, nil
}

/**
 * validateAndExtractFeedback validates and extracts common feedback data
 * @param {map[string]interface{}} data - Feedback data
 * @param {string} feedbackType - Type of feedback
 * @returns {*models.Feedback, error} Validated feedback and error if any
 * @description
 * - Validates common feedback fields
 * - Extracts feedback data
 * - Creates feedback object
 * @throws
 * - Validation errors for missing required fields
 */
func (s *FeedbackService) validateAndExtractFeedback(data map[string]interface{}, feedbackType string) (*models.Feedback, error) {
	conversationID, _ := data["conversation_id"].(string)
	userID, _ := data["user_id"].(string)
	content, _ := data["content"].(string)
	metadata, _ := data["metadata"].(string)

	// Create feedback
	feedback := &models.Feedback{
		ConversationID: conversationID,
		UserID:         userID,
		Content:        content,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return feedback, nil
}
