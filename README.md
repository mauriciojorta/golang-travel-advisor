# Travel Advisor API

**Travel Advisor API** is a RESTful backend service built with Golang, designed to help users create, manage, and generate AI-powered travel itineraries. It leverages LLMs for itinerary generation through langchain and supports asynchronous job processing for file exports. The API uses JWT authentication and provides endpoints for user management, itinerary CRUD operations on a SQL database, and job management.

---

## Features

- **User Authentication:** Sign up and login with JWT-based authentication.
- **Itinerary Management:** Create, update, retrieve, and delete travel itineraries with multiple destinations.
- **AI-Powered Itinerary Generation:** Integrates with LLM APIs through langchain to generate detailed travel plans. The current version only supports OpenAI API so far, but it could be extended to support other LLM providers/vendors in the future. 
- **Asynchronous Job Processing:** Export itineraries as files using background jobs (with Redis and Asynq). The current version supports only local storage of job files, but it could be extended to support cloud storage providers like AWS S3 or Google Cloud Storage in the future.
- **Job Management:** Start, stop, download, and delete itinerary file generation jobs.
- **Role-based Access:** All sensitive endpoints are protected and require authentication.
- **Configurable via Environment Variables:** Easily adapt to different environments and requirements.

---

## Getting Started

### Prerequisites

- Go 1.20+
- Redis server (for background job processing). See [Redis docker image](https://hub.docker.com/_/redis) and [Redis docker container with password](https://github.com/redis/docker-library-redis/issues/176#issuecomment-723535421) for setup instructions with docker (preferred option).
- SQLite DB (default, can be changed)
- OpenAI API Key

### Installation

1. **Clone the repository:**
   ```
   git clone https://github.com/mauriciojorta/travel-advisor.git
   cd travel-advisor
   ```

2. **Copy and configure environment variables:**
   ```
   cp .env.example.txt .env
   ```
   Edit `.env` to set your OpenAI API key, database, and Redis configuration.

3. **Install dependencies:**
   ```
   go mod tidy
   ```

4. **Run the API server:**
   ```
   go run main.go
   ```
   The server will start on `localhost:8080`.

---

## API Endpoints

### Authentication

- `POST /api/v1/signup` — Register a new user.
- `POST /api/v1/login` — Login and receive a JWT token.

### Itineraries (Authenticated)

- `POST /api/v1/itineraries` — Create a new itinerary.
- `PUT /api/v1/itineraries` — Update an existing itinerary.
- `GET /api/v1/itineraries` — List all itineraries for the authenticated user.
- `GET /api/v1/itineraries/:itineraryId` — Get details of a specific itinerary.
- `DELETE /api/v1/itineraries/:itineraryId` — Delete an itinerary.

### Itinerary File Jobs (Authenticated)

- `POST /api/v1/itineraries/:itineraryId/jobs` — Start a file generation job for an itinerary.
- `GET /api/v1/itineraries/:itineraryId/jobs` — List all jobs for an itinerary.
- `GET /api/v1/itineraries/:itineraryId/jobs/:itineraryJobId` — Get job status/details.
- `GET /api/v1/itineraries/:itineraryId/jobs/:itineraryJobId/file` — Download the generated file.
- `PUT /api/v1/itineraries/:itineraryId/jobs/:itineraryJobId/stop` — Stop a running job.
- `DELETE /api/v1/itineraries/:itineraryId/jobs/:itineraryJobId` — Delete a job.

---

## Environment Variables

See `.env.example.txt` for all available configuration options. Below is a list of all supported environment variables and their purposes:

### Database Configuration

- `DB_DRIVER` — SQL driver to use (e.g., `sqlite`).
- `DB_DATASOURCE` — Database connection string (e.g., `file:api.sql?cache=shared&_busy_timeout=5000`).
- `DB_MAX_OPEN_CONNECTIONS` — Maximum number of open DB connections.
- `DB_MAX_IDLE_CONNECTIONS` — Maximum number of idle DB connections.

### OpenAI API Configuration

- `OPENAI_API_KEY` — Your OpenAI API key for LLM-powered itinerary generation.

### LLM (Large Language Model) Configuration

- `LLM_VENDOR` — LLM provider/vendor (e.g., `openai`).
- `LLM_MODEL` — LLM model name (e.g., `gpt-4o-2024-08-06`).
- `LLM_TEMPERATURE` — Sampling temperature for LLM responses (higher values = more creative).
- `LLM_MIN_RESPONSE_LENGTH` — Minimum length of LLM-generated responses.
- `LLM_MAX_RESPONSE_LENGTH` — Maximum length of LLM-generated responses.

### Logging

- `LOGGER_LEVEL` — Logging level (e.g., `debug`, `info`, `warn`, `error`).

### JWT Configuration

- `JWT_SECRET_KEY` — Secret key for signing JWT tokens.

### Asynchronous Job Processing

- `JOBS_RUNNING_PER_USER_LIMIT` — Maximum number of concurrent jobs per user.
- `ASYNC_TASK_TIMEOUT_MINUTES` — Timeout (in minutes) for async tasks.

### File Manager

- `FILE_MANAGER` — File storage backend. Only `local` is supported in current version.

### Redis Configuration

- `REDIS_ADDR` — Redis server address (e.g., `127.0.0.1:6379`).
- `REDIS_PASSWORD` — Redis password.

### Deleted Itinerary File Jobs Garbage Collection

- `DEAD_ITINERARY_FILE_JOBS_TIMER_MINUTES_INTERVAL` — Interval (in minutes) for running garbage collection of deleted jobs.
- `DEAD_ITINERARY_FILE_JOBS_FETCH_LIMIT` — Maximum number of deleted jobs to fetch and clean up per interval.

### Continuous Integration Environment

- `CI` — Set to `"local"` for local development or `"ci"` for CI environments.

---

**Note:**  
Copy `.env.example.txt` to `.env` and adjust these variables as needed

---

## Development Notes

- All endpoints (except `/signup` and `/login`) require the `Authorization` header with a valid JWT.
- The API uses Gin for HTTP routing and Logrus for logging.
- Background jobs are managed with [Asynq](https://github.com/hibiken/asynq) and require a running Redis instance.
- Periodic cleanup of deleted jobs is handled automatically.
- Swagger documentation is available at `/swagger/index.html` after running the server. To use the itinerary endpoints, you need to authenticate first and obtain a JWT token by logging in with a user account. The JWT token should be included in the `Authorization` header of your requests to the itinerary endpoints.
- Static Redoc HTML documentation for the API is available at /docs/redoc-static.html.
- Run tests with `go test ./...` to ensure everything is working correctly.

---

## License

MIT License. See LICENSE for details.

---

## Contributing

Pull requests and issues are welcome! Please open an issue to discuss your ideas or report bugs.