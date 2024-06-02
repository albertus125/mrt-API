package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"web-scrapper/database"
	"web-scrapper/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your_secret_key")

var cacheInstance *cache.Cache

func init() {
	cacheInstance = cache.New(6*time.Hour, 10*time.Minute)
}
func RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "Data user dengan username " + user.Username + " telah disimpan",
			"success": false,
		})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Password = string(hashedPassword)
	query := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	err = database.DB.QueryRow(query, user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User registration failed"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data":    nil,
				"message": "Data employee dengan username " + user.Username + " telah disimpan",
				"success": false,
			})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "User berhasil registrasi",
		"success": true,
	})
}

func LoginUser(c *gin.Context) {
	var loginDetails models.User
	if err := c.ShouldBindJSON(&loginDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	query := `SELECT id, username, password, role FROM users WHERE username = $1`
	err := database.DB.QueryRow(query, loginDetails.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Username atau Password salah"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginDetails.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Username atau Password salah"})
		return
	}
	var expirationTime time.Time

	if user.Role == "admin" {
		expirationTime = time.Now().Add(365 * 24 * time.Hour)
	} else {
		expirationTime = time.Now().Add(24 * time.Hour)
	}
	claims := &models.Claims{
		Username: user.Username,
		UserID:   user.ID,
		Role:     user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    tokenString,
		"message": "login success",
		"success": true,
	})
}

func CreateReview(c *gin.Context) {
	db := database.GetDB()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var review models.Review
	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review.UserID = userID.(int)

	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", review.UserID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking user existence"})
		return
	}
	if !userExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes":  false,
			"message": "user tidak ditemukan"})
		return
	}

	query := `INSERT into reviews(user_id, rating, comment) VALUES ($1,$2,$3) RETURNING user_id, created_at`
	err = db.QueryRow(query, review.UserID, review.Rating, review.Comment).Scan(&review.ID, &review.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"succes":  true,
		"message": "review berhasil ditambahkan",
		"data":    review})
}

func GetAllReviews(c *gin.Context) {
	db := database.GetDB()

	var reviews []models.Review

	query := `SELECT id, rating, comment, created_at FROM reviews  ORDER BY created_at DESC `
	rows, err := db.Query(query)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var review models.Review
		if err := rows.Scan(&review.ID, &review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reviews = append(reviews, review)
	}
	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "berhasil mengambil seluruh data review",
		"data":    reviews})
}

func GetAllSchedules(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	cacheKey := "all_schedules"
	cachedData, found := c.Get(cacheKey)
	if found {
		log.Println("fetching cached data")
		c.Header("X-Data-Source", "Cache")
		c.JSON(http.StatusOK, cachedData)
		return
	}
	rows, err := db.Query("SELECT id, station_id, stasiun_name, arah, to_char(jadwal, 'HH24:MI') as jadwal FROM schedules")
	if err != nil {
		log.Printf("Error fetching schedules: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data"})
		return
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		if err := rows.Scan(&schedule.ID, &schedule.StasiunID, &schedule.StasiunName, &schedule.Arah, &schedule.Jadwal); err != nil {
			log.Printf("Error scanning row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing data"})
			return
		}
		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating rows"})
		return
	}
	cacheInstance.Set(cacheKey, schedules, 6*time.Hour)

	c.Header("X-Data-Source", "API")
	c.JSON(http.StatusOK, schedules)
}

func GetAllStasiun(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	cacheKey := "all_stasiun"
	cachedData, found := c.Get(cacheKey)
	if found {
		log.Println("fetching cached data")
		c.Header("X-Data-Source", "Cache")
		c.JSON(http.StatusOK, cachedData)
		return
	}

	rows, err := db.Query("SELECT * FROM stations")
	if err != nil {
		log.Printf("Error fetching stations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data"})
		return
	}
	defer rows.Close()

	var stasiun []models.Stasiun
	for rows.Next() {
		var station models.Stasiun
		if err := rows.Scan(&station.StasiunID, &station.StasiunName); err != nil {
			log.Printf("Error scanning row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing data"})
			return
		}
		stasiun = append(stasiun, station)
	}
	cacheInstance.Set(cacheKey, stasiun, 6*time.Hour)
	c.Header("X-Data-Source", "API")
	c.JSON(http.StatusOK, stasiun)
}
func GetSchedulesByID(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
	}
	stationIDStr := c.Param("id")
	stationID, _ := strconv.Atoi(stationIDStr)

	cacheKey := fmt.Sprintf("schedules_stations_%d", stationID)

	if cachedData, found := cacheInstance.Get(cacheKey); found {
		c.Header("x-Data-Source", "cache")
		c.JSON(http.StatusOK, cachedData)
		return
	}

	query := `
	SELECT id, station_id, stasiun_name, arah, to_char(jadwal, 'HH24:MI') as jadwal 
		FROM schedules WHERE station_id = $1
		
	`
	rows, err := db.Query(query, stationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var schedules []models.Schedule

	for rows.Next() {
		var schedule models.Schedule
		if err := rows.Scan(&schedule.ID, &schedule.StasiunID, &schedule.StasiunName, &schedule.Arah, &schedule.Jadwal); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		schedules = append(schedules, schedule)
	}
	cacheInstance.Set(cacheKey, schedules, 6*time.Hour)

	c.Header("X-Data-Source", "API")
	c.JSON(http.StatusOK, schedules)
}

func GetSchedulesByIDAndTrip(c *gin.Context) {
	stationIDStr := c.Param("id")
	arah := c.Param("arah")
	cacheKey := fmt.Sprintf("%s_%s", stationIDStr, arah)

	//load asia/jakarta
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return
	}

	//get current time in the jakarta
	now := time.Now().In(loc)

	// Check if data is cached
	if cachedData, found := cacheInstance.Get(cacheKey); found {
		if schedules, ok := cachedData.([]models.Schedule); ok {
			c.Header("X-Data-Source", "Cache")
			c.JSON(http.StatusOK, schedules) // Return cached data
			return
		}
	}

	// Fetch schedules from database
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
		return
	}

	var rows *sql.Rows
	var queryErr error

	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		rows, queryErr = db.Query(`
		SELECT id, station_id, stasiun_name, arah, to_char(jadwal, 'HH24:MI') as jadwal 
		FROM schedules 
		WHERE  station_id = $1 AND arah = $2 AND stasiun_name = ''`, stationIDStr, arah)
	} else {
		rows, queryErr = db.Query(`
		SELECT id, station_id, stasiun_name, arah, to_char(jadwal, 'HH24:MI') as jadwal 
		FROM schedules 
		WHERE  station_id = $1 AND arah = $2 AND stasiun_name <> ''
	`, stationIDStr, arah)
	}

	if queryErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": queryErr.Error()})
		return
	}
	defer rows.Close()
	// Process rows and filter out duplicates
	seen := make(map[int]bool)
	var uniqueSchedules []models.Schedule

	for rows.Next() {
		var s models.Schedule
		var stasiunName sql.NullString
		err := rows.Scan(&s.ID, &s.StasiunID, &stasiunName, &s.Arah, &s.Jadwal)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if the schedule with the same ID has been seen before
		s.StasiunName = stasiunName.String
		if !seen[s.ID] {
			seen[s.ID] = true
			uniqueSchedules = append(uniqueSchedules, s)
		}
	}

	// Cache the fetched schedules
	cacheInstance.Set(cacheKey, uniqueSchedules, 6*time.Hour)

	// Set response header to indicate API source
	c.Header("X-Data-Source", "API")
	c.JSON(http.StatusOK, uniqueSchedules) // Return the fetched schedules
}
