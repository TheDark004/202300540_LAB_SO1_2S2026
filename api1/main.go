package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const CARNET = "202300540" 
const VM = "VM1"
const API2_URL = "http://127.0.0.1:8082" // API2 corre en la misma VM1
const API3_URL = "http://192.168.122.245:8083" // API3 corre en VM2


type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	VM        string `json:"VM"`
	Carnet    string `json:"carnet"`
}

type CallResponse struct {
	APIName    string `json:"apiname"`
	Message    string `json:"message"`
	Connection bool   `json:"connection"`
	Carnet     string `json:"carnet"`
}

// callHealth intenta contactar el /health de otra API y arma la respuesta en base a si tuvo exito o no
func callHealth(targetURL, targetAPIName, targetVM string) CallResponse {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL + "/health")
	if err != nil {
		return CallResponse{
			APIName:    targetAPIName,
			Message:    fmt.Sprintf("ERROR: The %s located on the %s is not working", targetAPIName, targetVM),
			Connection: false,
			Carnet:     CARNET,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CallResponse{
			APIName:    targetAPIName,
			Message:    fmt.Sprintf("ERROR: The %s located on the %s is not working", targetAPIName, targetVM),
			Connection: false,
			Carnet:     CARNET,
		}
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil || health.Status != "UP" {
		return CallResponse{
			APIName:    targetAPIName,
			Message:    fmt.Sprintf("ERROR: The %s located on the %s is not working", targetAPIName, targetVM),
			Connection: false,
			Carnet:     CARNET,
		}
	}

	return CallResponse{
		APIName:    targetAPIName,
		Message:    fmt.Sprintf("The %s located on the %s is working", targetAPIName, targetVM),
		Connection: true,
		Carnet:     CARNET,
	}
}

func main() {
	app := fiber.New()
	app.Use(logger.New())

	// GET /health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(HealthResponse{
			Status:    "UP",
			Message:   "API1 is Ready",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			VM:        VM,
			Carnet:    CARNET,
		})
	})

	// GET /api1/202300540/call-api2
	app.Get("/api1/:carnet/call-api2", func(c *fiber.Ctx) error {
		result := callHealth(API2_URL, "API2", "VM1")
		return c.JSON(result)
	})

	// GET /api1/202300540/call-api3
	app.Get("/api1/:carnet/call-api3", func(c *fiber.Ctx) error {
		result := callHealth(API3_URL, "API3", "VM2")
		return c.JSON(result)
	})

	fmt.Println("API1 corriendo en puerto 8081")
	log.Fatal(app.Listen(":8081"))
}