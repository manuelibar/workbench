// Package config loads workbench-mcp runtime configuration from environment
// variables and returns a typed [Config]. It is the only place in the binary
// that reads the process environment.
//
// Defaults are chosen so that workbench-mcp can be started against the
// docker-compose-managed Postgres without setting any environment variables.
// Override individual fields by setting WORKBENCH_BIND, WORKBENCH_DB_URL, and
// WORKBENCH_LOG_LEVEL. See [Load].
package config
