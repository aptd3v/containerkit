// Package model holds the Compose models e2e test. It brings up a project
// through pkg/kit AddModel/Model, binds Google EmbeddingGemma via Docker
// Model Runner, and calls the OpenAI-compatible embeddings API.
//
// Default `go test ./...` does not run these tests. Against a live Engine
// with Docker Model Runner enabled:
//
//	go test -tags e2e -count=1 ./e2e/model
package model
