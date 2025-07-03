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

See .env.example.txt for all available configuration options. Key variables include:

- `DB_DRIVER`, `DB_DATASOURCE` — SQL Database configuration.
- `OPENAI_API_KEY` — Your OpenAI API key.
- `LLM_VENDOR`, `LLM_MODEL` — LLM provider and model.
- `REDIS_ADDR`, `REDIS_PASSWORD` — Redis connection for job queue.
- `JWT_SECRET_KEY` — Secret for JWT signing.
- `JOBS_RUNNING_PER_USER_LIMIT` — Limit concurrent jobs per user.

---

## Development Notes

- All endpoints (except `/signup` and `/login`) require the `Authorization` header with a valid JWT.
- The API uses Gin for HTTP routing and Logrus for logging.
- Background jobs are managed with [Asynq](https://github.com/hibiken/asynq) and require a running Redis instance.
- Periodic cleanup of deleted jobs is handled automatically.
- Swagger documentation is available at `/swagger/index.html` after running the server.

---

## License

MIT License. See LICENSE for details.

---

## Contributing

Pull requests and issues are welcome! Please open an issue to discuss your ideas or report bugs.

---

## Contact

For questions or support, please contact [your-email@example.com].