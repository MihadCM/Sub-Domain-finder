package main

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// runSubfinder executes the Subfinder tool and returns a slice of discovered subdomains.
func runSubfinder(domain string) ([]string, error) {
	cmd := exec.Command("subfinder", "-d", domain, "-silent")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	// fmt.Println(stdout) //pipe

	if err := cmd.Start(); err != nil {
		fmt.Println("err", err)
		fmt.Println("start", cmd.Start())
		return nil, fmt.Errorf("failed to start command: %v", err)
	}

	var subdomains []string
	// fmt.Println("subdomains", subdomains)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		subdomains = append(subdomains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading command output: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command execution failed: %v", err)
	}

	return subdomains, nil
}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	app.Post("/find", func(c *fiber.Ctx) error {
		type Request struct {
			Domain string `json:"domain"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		subdomains, err := runSubfinder(req.Domain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(subdomains)
	})

	app.Listen(":3000")

}
