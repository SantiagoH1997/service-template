# Service Template

Service Template is a model to take as a reference when building services in Go, containing a set of web utilities in the `web` package, under the `/pkg/web` folder.

It uses a Postgres as the main data storage, running locally as a Docker container.

## Features

- [x]  Graceful Shutdown
- [x]  Liveness handler
- [x]  Authentication
- [x]  Authorization
- [x]  JWT generation
- [x]  Request parsing
- [x]  Response handling
- [x]  Logging middleware
- [x]  Panic handling middleware
- [x]  Error handling middleware
- [x]  Metrics
- [x]  Docker support
- [x]  Kubernetes support
- [x]  Postgres connection
- [x]  Migrations support
- [x]  Tracing (Open Telemetry)

## License
[MIT](https://choosealicense.com/licenses/mit/)