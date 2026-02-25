package main

import (
	"log"
	"os"

	"leeforge-example-service/bootstrap"
)

// @title           Leeforge Example API
// @version         1.0
// @description     Leeforge Headless CMS Backend API with modular architecture
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@leeforge.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey APIKey
// @in header
// @name X-API-Key
// @description API Key for programmatic access.

// @securityDefinitions.apikey DomainType
// @in header
// @name X-Domain-Type
// @description Domain type for the request (e.g. "tenant", "org"). Used together with X-Domain-Key.

// @securityDefinitions.apikey DomainKey
// @in header
// @name X-Domain-Key
// @description Domain key for the request (e.g. tenant UUID). Used together with X-Domain-Type.

// @Security BearerAuth
// @Security APIKey
// @Security DomainType
// @Security DomainKey
func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := app.Engine().Run(":" + port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
