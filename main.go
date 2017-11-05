package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nada-hammad/readily/chatbot"
)

// Autoload environment variables in .env
import _ "github.com/joho/godotenv/autoload"

func main() {
	// Use the PORT environment variable
	port := os.Getenv("PORT")

	// Default to 3000 if no PORT environment variable was defined
	if port == "" {
		port = "3000"
	}

	// Start the server
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatalln(chatbot.Engage(":" + port))
}
