package mcp

import (
	"errors"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

const (
	errCodeArtifactSelectionMissing errs.Code = "workbench.artifact.selection_missing"
	errCodeContextPatchInvalid      errs.Code = "workbench.context.patch.invalid"
	errCodePlannerUnavailable       errs.Code = "workbench.planner.unavailable"
	errCodeResourceURIInvalid       errs.Code = "workbench.resource.uri_invalid"
)

func defaultPublicTitle(sentinel error, code errs.Code) string {
	switch code {
	case artifacts.CodeNotFound:
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
