package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/Shopify/sarama"
	"github.com/gocql/gocql"
)

type FileMetadata struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Status      string `json:"status"`
}

func main() {
	// Initialize Kafka producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{"kafka:9092"}, config)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Initialize Cassandra session
	cluster := gocql.NewCluster("cassandra")
	cluster.Keyspace = "cad_automation"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to create Cassandra session: %v", err)
	}
	defer session.Close()

	// Initialize Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// File upload endpoint
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
			return
		}

		// Create metadata
		metadata := FileMetadata{
			ID:          uuid.New().String(),
			Filename:    file.Filename,
			ContentType: file.Header.Get("Content-Type"),
			Size:        file.Size,
			Status:      "pending",
		}

		// Save metadata to Cassandra
		if err := session.Query(`
			INSERT INTO files (id, filename, content_type, size, status)
			VALUES (?, ?, ?, ?, ?)`,
			metadata.ID, metadata.Filename, metadata.ContentType, metadata.Size, metadata.Status,
		).Exec(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
			return
		}

		// Send message to Kafka
		msg := &sarama.ProducerMessage{
			Topic: "cad-files",
			Value: sarama.StringEncoder(metadata.ID),
		}
		_, _, err = producer.SendMessage(msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue file for processing"})
			return
		}

		c.JSON(http.StatusOK, metadata)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 