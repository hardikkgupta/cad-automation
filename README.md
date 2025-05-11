# CAD Automation Platform

A cloud-native platform for automating CAD file processing, conversion, and validation using microservices architecture.

## Architecture

The platform consists of the following components:

- **Ingestion Service** (Go): Handles CAD file uploads and initial validation
- **Conversion Service** (Go): Converts CAD files between different formats
- **Validation Service** (Node.js): Performs detailed CAD file validation
- **API Gateway** (Node.js): Provides a unified REST API interface
- **Event Bus** (Apache Kafka): Handles asynchronous communication between services
- **Metadata Store** (Cassandra): Stores file metadata and processing status

## Prerequisites

- Docker
- Kubernetes cluster
- Helm 3.x
- Go 1.21+
- Node.js 18+
- Apache Kafka
- Apache Cassandra

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/hardikkgupta/cad-automation.git
cd cad-automation
```

2. Install dependencies:
```bash
# Install Go dependencies
cd services/ingestion && go mod download
cd ../conversion && go mod download
cd ../validation && npm install
cd ../api-gateway && npm install
```

3. Deploy using Helm:
```bash
helm install cad-automation ./helm
```

## Development

### Local Development

1. Start required services:
```bash
docker-compose up -d
```

2. Run services locally:
```bash
# Run Go services
cd services/ingestion && go run main.go
cd ../conversion && go run main.go

# Run Node.js services
cd services/validation && npm run dev
cd ../api-gateway && npm run dev
```

### Testing

```bash
# Run Go tests
go test ./...

# Run Node.js tests
npm test
```

## API Documentation

API documentation is available at `/api-docs` when running the API Gateway service.

## License

MIT License - see LICENSE file for details