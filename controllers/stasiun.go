package controllers

import (
	"log"
	"net/http"
	"time"
	"web-scrapper/database"
	"web-scrapper/models"

	"github.com/gin-gonic/gin"
)

func GetAllStasiunV1(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Berhasil mengambil semua data stasiun",
		"data":    stasiun,
	})

}
