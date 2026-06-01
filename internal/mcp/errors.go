package mcp

import (
	"errors"

	"github.com/manuelibar/workbench/internal/errs"
)

const (
	errCodeArtifactDirInvalid       errs.Code = "workbench.artifact.dir_invalid"
	errCodeArtifactInvalid          errs.Code = "workbench.artifact.invalid"
	errCodeArtifactIDInvalid        errs.Code = "workbench.artifact.id_invalid"
	errCodeArtifactListFailed       errs.Code = "workbench.artifact.list_failed"
	errCodeArtifactNotFound         errs.Code = "workbench.artifact.not_found"
	errCodeArtifactReadFailed       errs.Code = "workbench.artifact.read_failed"
	errCodeArtifactSelectionMissing errs.Code = "workbench.artifact.selection_missing"
	errCodeArtifactStoreUnavailable errs.Code = "workbench.artifact.store_unavailable"
	errCodeArtifactTypeUnknown      errs.Code = "workbench.artifact.type_unknown"
	errCodeArtifactWriteFailed      errs.Code = "workbench.artifact.write_failed"
	errCodeContextPatchInvalid      errs.Code = "workbench.context.patch.invalid"
	errCodePlannerUnavailable       errs.Code = "workbench.planner.unavailable"
	errCodeResourceURIInvalid       errs.Code = "workbench.resource.uri_invalid"
)

func unknownArtifactTypeError(operation, typ string) error {
	attrs := map[string]any{
		"operation":     operation,
		"artifact_type": typ,
	}
	return errs.New(
		"Unknown artifact type",
		errs.WithSentinel(errs.ErrInvalid),
		errs.WithCode(errCodeArtifactTypeUnknown),
		errs.WithSeverity(errs.SeverityWarning),
		errs.WithAttrs(attrs),
		errs.WithRetryable(false),
	)
}

func artifactWriteError(id, operation string, cause error) error {
	attrs := map[string]any{
		"operation":   operation,
		"artifact_id": id,
	}
	return errs.New(
		"Artifact write failed",
		errs.WithSentinel(errs.ErrDependencyFailed),
		errs.WithCode(errCodeArtifactWriteFailed),
		errs.WithSeverity(errs.SeverityError),
		errs.WithCause(cause),
		errs.WithAttrs(attrs),
		errs.WithRetryable(false),
	)
}

func defaultPublicTitle(sentinel error, code errs.Code) string {
	switch code {
	case errCodeArtifactNotFound:
		return "Artifact not found"
	case errCodeArtifactSelectionMissing:
		return "Artifact selection required"
	case errCodeContextPatchInvalid:
		return "Context patch is invalid"
	case errCodeResourceURIInvalid:
		return "Resource URI is invalid"
	}
	switch {
	case errors.Is(sentinel, errs.ErrInvalid):
		return "Invalid request"
	case errors.Is(sentinel, errs.ErrNotFound):
		return "Not found"
	case errors.Is(sentinel, errs.ErrUnavailable):
		return "Workbench unavailable"
	case errors.Is(sentinel, errs.ErrTimeout):
		return "Workbench timed out"
	default:
		return "Internal error"
	}
}

func defaultPublicCode(sentinel error) errs.Code {
	switch {
	case errors.Is(sentinel, errs.ErrInvalid):
		return "workbench.invalid"
	case errors.Is(sentinel, errs.ErrNotFound):
		return "workbench.not_found"
	case errors.Is(sentinel, errs.ErrUnavailable):
		return "workbench.unavailable"
	case errors.Is(sentinel, errs.ErrTimeout):
		return "workbench.timeout"
	default:
		return "workbench.internal"
	}
}
