# Service Template

Service Template is a model to take as a reference when building services in Go.  

## Description

The folder structure is based on the project layout suggested in [this repo](https://github.com/golang-standards/project-layout).

It contains a set of web utilities in the `web` package, under the `/pkg/web` folder.

Postgres is the main data storage for the service, but it can be easily replaced by any other database.

## Features

- ✔  Graceful Shutdown
- ✔  Debug/Metrics endpoint
- ✔  Health and liveness checks
- ✔  Authentication
- ✔  Authorization
- ✔  Data persistence using Postgres
- ✔  JWT generation
- ✔  Middleware (authorization, logging, metrics, panic and error handling)
- ✔  Docker support
- ✔  Kubernetes support
- ✔  Metrics (Prometheus)
- ✔  Tracing (Open Telemetry)

## License
[MIT](https://choosealicense.com/licenses/mit/)
