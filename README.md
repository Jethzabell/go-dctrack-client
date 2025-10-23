# Go DCTrack Client

A comprehensive Go client library for [Sunbird DCTrack](https://www.sunbirddcim.com/) datacenter infrastructure management API.

[![Go Reference](https://pkg.go.dev/badge/github.com/Jethzabell/go-dctrack-client.svg)](https://pkg.go.dev/github.com/Jethzabell/go-dctrack-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/Jethzabell/go-dctrack-client)](https://goreportcard.com/report/github.com/Jethzabell/go-dctrack-client)
[![Test](https://github.com/Jethzabell/go-dctrack-client/actions/workflows/test.yml/badge.svg)](https://github.com/Jethzabell/go-dctrack-client/actions/workflows/test.yml)

## Features

* **Complete DCTrack API v2 Integration**: Full support for items, power, analytics
* **Automatic Authentication**: JWT token management with automatic renewal
* **Flexible Filtering**: Multiple ways to filter and search assets
* **Production Ready**: Comprehensive error handling, retries, and logging
* **Type Safety**: Strongly typed models for all DCTrack data
* **Performance Optimized**: Configurable field selection and pagination
* **Development Tools**: CLI tool for testing and exploration
* **Comprehensive Testing**: Unit tests, integration tests, and benchmarks

## Installation

```bash
go get github.com/Jethzabell/go-dctrack-client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    dctrack "github.com/Jethzabell/go-dctrack-client"
)

func main() {
    // Create client with simple configuration
    client, err := dctrack.New(
        "https://dctrack.company.com/api/v2",
        "username",
        "password",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Get all installed assets
    items, err := client.GetItems(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d assets\n", len(items))
}
```

## API Reference

### Client Creation

```go
// Simple client creation
client, err := dctrack.New("https://dctrack.company.com/api/v2", "user", "pass")

// Advanced client with custom configuration
config := dctrack.Config{
    URL:              "https://dctrack.company.com/api/v2",
    Username:         "api-user",
    Password:         "secure-password",
    PageSize:         1000,
    MaxRetries:       5,
    RetryDelay:       2 * time.Second,
    VerifySSL:        true,
    RequestAllFields: true,
}
client, err := dctrack.NewClient(config)
```

### Basic Operations

```go
// Get all items
items, err := client.GetItems(ctx)

// Get specific item by ID
item, err := client.GetItemByID(ctx, "12345")

// Search items by text
items, err := client.SearchItems(ctx, "PowerEdge R730")
```

### Advanced Filtering

```go
// Method 1: Parameter struct
items, err := client.GetItemsWithParams(ctx, dctrack.ItemsParams{
    Location:  "RDU2",
    Status:    "Installed",
    Make:      "Dell",
    PageSize:  100,
})

// Method 2: Filter builder (fluent interface)
filters := dctrack.NewFilterBuilder().
    Location("RDU2").
    Status("Installed").
    Make("Dell").
    WithPagination(1, 100).
    Build()
items, err := client.GetItemsWithParams(ctx, filters)

// Method 3: Preset filters
items, err := client.GetItemsWithParams(ctx, dctrack.ByLocation("RDU2"))
items, err := client.GetItemsWithParams(ctx, dctrack.ByVendor("HPE"))
items, err := client.GetItemsWithParams(ctx, dctrack.InstalledOnly())
```

## CLI Tool

The library includes a command-line tool for testing and exploration:

```bash
# Build the CLI
go build ./cmd/dctrackcheck

# Set environment variables
export DCTRACK_URL="https://dctrack.company.com/api/v2"
export DCTRACK_USERNAME="your-username"
export DCTRACK_PASSWORD="your-password"

# Search for assets
./dctrackcheck PowerEdge

# List assets by location
./dctrackcheck list RDU2

# Get specific item details
./dctrackcheck item 12345

# Power analysis for location
./dctrackcheck power RDU2
```

### CLI Environment Variables

* `DCTRACK_URL`: DCTrack API base URL (required)
* `DCTRACK_USERNAME`: DCTrack username (required)
* `DCTRACK_PASSWORD`: DCTrack password (required)
* `DCTRACK_PAGE_SIZE`: Page size for requests (default: 100)
* `DCTRACK_VERIFY_SSL`: Verify SSL certificates (default: true)
* `DCTRACK_ALL_FIELDS`: Request all available fields (default: false)

## Configuration

### YAML Configuration

Create a `config.yaml` file for multi-environment support:

```yaml
integration:
  url: "https://dctrack-dev.company.com/api/v2"
  username: "dev-api-user"
  password_file: "~/.secrets/on-premise-asset-hub-secrets/dctrack_password.txt"
  page_size: 100
  verify_ssl: false

production:
  url: "https://dctrack.company.com/api/v2"
  username: "prod-api-user"
  password_file: "~/.secrets/on-premise-asset-hub-secrets/dctrack_password.txt"
  page_size: 1000
  verify_ssl: true

local:
  url: "https://dctrack.company.com/api/v2"
  username: "prod-api-user"
  password_file: "~/.secrets/on-premise-asset-hub-secrets/dctrack_password.txt"
  page_size: 1000
  verify_ssl: false
```

### Environment Variables

All configuration options can be set via environment variables:

* `DCTRACK_URL`
* `DCTRACK_USERNAME`
* `DCTRACK_PASSWORD`
* `DCTRACK_PAGE_SIZE`
* `DCTRACK_MAX_RETRIES`
* `DCTRACK_VERIFY_SSL`

## Development

### Building and Testing

```bash
# Install dependencies
make install

# Run all tests
make test

# Build library and CLI
make build
make cli

# Run CLI tool
make run-cli DCTRACK_URL=https://dctrack.company.com/api/v2 DCTRACK_USER=username DCTRACK_PASS=password SEARCH_QUERY=PowerEdge

# Check release readiness
make release-check
```

### Available Make Commands

Run `make help` to see all available commands including:

* **Build automation** (`make build`, `make cli`)
* **Testing** (`make test`, `make benchmark`, `make coverage`)
* **Code quality** (`make fmt`, `make vet`, `make check`)
* **Development workflow** (`make dev`, `make quick`)
* **Examples** (`make example-basic`, `make example-advanced`)

## Error Handling

The library provides comprehensive error handling:

* **Connection failures**: Network or server connectivity issues
* **Authentication failures**: Invalid credentials or token expiration
* **API errors**: DCTrack server errors with detailed messages
* **Data validation**: Type conversion and field validation errors
* **Rate limiting**: Automatic retry logic with exponential backoff

## Security Considerations

* **Credentials**: Store passwords securely, never in code
* **TLS**: Use HTTPS and verify SSL certificates in production
* **Service Accounts**: Use dedicated service accounts with minimal permissions
* **Token Management**: Automatic token refresh handles expiration
* **Connection Management**: Close clients when done to free resources

## Performance

* **Pagination**: Configurable page sizes for optimal performance
* **Field Selection**: Choose between all fields or essential fields only
* **Connection Pooling**: Efficient HTTP client with connection reuse
* **Parallel Requests**: Context-aware operations support cancellation
* **Caching**: Client-side caching of authentication tokens

## Data Models

### DCTrackItem

Complete representation of DCTrack hardware items including:

* **Core Identity**: ID, Name, Location, Status, Type
* **Physical Attributes**: Cabinet, Rack, Position, Height, Weight
* **Technical Specifications**: Power, CPU, RAM, Network configuration
* **Administrative Data**: Contact information, admin teams, purchase details
* **DCTrack Custom Fields**: All ti_custom_field_* attributes
* **Lifecycle Information**: Install dates, contract end dates, warranty

See the [API documentation](https://pkg.go.dev/github.com/Jethzabell/go-dctrack-client) for complete field reference.

## Examples

* [Basic Usage](examples/basic_usage/) - Simple client setup and asset retrieval
* [Advanced Filtering](examples/advanced_filtering/) - Complex filtering and analytics

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Run the test suite (`make test`)
5. Run code quality checks (`make check`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0

* Initial release with complete DCTrack API v2 support
* Comprehensive test suite with 20+ tests
* Professional CLI tool (dctrackcheck)
* Multi-environment configuration support
* Production-ready error handling and logging
* Advanced filtering and search capabilities
* Performance optimizations and benchmarks

## Support

* **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/Jethzabell/go-dctrack-client)
* **Issues**: [GitHub Issues](https://github.com/Jethzabell/go-dctrack-client/issues)
* **Discussions**: [GitHub Discussions](https://github.com/Jethzabell/go-dctrack-client/discussions)

## Related Projects

* [go-ldap-redhat](https://github.com/openshift-eng/go-ldap-redhat) - Red Hat LDAP integration
* [DCTrack Documentation](https://www.sunbirddcim.com/dctrack) - Official Sunbird DCTrack docs

---

**Developed by Red Hat Engineering** for the datacenter management community.