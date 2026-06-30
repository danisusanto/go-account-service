package main

import (
	"log"

	"github.com/danisusanto/go-account-service/internal/bootstrap"
)

// @title			Account Service API
// @version		1.0.0
// @description	Account Service API provides authentication, authorization, account management, and identity-related operations for applications and external integrations.
// @description
// @description					Features include:
// @description					- User authentication and session management
// @description					- Access token and refresh token lifecycle
// @description					- User registration and account provisioning
// @description					- Password management and credential verification
// @description					- Account profile management
// @description					- Role and permission assignment
// @description					- Secure API access using Bearer Authentication
//
// @termsOfService					https://your-domain.com/terms
//
// @contact.name					API Support
// @contact.url					https://your-domain.com/support
// @contact.email					support@your-domain.com
//
// @license.name					Apache 2.0
// @license.url					https://www.apache.org/licenses/LICENSE-2.0.html
//
// @BasePath						/api
//
// @securityDefinitions.apikey		ApiKeyAuth
// @in								header
// @name							Authorization
// @description					Bearer access token. Format: "Bearer <access_token>"
//
// @externalDocs.descriptionAPI	Documentation & Integration Guide
// @externalDocs.url				https://docs.your-domain.com
func main() {
	server, err := bootstrap.NewRestServer()
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
