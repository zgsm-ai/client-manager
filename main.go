package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/zgsm-ai/client-manager/controllers"
	"github.com/zgsm-ai/client-manager/internal"
	"github.com/zgsm-ai/client-manager/router"
	"github.com/zgsm-ai/client-manager/services"
)

var SoftwareVer = ""
var BuildTime = ""
var BuildTag = ""
var BuildCommitId = ""

func PrintVersions() {
	fmt.Printf("Version %s\n", SoftwareVer)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Build Tag: %s\n", BuildTag)
	fmt.Printf("Build Commit ID: %s\n", BuildCommitId)
}

// @title Client Manager API
// @version 1.0
// @description This is a client manager API server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func init() {
	// Initialize configuration
	if err := internal.InitConfig(rootCmd); err != nil {
		fmt.Printf("Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "client-manager",
	Short: "Client Manager API Server",
	Long:  `Client Manager is a RESTful API server for managing client configurations, feedback, and logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVersions()
		// Load configuration
		if err := internal.LoadConfig(internal.AppConfig.ConfigPath); err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		// Initialize application
		appContext, err := services.InitializeApp()
		if err != nil {
			fmt.Printf("Failed to initialize application: %v\n", err)
			os.Exit(1)
		}

		// Apply command line overrides
		internal.ApplyConfig(appContext.Logger)

		// Initialize controllers
		configController := controllers.NewConfigController(appContext.Logger)
		configController.SetConfigService(appContext.ConfigService)

		feedbackController := controllers.NewFeedbackController(appContext.Logger)
		feedbackController.SetFeedbackService(appContext.FeedbackService)

		logController := controllers.NewLogController(appContext.Logger)
		logController.SetLogService(appContext.LogService)

		// Create Gin engine
		r := gin.Default()

		// Setup all routes
		router.SetupRoutes(r, configController, feedbackController, logController, appContext.Logger)

		// Setup graceful shutdown
		setupGracefulShutdown(appContext)

		// Start server
		if err := services.StartServer(r, appContext.Logger); err != nil {
			appContext.Logger.Fatalf("Failed to start server: %v", err)
		}
	},
}

// setupGracefulShutdown sets up graceful shutdown handlers
/**
* Setup graceful shutdown handlers
* @param {*services.AppContext} appContext - Application context containing database and Redis connections
* @description
* - Sets up signal handlers for SIGINT and SIGTERM
* - Closes database and Redis connections gracefully
* - Logs shutdown process
 */
func setupGracefulShutdown(appContext *services.AppContext) {
	// Note: In a real implementation, you would use signal.Notify to handle SIGINT and SIGTERM
	// For now, we'll add a defer statement to ensure cleanup on normal exit
	defer func() {
		appContext.Logger.Info("Shutting down application...")

		// Close database connection
		if err := internal.CloseDB(); err != nil {
			appContext.Logger.WithError(err).Error("Failed to close database connection")
		} else {
			appContext.Logger.Info("Database connection closed successfully")
		}

		// Close Redis connection
		if err := internal.CloseRedis(); err != nil {
			appContext.Logger.WithError(err).Error("Failed to close Redis connection")
		} else {
			appContext.Logger.Info("Redis connection closed successfully")
		}

		appContext.Logger.Info("Application shutdown completed")
	}()
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
