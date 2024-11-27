package diagutils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewClientError(action string, err error) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Client error", fmt.Sprintf("Unable to %s, got error: %s", action, err))
}

func NewNotFoundError(resource string) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Not found", fmt.Sprintf("No matching %s found", resource))
}

func NewNotSupportedError(action string) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Not supported", fmt.Sprintf("Action %q is not supported", action))
}

func NewFillError(err error) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Fill error", fmt.Sprintf("Unable to fill model: %s", err))
}

func NewImportError(err error) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Import error", fmt.Sprintf("Unable to import: %s", err))
}
