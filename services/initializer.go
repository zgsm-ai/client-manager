package services

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zgsm-ai/client-manager/dao"
	"github.com/zgsm-ai/client-manager/internal"
	"github.com/zgsm-ai/client-manager/utils"
)

// AppContext holds all the core application objects
type AppContext struct {
	DB              *gorm.DB
	Redis           *redis.Client
	Logger          *logrus.Logger
	ConfigDAO       *dao.ConfigDAO
	FeedbackDAO     *dao.FeedbackDAO
	LogDAO          *dao.LogDAO
	ConfigService   *ConfigService
	FeedbackService *FeedbackService
	LogService      *LogService
}

// InitializeApp initializes all core application objects and returns AppContext
/**
 * Initialize application core objects
 * @returns {*AppContext, error} Application context and error if initialization fails
 * @description
 * - Initializes database connection
 * - Initializes Prometheus metrics
 * - Creates all DAO objects
 * - Creates all service objects
 * - Creates all controller objects
 * @throws
 * - Database initialization error
 */
func InitializeApp() (*AppContext, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database
	db, err := internal.InitDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	// Initialize Redis
	redisClient, err := internal.InitRedis()
	if err != nil {
		logger.Warnf("Failed to initialize Redis: %v, continuing without Redis", err)
		redisClient = nil
	}

	// Initialize Prometheus metrics
	internal.InitMetrics()

	// Initialize DAOs
	configDAO := dao.NewConfigDAO(db, redisClient, logger)
	feedbackDAO := dao.NewFeedbackDAO(db, logger)
	logDAO := dao.NewLogDAO(db, logger)

	// Initialize services
	configService := NewConfigService(configDAO, logger)
	feedbackService := NewFeedbackService(feedbackDAO, logger)
	logService := NewLogService(logDAO, logger)

	// Create and return app context
	appContext := &AppContext{
		DB:              db,
		Redis:           redisClient,
		Logger:          logger,
		ConfigDAO:       configDAO,
		FeedbackDAO:     feedbackDAO,
		LogDAO:          logDAO,
		ConfigService:   configService,
		FeedbackService: feedbackService,
		LogService:      logService,
	}

	return appContext, nil
}

// StartServer starts the HTTP server
/**
 * Start HTTP server
 * @param {*gin.Engine} r - Gin engine
 * @param {*logrus.Logger} logger - Application logger
 * @description
 * - Gets server port from configuration
 * - Records startup time
 * - Starts the HTTP server
 * @throws
 * - Server start error
 */
func StartServer(r *gin.Engine, logger *logrus.Logger) error {
	// Get port from configuration
	port := internal.GetServerPort()

	// Start server
	serverAddr := fmt.Sprintf(":%s", port)
	logger.Infof("Starting server on %s", serverAddr)

	// Record startup time
	utils.SetStartupTime(time.Now())

	return r.Run(serverAddr)
}
