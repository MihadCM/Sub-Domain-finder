package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SubdomainRecord represents a subdomain lookup record
type SubdomainRecord struct {
	Domain     string    `json:"domain"`
	Subdomains []string  `json:"subdomains"`
	Timestamp  time.Time `json:"timestamp"`
}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	// Initialize DynamoDB session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Change to your AWS region
	})
	if err != nil {
		log.Fatal("Failed to create AWS session:", err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Store subdomains
	app.Post("/store", func(c *fiber.Ctx) error {
		var record SubdomainRecord
		if err := c.BodyParser(&record); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Set timestamp if not provided
		if record.Timestamp.IsZero() {
			record.Timestamp = time.Now()
		}

		// Marshal to DynamoDB format
		av, err := dynamodbattribute.MarshalMap(record)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to marshal data"})
		}

		// Create input for DynamoDB
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("subdomain-lookups"), // Your DynamoDB table name
		}

		// Put item in DynamoDB
		_, err = svc.PutItem(input)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to store data: " + err.Error()})
		}

		return c.JSON(fiber.Map{"message": "Subdomains stored successfully"})
	})

	// Get subdomains for a specific domain
	app.Get("/subdomains/:domain", func(c *fiber.Ctx) error {
		domain := c.Params("domain")

		// Query DynamoDB
		input := &dynamodb.GetItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"domain": {
					S: aws.String(domain),
				},
			},
			TableName: aws.String("subdomain-lookups"),
		}

		result, err := svc.GetItem(input)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to retrieve data: " + err.Error()})
		}

		if result.Item == nil {
			return c.Status(404).JSON(fiber.Map{"error": "No data found for domain: " + domain})
		}

		// Unmarshal result
		var record SubdomainRecord
		err = dynamodbattribute.UnmarshalMap(result.Item, &record)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to unmarshal data"})
		}

		return c.JSON(record)
	})

	// Get all stored lookups
	app.Get("/history", func(c *fiber.Ctx) error {
		// Scan DynamoDB table
		input := &dynamodb.ScanInput{
			TableName: aws.String("subdomain-lookups"),
		}

		result, err := svc.Scan(input)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to scan table: " + err.Error()})
		}

		var records []SubdomainRecord
		for _, item := range result.Items {
			var record SubdomainRecord
			err := dynamodbattribute.UnmarshalMap(item, &record)
			if err != nil {
				continue // Skip invalid records
			}
			records = append(records, record)
		}

		return c.JSON(records)
	})

	fmt.Println("Storage service starting on port 3001...")
	log.Fatal(app.Listen(":3001"))
}
