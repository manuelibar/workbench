package errs_test

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/manuelibar/workbench/internal/errs"
)

type customCause struct{}

func (*customCause) Error() string {
	return "custom cause"
}

func TestErrorsIsSentinelThroughErrorAndMulti(t *testing.T) {
	leaf := errs.New("Lookup user", errs.WithSentinel(errs.ErrNotFound))
	wrapped := errs.New("Load profile", errs.WithCause(leaf))

	if !errors.Is(wrapped, errs.ErrNotFound) {
		t.Fatal("errors.Is did not match sentinel through Error cause")
	}

	multi := errs.NewMulti(
		"Batch failed",
		errs.WithSentinel(errs.ErrInvalid),
		errs.WithCause(io.EOF),
		errs.WithCause(leaf),
	)

	if !errors.Is(multi, errs.ErrInvalid) {
		t.Fatal("errors.Is did not match Multi sentinel")
	}
	if !errors.Is(multi, errs.ErrNotFound) {
		t.Fatal("errors.Is did not match sentinel through Multi child")
	}
	if errors.Unwrap(multi) != nil {
		t.Fatal("errors.Unwrap should not expose Multi children; use errors.Is/As")
	}
}

func TestErrorsAsForTypesAndWrappedCause(t *testing.T) {
	cause := &customCause{}
	err := errs.New("Operation failed", errs.WithCause(cause))

	var gotErr *errs.Error
	if !errors.As(err, &gotErr) {
		t.Fatal("errors.As did not find *errs.Error")
	}
	if gotErr != err {
		t.Fatal("errors.As returned unexpected *errs.Error")
	}

	var gotCause *customCause
	if !errors.As(err, &gotCause) {
		t.Fatal("errors.As did not find wrapped standard cause")
	}

	multi := errs.NewMulti("Many", errs.WithCause(err))
	var gotMulti *errs.Multi
	if !errors.As(multi, &gotMulti) {
		t.Fatal("errors.As did not find *errs.Multi")
	}
	if gotMulti != multi {
		t.Fatal("errors.As returned unexpected *errs.Multi")
	}
}

func TestMultiAggregatesIncrementallyWithStandardUnwrap(t *testing.T) {
	multi := errs.NewMulti("Batch failed", errs.WithSentinel(errs.ErrInvalid))
	if multi.Len() != 0 {
		t.Fatalf("empty Multi length = %d", multi.Len())
	}

	first := errs.New("Missing row", errs.WithSentinel(errs.ErrNotFound))
	second := errors.New("permission denied")
	multi.Add(nil, first)
	multi.Add(second)

	if multi.Len() != 2 {
		t.Fatalf("Multi length = %d", multi.Len())
	}
	if !errors.Is(multi, errs.ErrInvalid) {
		t.Fatal("errors.Is did not match Multi sentinel")
	}
	if !errors.Is(multi, errs.ErrNotFound) {
		t.Fatal("errors.Is did not inspect Multi children")
	}
	if !errors.Is(multi, second) {
		t.Fatal("errors.Is did not match standard child")
	}
	if errors.Unwrap(multi) != nil {
		t.Fatal("errors.Unwrap should not expose Multi children; use errors.Is/As")
	}

	children := multi.Unwrap()
	if len(children) != 2 {
		t.Fatalf("children length = %d", len(children))
	}
	children[0] = nil
	if multi.Unwrap()[0] != first {
		t.Fatal("Multi.Unwrap output was not copied")
	}
}

func TestDecoratePreservesChainStackAndDoesNotMutate(t *testing.T) {
	base := errs.New(
		"Load widget",
		errs.WithSentinel(errs.ErrNotFound),
		errs.WithAttrs(map[string]any{"component": "repo"}),
	)
	baseStack := errs.StackOf(base)

	decorated := errs.Decorate(
		base,
		errs.WithCode("widget.not_found"),
		errs.WithSeverity(errs.SeverityWarning),
		errs.WithAttrs(map[string]any{"operation": "GetWidget"}),
	)

	if decorated == base {
		t.Fatal("Decorate should return a non-mutating copy for *errs.Error")
	}
	if !errors.Is(decorated, errs.ErrNotFound) {
		t.Fatal("decorated error lost sentinel matching")
	}
	if errs.CodeOf(decorated) != "widget.not_found" {
		t.Fatalf("decorated code = %q", errs.CodeOf(decorated))
	}
	if errs.SeverityOf(decorated) != errs.SeverityWarning {
		t.Fatalf("decorated severity = %q", errs.SeverityOf(decorated))
	}

	baseAttrs := errs.AttrsOf(base)
	if _, ok := baseAttrs["operation"]; ok {
		t.Fatal("Decorate mutated original attrs")
	}
	decoratedAttrs := errs.AttrsOf(decorated)
	if decoratedAttrs["component"] != "repo" || decoratedAttrs["operation"] != "GetWidget" {
		t.Fatalf("decorated attrs = %#v", decoratedAttrs)
	}

	if !reflect.DeepEqual(errs.StackOf(decorated), baseStack) {
		t.Fatal("Decorate should preserve stack for existing *errs.Error")
	}

	standard := errors.New("disk full")
	decoratedStandard := errs.Decorate(standard, errs.WithRetryable(true))
	if !errors.Is(decoratedStandard, standard) {
		t.Fatal("Decorate should wrap standard errors as the cause")
	}
	if !errs.IsRetryable(decoratedStandard) {
		t.Fatal("decorated standard error lost retryable metadata")
	}
	if len(errs.StackOf(decoratedStandard)) == 0 {
		t.Fatal("Decorate should capture stack for standard errors")
	}
}

