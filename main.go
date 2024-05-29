package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	_ "time/tzdata"
	"web-scrapper/controllers"
	"web-scrapper/database"
	"web-scrapper/middleware"
	"web-scrapper/scraping"

	_ "github.com/lib/pq"
)

var err error

func init() {
	// Load environment variables
	err = godotenv.Load("config/.env")
	if err != nil {
		fmt.Println("Failed to load environment file")
	} else {
		fmt.Println("Successfully read environment file")
	}

	// Initialize database connection
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGDATABASE"))
	err = database.InitDB(dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Initialize database tables
	if err := initDBTables(); err != nil {
		log.Fatalf("Error initializing database tables: %v", err)
	}

	// Ensure the /data directory exists
	ensureDataDirectory()

	// Start cron scheduler
	startCronScheduler()
}

func main() {
	// Set up Gin router
	router := gin.Default()
	router.Use(middleware.CORSMidleware()) // Apply CORS middleware
	router.POST("/api/v1/register", controllers.RegisterUser)
	router.POST("/api/v1/login", controllers.LoginUser)
	router.GET("/api/v1/reviews", controllers.GetAllReviews)
	protected := router.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		//these are for the website
		protected.GET("/secure_endpoint", secureEndpointHandler)
		protected.GET("/stasiun", controllers.GetAllStasiun)
		protected.GET("/schedules", controllers.GetAllSchedules)
		protected.GET("/schedules/:id", controllers.GetSchedulesByID)
		protected.GET("/schedules/:id/:arah", controllers.GetSchedulesByIDAndTrip)

		//these below are for sanber

		protected.GET("/v1/stasiun", controllers.GetAllStasiunV1)
		protected.GET("/v1/schedules", controllers.GetAllSchedulesV1)
		protected.GET("/v1/schedules/:id", controllers.GetSchedulesByStationIDV1)
		protected.GET("/v1/schedules/:id/:arah", controllers.GetSchedulesByIDAndTripV1)
		protected.POST("/v1/reviews", controllers.CreateReview)
	}

	// Serve HTTP requests with Gin router
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	err := router.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDBTables() error {
	// Read init.sql file
	initSQL, err := os.ReadFile("config/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read init.sql: %v", err)
	}

	// Execute init.sql statements
	_, err = database.GetDB().Exec(string(initSQL))
	if err != nil {
		return fmt.Errorf("error executing init.sql: %v", err)
	}

	log.Println("Database tables initialized successfully!")
	return nil
}

func startCronScheduler() {
	// Load Asia/Jakarta timezone
	loc, timeErr := time.LoadLocation("Asia/Jakarta")
	if timeErr != nil {
		log.Fatalf("Error loading timezone: %v", timeErr)
	}

	// Set up cron scheduler logic with timezone
	c := cron.New(cron.WithLocation(loc))
	// Schedule the cron job
	_, cronErr := c.AddFunc("0 0 * * *", func() {
		// Cron job to run daily at midnight
		if err := runScrapingTask(); err != nil {
			log.Printf("Error running scraping task: %v", err)
			return
		}
	})

	if cronErr != nil {
		log.Fatalf("Error adding cron job: %v", cronErr)
	}

	// Start the cron scheduler
	c.Start()
	log.Println("Cron job started. Waiting for signals...")
}

func runScrapingTask() error {
	log.Println("Running daily scraping task...")

	// Cleanup old csv files
	if err := scraping.CleanupOldCSVFiles("/data", 0); err != nil {
		return fmt.Errorf("error cleaning up CSV files: %v", err)
	}

	// Run the scraping logic
	if err := scraping.RunScraping(); err != nil {
		return fmt.Errorf("error running scraping: %v", err)
	}

	// Ensure the database connection is alive
	if err := database.GetDB().Ping(); err != nil {
		return fmt.Errorf("database connection lost: %v", err)
	}

	// Perform database operations
	if err := scraping.CreateTable(database.GetDB()); err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	if err := scraping.RemoveSchedules(database.GetDB()); err != nil {
		return fmt.Errorf("error removing schedules: %v", err)
	}
	if err := scraping.RemoveStations(database.GetDB()); err != nil {
		return fmt.Errorf("error removing stations: %v", err)
	}

	// Insert scraped data into the database
	if err := scraping.InsertData(database.GetDB()); err != nil {
		return fmt.Errorf("error inserting data: %v", err)
	}
	log.Println("Data inserted successfully!")
	return nil
}

func ensureDataDirectory() {
	dataDir := "/data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", dataDir, err)
		}
		log.Printf("Directory %s created successfully.", dataDir)
	} else if err != nil {
		log.Fatalf("Error checking directory %s: %v", dataDir, err)
	} else {
		log.Printf("Directory %s already exists.", dataDir)
	}
}

func secureEndpointHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "You have access to this endpoint"})
}
