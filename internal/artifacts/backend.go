package artifacts

import (
	"context"
	"errors"
)

var errBackendNotFound = errors.New("artifact backend: not found")

type storedMarkdown struct {
	ID       string
	Location string
	Markdown string
}

type markdownBackend interface {
	Root() string
	Location(string) string
	List(context.Context) ([]storedMarkdown, error)
	Read(context.Context, string) (storedMarkdown, error)
	Write(context.Context, string, string) (storedMarkdown, error)
}