func TestDecorateCanApplyZeroValueMetadata(t *testing.T) {
	base := errs.New(
		"Operation",
		errs.WithSentinel(errs.ErrUnavailable),
		errs.WithCode("operation.failed"),
		errs.WithSeverity(errs.SeverityCritical),
		errs.WithRetryable(true),
	)

	decorated := errs.Decorate(
		base,
		errs.WithSentinel(errs.ErrInvalid),
		errs.WithCode(""),
		errs.WithSeverity(""),
		errs.WithRetryable(false),
	)

	if !errors.Is(decorated, errs.ErrInvalid) {
		t.Fatal("Decorate did not patch sentinel metadata")
	}
	if errors.Is(decorated, errs.ErrUnavailable) {
		t.Fatal("Decorate kept old sentinel metadata")
	}
	if errs.CodeOf(decorated) != "" {
		t.Fatalf("CodeOf = %q", errs.CodeOf(decorated))
	}
	if errs.SeverityOf(decorated) != "" {
		t.Fatalf("SeverityOf = %q", errs.SeverityOf(decorated))
	}
	if errs.IsRetryable(decorated) {
		t.Fatal("IsRetryable = true")
	}
}

func TestAttrsAndSlicesAreCopiedOnInputAndOutput(t *testing.T) {
	inputAttrs := map[string]any{"k": "v"}
	err := errs.New("Copy", errs.WithAttrs(inputAttrs))
	inputAttrs["k"] = "changed"

	gotAttrs := errs.AttrsOf(err)
	if gotAttrs["k"] != "v" {
		t.Fatalf("attrs were not copied on input: %#v", gotAttrs)
	}
	gotAttrs["k"] = "mutated"
	if errs.AttrsOf(err)["k"] != "v" {
		t.Fatalf("attrs were not copied on output: %#v", errs.AttrsOf(err))
	}

	stack := errs.StackOf(err)
	if len(stack) == 0 {
		t.Fatal("expected captured stack")
	}
	originalTop := stack[0]
	stack[0] = errs.Frame{}
	if reflect.DeepEqual(errs.StackOf(err)[0], errs.Frame{}) {
		t.Fatal("stack output was not copied")
	}
	if !reflect.DeepEqual(errs.StackOf(err)[0], originalTop) {
		t.Fatal("stack changed unexpectedly")
	}

	child := errs.New("Child")
	multi := errs.NewMulti("Many", errs.WithCause(nil), errs.WithCause(child))
	children := multi.Unwrap()
	if len(children) != 1 {
		t.Fatalf("children length = %d", len(children))
	}
	children[0] = nil
	if multi.Unwrap()[0] != child {
		t.Fatal("Multi.Unwrap output was not copied")
	}
}

func TestLookupHelpersTraverseNativeTree(t *testing.T) {
	err := fmt.Errorf("outer: %w", errs.New(
		"Inner",
		errs.WithSentinel(errs.ErrUnavailable),
		errs.WithCode("dependency.unavailable"),
		errs.WithSeverity(errs.SeverityCritical),
		errs.WithRetryable(true),
		errs.WithAttrs(map[string]any{"service": "inventory"}),
	))

	if errs.SentinelOf(err) != errs.ErrUnavailable {
		t.Fatalf("SentinelOf = %v", errs.SentinelOf(err))
	}
	if errs.CodeOf(err) != "dependency.unavailable" {
		t.Fatalf("CodeOf = %q", errs.CodeOf(err))
	}
	if errs.SeverityOf(err) != errs.SeverityCritical {
		t.Fatalf("SeverityOf = %q", errs.SeverityOf(err))
	}
	if !errs.IsRetryable(err) {
		t.Fatal("IsRetryable = false")
	}
	if attrs := errs.AttrsOf(err); attrs["service"] != "inventory" {
		t.Fatalf("AttrsOf = %#v", attrs)
	}
	if len(errs.StackOf(err)) == 0 {
		t.Fatal("StackOf was empty")
	}
}

func TestAttrsCloserToRootOverrideDeeperAttrs(t *testing.T) {
	cause := errs.New("Cause", errs.WithAttrs(map[string]any{"key": "cause", "cause": true}))
	root := errs.New("Root", errs.WithCause(cause), errs.WithAttrs(map[string]any{"key": "root"}))

	attrs := errs.AttrsOf(root)
	if attrs["key"] != "root" {
		t.Fatalf("root attr did not override cause attr: %#v", attrs)
	}
	if attrs["cause"] != true {
		t.Fatalf("cause attr was not preserved: %#v", attrs)
	}
}
