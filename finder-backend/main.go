package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

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
		if subdomainRegex.MatchString(line) {
			sub := subdomainRegex.FindString(line)
			sublist3rFiltered = append(sublist3rFiltered, sub)
		}
	}

	// Combine and deduplicate
	combined := append(subfinderResults, sublist3rFiltered...)
	combined_subs := deduplicate(combined)

	return combined_subs, nil
}

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Minute, // Allow long-running scans
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

		// Run subdomain enumeration
		combined_subs, err := runSubdomainEnumeration(request.Domain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return the combined_subs variable directly as an array
		// This matches what the frontend expects
		return c.JSON(combined_subs)
	})

	// Additional GET endpoint for convenience
	app.Get("/api/subdomains/:domain", func(c *fiber.Ctx) error {
		domain := c.Params("domain")

		// Run subdomain enumeration
		combined_subs, err := runSubdomainEnumeration(domain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"domain":     domain,
			"subdomains": combined_subs, // This is your combined_subs variable
			"count":      len(combined_subs),
		})
	})

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})
	// Start the server
	app.Listen(":3000")
}
