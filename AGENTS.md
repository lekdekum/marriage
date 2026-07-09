# Project Guidelines

## Purpose

This repository contains the backend for a marriage website. It should also be suitable as a portfolio project, so favor clear structure, readable code, tests for important behavior, and simple operational practices.

## Architecture

- Use Go for backend code.
- Keep the app close to a simple MVC structure:
  - `cmd/api`: executable entrypoint and server wiring.
  - `internal/routes`: HTTP route registration.
  - `internal/controllers`: request handling, validation, and response mapping.
  - `internal/services`: business logic and orchestration.
  - `internal/response`: shared HTTP response helpers.
- Keep controllers thin. Put business behavior in services.
- Keep dependencies minimal. Add external packages only when they solve a concrete problem better than the standard library.

## Development

- Run the API with `go run ./cmd/api`.
- Run tests with `go test ./...`.
- Format Go code with `gofmt`.
- Use `PORT` to override the local server port. The default is `8080`.

## API Conventions

- Return JSON responses.
- Prefer explicit HTTP methods in route definitions.
- Use predictable, resource-oriented paths.
- Keep health and readiness endpoints lightweight.

## Code Quality

- Add focused tests for new routes, services, and non-trivial behavior.
- Avoid broad refactors while implementing small features.
- Keep configuration in environment variables until the project needs a richer config layer.
- Do not commit secrets, credentials, or generated build artifacts.
