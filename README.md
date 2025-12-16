# Order Packs Calculator

This is a small “production-grade” Go application that:

- Lets you **manage pack sizes** (CRUD + reset-to-default) stored in **SQLite**
- Lets you **calculate** a pack allocation for a requested quantity using the rules:
  - Use **whole packs only**
  - **Minimize total items shipped** (least overage)
  - If there’s a tie, **minimize the number of packs**
- Exposes both:
  - A JSON **HTTP API**
  - A simple **UI** at `GET /` to interact with the API

## Tech stack

- **Go**: HTTP server uses `net/http` + `chi`
- **SQLite**: `database/sql` + `modernc.org/sqlite`
- **Logging**: `log/slog` JSON to stdout
- **Config**: environment variables with optional `.env` (local)

## Run locally

1) Create a `.env` from the example:

```bash
cp env.example .env
```

2) Run the API:

```bash
make run
```

3) Open the UI:

- `GET http://localhost:<HTTP_PORT>/`

## Run tests

```bash
make test
```

## Configuration

Configuration is read from real environment variables. If a `.env` file exists in the project root, it is loaded for local development.

`env.example` contains the supported keys:

- **APP_ENV**: `dev|prod` (used for logging/behavior where applicable)
- **LOG_LEVEL**: `DEBUG|INFO|WARN|ERROR`
- **HTTP_PORT**: port number (e.g. `8080`)
- **DB_PATH**: SQLite file path (e.g. `./data/app.db`)

## API conventions

All responses are JSON.

- **Success envelope**:

```json
{"data": ...}
```

- **Error envelope**:

```json
{"error":{"message":"..."}} 
```

## Endpoints

### UI

- **GET `/`**: UI page
- **GET `/assets/*`**: UI static assets

### Pack sizes

- **GET `/api/packs/`**: list pack sizes

Response:

```json
{"data":{"packs":[{"id":1,"size":250}]}}
```

- **POST `/api/packs/`**: create pack size

Request:

```json
{"size":250}
```

Responses:
- `201` with created pack size: `{"data":{"id":10,"size":250}}`
- `409` if size already exists: `{"error":{"message":"pack size already exists"}}`

- **PUT `/api/packs/{id}`**: update pack size

Request:

```json
{"size":500}
```

Responses:
- `200` with updated pack size: `{"data":{"id":10,"size":500}}`
- `404` if not found: `{"error":{"message":"not found"}}`
- `409` if size already exists: `{"error":{"message":"pack size already exists"}}`

- **DELETE `/api/packs/{id}`**: delete pack size

Response:

```json
{"data":{}}
```

- **POST `/api/packs/reset`**: reset pack sizes to defaults

Response:

```json
{"data":{"sizes":[250,500,1000,2000,5000]}}
```

### Calculate

- **POST `/api/calculate`**: calculate pack allocation

Request:

```json
{"quantity":12001}
```

Response:

```json
{"data":{"packs":[{"size":5000,"count":2},{"size":2000,"count":1},{"size":250,"count":1}]}}
```

Notes:
- Very large quantities are rejected to avoid excessive memory usage:
  - `400` with `{"error":{"message":"quantity too large"}}`

## Run with Docker

### Build

```bash
docker build -t order-packs-calculator .
```

### Run

```bash
docker run --rm -p 8080:8080 \
  -e HTTP_PORT=8080 \
  -e DB_PATH=/app/data/app.db \
  -v order_packs_data:/app/data \
  order-packs-calculator
```

Then open:

- `http://localhost:8080/`

