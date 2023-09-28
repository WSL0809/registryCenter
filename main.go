package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Service struct {
	ID            uint      `gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"unique"`
	Host          string    `json:"host"`
	Port          int       `json:"port"`
	LastHeartbeat time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

var db *gorm.DB

func main() {
	go checkHeartbeats()

	var err error
	db, err = gorm.Open(sqlite.Open("./services.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Service{})

	router := gin.Default()
	router.POST("/register", registerService)
	router.GET("/services", getRegisteredServices)
	router.POST("/heartbeat", heartbeatService)
	err = router.Run(":8080")

	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

func registerService(c *gin.Context) {
	var service Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert service"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Service registered successfully"})
}

func getRegisteredServices(c *gin.Context) {
	var services []Service
	if err := db.Select("name, host, port").Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve services"})
		return
	}

	c.JSON(http.StatusOK, services)
}

func heartbeatService(c *gin.Context) {
	var service Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Service{}).Where("name = ?", service.Name).Update("last_heartbeat", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service heartbeat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat received"})
}

func checkHeartbeats() {
	fmt.Println("checkHeartbeats is started")
	ticker := time.NewTicker(40 * time.Second)
	for {
		select {
		case <-ticker.C:
			var servicesToCheck []Service
			if err := db.Where("strftime('%s', 'now') - strftime('%s', last_heartbeat) > 300").Find(&servicesToCheck).Error; err != nil {
				log.Println("Failed to retrieve services for heartbeat check:", err)
				continue
			}

			for _, s := range servicesToCheck {
				if err := db.Delete(&Service{}, s.ID).Error; err != nil {
					log.Println("Failed to remove stale service:", s.Name, err)
				}
				fmt.Println("keep checking ...")
			}
		}
	}
}
