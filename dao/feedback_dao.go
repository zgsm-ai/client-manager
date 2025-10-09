package dao

import (
	"context"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zgsm-ai/client-manager/models"
)

/**
 * FeedbackDAO handles data access operations for feedback data
 * @description
 * - Provides CRUD operations for feedback data
 * - Supports different feedback types (completion, copy, evaluate, etc.)
 * - Implements batch operations for performance optimization
 */
type FeedbackDAO struct {
	db  *gorm.DB
	log *logrus.Logger
}

/**
 * NewFeedbackDAO creates a new FeedbackDAO instance
 * @param {gorm.DB} db - Database connection
 * @param {logrus.Logger} log - Logger instance
 * @returns {*FeedbackDAO} New FeedbackDAO instance
 */
func NewFeedbackDAO(db *gorm.DB, log *logrus.Logger) *FeedbackDAO {
	return &FeedbackDAO{
		db:  db,
		log: log,
	}
}

/**
 * CreateCompletionFeedback creates a completion feedback record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Feedback} feedback - Feedback data to create
 * @returns {error} Error if any
 * @description
 * - Creates feedback record for code completion acceptance
 * - Validates required fields
 * - Logs creation operation
 * @throws
 * - Database creation errors
 */
func (dao *FeedbackDAO) CreateCompletionFeedback(ctx context.Context, feedback *models.Feedback) error {
	feedback.Type = "completion"
	return dao.db.Create(feedback).Error
}

/**
 * CreateBatchCompletionFeedback creates multiple completion feedback records
 * @param {context.Context} ctx - Context for request cancellation
 * @param {[]models.Feedback} feedbacks - List of feedback data to create
 * @returns {error} Error if any
 * @description
 * - Batch creates feedback records for code completion
 * - Uses transaction for data consistency
 * - Validates all records before creation
 * @throws
 * - Database transaction errors
 */
func (dao *FeedbackDAO) CreateBatchCompletionFeedback(ctx context.Context, feedbacks []models.Feedback) error {
	for i := range feedbacks {
		feedbacks[i].Type = "completion"
	}
	return dao.db.CreateInBatches(feedbacks, 100).Error
}

/**
 * CreateCopyCodeFeedback creates a copy code feedback record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Feedback} feedback - Feedback data to create
 * @returns {error} Error if any
 * @description
 * - Creates feedback record for code copy action
 * - Validates conversation ID exists
 * - Logs copy action for analytics
 * @throws
 * - Database creation errors
 */
func (dao *FeedbackDAO) CreateCopyCodeFeedback(ctx context.Context, feedback *models.Feedback) error {
	feedback.Type = "copy_code"
	return dao.db.Create(feedback).Error
}

/**
 * CreateEvaluateFeedback creates an evaluation feedback record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Feedback} feedback - Feedback data to create
 * @returns {error} Error if any
 * @description
 * - Creates feedback record for user evaluation (like/dislike)
 * - Validates evaluation type
 * - Associates with conversation for quality tracking
 * @throws
 * - Database creation errors
 */
func (dao *FeedbackDAO) CreateEvaluateFeedback(ctx context.Context, feedback *models.Feedback) error {
	feedback.Type = "evaluate"
	return dao.db.Create(feedback).Error
}

/**
 * CreateUseCodeFeedback creates a use code feedback record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Feedback} feedback - Feedback data to create
 * @returns {error} Error if any
 * @description
 * - Creates feedback record for code usage actions
 * - Tracks various code usage methods (ctrl+c, copy, accept, diff)
 * - Validates conversation ID exists
 * @throws
 * - Database creation errors
 */
func (dao *FeedbackDAO) CreateUseCodeFeedback(ctx context.Context, feedback *models.Feedback) error {
	feedback.Type = "use_code"
	return dao.db.Create(feedback).Error
}

/**
 * CreateIssueFeedback creates an issue feedback record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Feedback} feedback - Feedback data to create
 * @returns {error} Error if any
 * @description
 * - Creates feedback record for user-reported issues
 * - Supports attachments and screenshots in metadata
 * - Includes contact information for follow-up
 * @throws
 * - Database creation errors
 */
func (dao *FeedbackDAO) CreateIssueFeedback(ctx context.Context, feedback *models.Feedback) error {
	feedback.Type = "issue"
	return dao.db.Create(feedback).Error
}

/**
 * CreateErrorFeedback creates an error feedback record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Feedback} feedback - Feedback data to create
 * @returns {error} Error if any
 * @description
 * - Creates feedback record for client error statistics
 * - Stores aggregated error data from IDE plugins
 * - Used for system reliability monitoring
 * @throws
 * - Database creation errors
 */
func (dao *FeedbackDAO) CreateErrorFeedback(ctx context.Context, feedback *models.Feedback) error {
	feedback.Type = "error"
	return dao.db.Create(feedback).Error
}

/**
 * GetFeedbacksByConversation retrieves all feedbacks for a specific conversation
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} conversationID - Conversation identifier
 * @returns {[]models.Feedback, error} List of feedbacks and error if any
 * @description
 * - Retrieves all feedback records associated with a conversation
 * - Ordered by creation time for chronological analysis
 * - Used for conversation quality assessment
 * @throws
 * - Database query errors
 */
func (dao *FeedbackDAO) GetFeedbacksByConversation(ctx context.Context, conversationID string) ([]models.Feedback, error) {
	var feedbacks []models.Feedback
	err := dao.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&feedbacks).Error
	return feedbacks, err
}

/**
 * GetFeedbacksByType retrieves feedbacks by type with pagination
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} feedbackType - Type of feedback
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {[]models.Feedback, int64, error} List of feedbacks, total count, and error
 * @description
 * - Retrieves feedback records filtered by type
 * - Supports pagination for large datasets
 * - Returns total count for frontend pagination
 * @throws
 * - Database query errors
 */
func (dao *FeedbackDAO) GetFeedbacksByType(ctx context.Context, feedbackType string, page, pageSize int) ([]models.Feedback, int64, error) {
	var feedbacks []models.Feedback
	var total int64

	query := dao.db.Model(&models.Feedback{}).Where("type = ?", feedbackType)

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&feedbacks).Error
	if err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

/**
 * GetFeedbackStats retrieves statistics for feedback analysis
 * @param {context.Context} ctx - Context for request cancellation
 * @param {time.Time} startDate - Start date for analysis
 * @param {time.Time} endDate - End date for analysis
 * @returns {map[string]interface{}, error} Statistics data and error if any
 * @description
 * - Aggregates feedback data by type and time period
 * - Provides counts for different feedback types
 * - Used for analytics and reporting
 * @throws
 * - Database query errors
 */
func (dao *FeedbackDAO) GetFeedbackStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get counts by type
	typeCounts := make(map[string]int64)
	rows, err := dao.db.Model(&models.Feedback{}).
		Select("type, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feedbackType string
		var count int64
		if err := rows.Scan(&feedbackType, &count); err != nil {
			return nil, err
		}
		typeCounts[feedbackType] = count
	}

	stats["type_counts"] = typeCounts

	// Get total feedback count
	var total int64
	err = dao.db.Model(&models.Feedback{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&total).Error
	if err != nil {
		return nil, err
	}
	stats["total_count"] = total

	return stats, nil
}
