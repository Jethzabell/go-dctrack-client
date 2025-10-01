package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Jethzabell/go-dctrack-client/dctrack"
)

// Demo showing DCTrack authentication flow working
func main() {
	fmt.Println("DCTrack Authentication Flow Demo")
	fmt.Println(strings.Repeat("=", 50))

	// Create mock DCTrack server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Server: %s %s\n", r.Method, r.URL.Path)

		switch r.URL.Path {
		case "/api/v2/authentication/login":
			username, password, ok := r.BasicAuth()
			if !ok || username != "demo-user" || password != "demo-pass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return JWT token in Authorization header
			w.Header().Set("Authorization", "Bearer demo-jwt-token-12345")
			w.WriteHeader(http.StatusOK)
			fmt.Println("Login successful, returned Bearer token")

		case "/api/v2/quicksearch/items":
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer demo-jwt-token-12345" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return mock DCTrack response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			mockResponse := `{
				"totalRows": 2,
				"pageNumber": 1,
				"pageSize": 10,
				"searchResults": {
					"items": [
						{
							"id": "demo-001",
							"tiName": "Demo Server 1",
							"cmbLocation": "Demo-DC",
							"cmbStatus": "Installed",
							"tiClass": "Device",
							"cmbMake": "Dell",
							"cmbModel": "PowerEdge R730",
							"tiItemOriginalPower": 500
						},
						{
							"id": "demo-002", 
							"tiName": "Demo Server 2",
							"cmbLocation": "Demo-DC",
							"cmbStatus": "Installed",
							"tiClass": "Device",
							"cmbMake": "HPE",
							"cmbModel": "ProLiant DL380",
							"tiItemOriginalPower": 600
						}
					]
				}
			}`
			w.Write([]byte(mockResponse))
			fmt.Println("API request successful, returned mock data")

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create DCTrack client
	client, err := dctrack.New(
		server.URL+"/api/v2",
		"demo-user",
		"demo-pass",
		dctrack.WithPageSize(10),
		dctrack.WithInsecureSSL(),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("\n Making API call (triggers automatic authentication)...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	items, err := client.GetItems(ctx)
	if err != nil {
		log.Fatalf("API call failed: %v", err)
	}

	fmt.Printf("\n🎉 Success! Retrieved %d items:\n", len(items))
	for i, item := range items {
		fmt.Printf("  %d. %s (ID: %s) - %s %s in %s\n",
			i+1, item.Name, item.ID, item.Make, item.Model, item.Location)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Authentication Flow Verified:")
	fmt.Println("  1. Basic Auth (username/password) → Login")
	fmt.Println("  2. JWT Bearer token ← Response")
	fmt.Println("  3. Bearer token → API requests")
	fmt.Println("  4. Automatic & transparent!")
}
