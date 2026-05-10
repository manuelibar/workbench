// Package pgstore is the Postgres-backed persistence layer for workbench-mcp.
//
// It exposes a [Store] type whose methods are organised one file per
// resource (users, work_sessions, namespaces, ...). All schema creation is
// driven by embedded SQL files under migrations/, applied in numeric order
// by [Store.Migrate]; the runner is hand-rolled (no goose) to keep the
// dependency footprint small.
//
// Connection management uses [pgxpool] with [Store.Open] returning a fully
// configured pool; callers should defer [Store.Close]. [Store.WithTx] is the
// canonical way to run multi-statement work inside a transaction.
package pgstore
