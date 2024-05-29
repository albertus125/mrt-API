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

	"github.com/gin-gonic/gin"
)

func GetAllSchedulesV1(c *gin.Context) {
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
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Sukses mengambil seluruh data schedules",
			"data":    cachedData,
		})
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
	c.JSON(http.StatusOK, gin.H{
		"data":    schedules,
		"message": "Sukses mengambil seluruh data schedules",
		"success": true,
	})
}

func GetSchedulesByStationIDV1(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
	}
	stationIDStr := c.Param("id")
	stationID, _ := strconv.Atoi(stationIDStr)

	cacheKey := fmt.Sprintf("schedules_stations_%d", stationID)

	if cachedData, found := cacheInstance.Get(cacheKey); found {
		c.Header("x-Data-Source", "cache")
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Data schedule dengan stasiun id" + stationIDStr + "berhasil diambil",
			"data":    cachedData,
		})
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
	if len(schedules) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "Data schedule dengan stasiun id " + stationIDStr + " tidak ditemukan",
			"success": false,
		})
		return
	}
	cacheInstance.Set(cacheKey, schedules, 6*time.Hour)

	c.Header("X-Data-Source", "API")
	c.JSON(http.StatusOK, gin.H{
		"data":    schedules,
		"message": "Data schedule dengan stasiun id " + stationIDStr + " berhasil diambil",
		"success": true,
	})
}

func GetSchedulesByIDAndTripV1(c *gin.Context) {
	stationIDStr := c.Param("id")
	arah := c.Param("arah")
	cacheKey := fmt.Sprintf("%s_%s", stationIDStr, arah)

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

	rows, err := db.Query(`
		SELECT id, station_id, stasiun_name, arah, to_char(jadwal, 'HH24:MI') as jadwal 
		FROM schedules 
		WHERE  station_id = $1 AND arah = $2 AND stasiun_name <> ''
	`, stationIDStr, arah)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if len(uniqueSchedules) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":    nil,
			"message": "Data schedule dengan stasiun ID: " + stationIDStr + " dan arah " + arah + " tidak ditemukan",
			"success": false,
		})
		return
	}
	// Cache the fetched schedules
	cacheInstance.Set(cacheKey, uniqueSchedules, 6*time.Hour)

	// Set response header to indicate API source
	c.Header("X-Data-Source", "API")
	c.JSON(http.StatusOK, gin.H{
		"data":    uniqueSchedules,
		"message": "Data schedule dengan stasiun ID: " + stationIDStr + " dan arah: " + arah + " berhasil diambil",
		"success": true,
	}) // Return the fetched schedules
}
