package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Executes a command and returns its stdout lines as a string slice
func runCommand(name string, args ...string) ([]string, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %v", err)
	}

	var results []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "[+]") {
			results = append(results, line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command failed: %v", err)
	}

	return results, nil
}

// Deduplicates and sorts a string slice
func deduplicate(subs []string) []string {
	subMap := make(map[string]bool)
	for _, sub := range subs {
		subMap[sub] = true
	}

	var unique []string
	for sub := range subMap {
		unique = append(unique, sub)
	}

	sort.Strings(unique)
	return unique
}

// Function to run subdomain enumeration and return combined_subs
func runSubdomainEnumeration(domain string) ([]string, error) {
	// Run Subfinder
	fmt.Println("\n🔍 Running Subfinder...")
	subfinderResults, err := runCommand("subfinder", "-d", domain, "-silent")
	if err != nil {
		return nil, fmt.Errorf("error running Subfinder: %v", err)
	}

	// Run Sublist3r
	fmt.Println(" Running Sublist3r...")

	pythonPath := "/Users/mihad/Desktop/SecIq/Sublist3r/venv/bin/python" // path to venv Python
	sublist3rPath := "/Users/mihad/Desktop/SecIq/Sublist3r"

	sublist3rCmd := exec.Command(pythonPath, "sublist3r.py", "-d", domain)
	sublist3rCmd.Dir = sublist3rPath

	stdout, err := sublist3rCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get Sublist3r stdout: %v", err)
	}

	if err := sublist3rCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Sublist3r: %v", err)
	}

	var sublist3rRaw []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		sublist3rRaw = append(sublist3rRaw, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Sublist3r output: %v", err)
	}

	if err := sublist3rCmd.Wait(); err != nil {
		return nil, fmt.Errorf("Sublist3r command failed: %v", err)
	}

	// Apply regex to Sublist3r output only
	subdomainRegex := regexp.MustCompile(`(?i)[a-z0-9][a-z0-9_.-]*\.[a-z]{2,}`)
	var sublist3rFiltered []string
	for _, line := range sublist3rRaw {
		line = strings.TrimSpace(line)

		// Remove ANSI color codes (like \033[92m, \033[0m, etc.)
		ansiRegex := regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`)
		line = ansiRegex.ReplaceAllString(line, "")

		// Find all subdomains in the line, not just check if line matches
		matches := subdomainRegex.FindAllString(line, -1)
		sublist3rFiltered = append(sublist3rFiltered, matches...)
	}

	// Combine and deduplicate
	combined := append(subfinderResults, sublist3rFiltered...)
	combined_subs := deduplicate(combined)

	return combined_subs, nil
}

// DynamoDB client initialization
func getDynamoDBClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // or your AWS region
	)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// Store subdomains result in DynamoDB
func storeSubdomainsToDynamoDB(domain string, subdomains []string) error {
	client, err := getDynamoDBClient()
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String("subdomains"),
		Item: map[string]types.AttributeValue{
			"domain":           &types.AttributeValueMemberS{Value: domain},
			"subdomain_length": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", len(subdomains))},
			"subdomain_list": &types.AttributeValueMemberL{Value: func() []types.AttributeValue {
				list := make([]types.AttributeValue, len(subdomains))
				for i, sub := range subdomains {
					list[i] = &types.AttributeValueMemberS{Value: sub}
				}
				return list
			}()},
		},
	}
	_, err = client.PutItem(context.TODO(), input)
	return err
}

// Retrieve subdomains from DynamoDB
func getSubdomainsFromDynamoDB(domain string) ([]string, bool, error) {
	client, err := getDynamoDBClient()
	if err != nil {
		return nil, false, err
	}
	input := &dynamodb.GetItemInput{
		TableName: aws.String("subdomains"),
		Key: map[string]types.AttributeValue{
			"domain": &types.AttributeValueMemberS{Value: domain},
		},
	}
	result, err := client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, false, err
	}
	if len(result.Item) == 0 {
		return nil, false, nil
	}
	attr, ok := result.Item["subdomain_list"]
	if !ok {
		return nil, false, nil
	}
	listAttr, ok := attr.(*types.AttributeValueMemberL)
	if !ok {
		return nil, false, nil
	}
	subdomains := make([]string, 0, len(listAttr.Value))
	for _, v := range listAttr.Value {
		if s, ok := v.(*types.AttributeValueMemberS); ok {
			subdomains = append(subdomains, s.Value)
		}
	}
	fmt.Println("got subdomains from dynamo db")
	return subdomains, true, nil
}

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
	})

	// Enable CORS for frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// API endpoint to find subdomains (POST method for compatibility with existing frontend)
	app.Post("/find", func(c *fiber.Ctx) error {
		var request struct {
			Domain string `json:"domain"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if request.Domain == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Domain is required",
			})
		}

		// 1. Check DynamoDB first
		subdomains, found, err := getSubdomainsFromDynamoDB(request.Domain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "DB error: " + err.Error(),
			})
		}
		if found {
			return c.JSON(subdomains)
		}

		// 2. Run subdomain enumeration
		combined_subs, err := runSubdomainEnumeration(request.Domain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if len(combined_subs) > 0 {
			if err := storeSubdomainsToDynamoDB(request.Domain, combined_subs); err != nil {
				fmt.Println("Failed to store in DynamoDB:", err)
			}
			return c.JSON(combined_subs)
		} else {
			fmt.Println("No subdomains found, not storing to DynamoDB.")
			return c.Status(404).JSON(fiber.Map{
				"error": "No subdomains found for this domain",
			})
		}
	})

	app.Listen(":3000")
}
