# poolside 

Configurable auto-reload proxy to put services behind domain names for better cookie management.

#### local development macOS

* Running docker on macOS `docker-compose -f docker-compose.mac.yml up`
* Running test server `make test-server`
* Running margarita `go run main.go`
* Sample config file `[{"domain":"test.local.bimmer-tech.com","proxy":"http://host.docker.internal:8000"}]`
* Default ui endpoint http://localhost:3000/
* Envoy admin endpoint http://localhost:9901/