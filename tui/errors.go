package tui

import (
	"errors"
	"fmt"
)

// ErrFileExists reports that an export would overwrite a file that is already
// there.  Export functions return it wrapped with the destination path when
// called with overwrite=false, so the caller can ask before clobbering.
var ErrFileExists = errors.New("file already exists")

// ExistsError names the destination that blocked an export.
type ExistsError struct {
	Path string
}

func (e *ExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.Path)
}

// Unwrap lets errors.Is(err, ErrFileExists) match.
func (e *ExistsError) Unwrap() error { return ErrFileExists }

// ExportError represents an error during export operation.
//
// This structured error type provides context about which stage of the export
// failed (validation, file writing, or config update) and which path was involved.
//
// Example usage:
//
//	if err != nil {
//	    var exportErr *ExportError
//	    if errors.As(err, &exportErr) {
//	        log.Printf("Export failed at %s stage for %s", exportErr.Operation, exportErr.Path)
//	    }
//	}
type ExportError struct {
	// Operation indicates which stage failed: "validate", "write", or "config"
	Operation string
	// Path is the file or directory path that was being accessed
	Path string
	// Err is the underlying error that caused the failure
	Err error
}

// Error implements the error interface, providing a formatted error message.
func (e *ExportError) Error() string {
	return fmt.Sprintf("export %s failed [%s]: %v", e.Operation, e.Path, e.Err)
}

// Unwrap returns the underlying error, enabling error chain inspection with errors.Is and errors.As.
func (e *ExportError) Unwrap() error {
	return e.Err
}

// ValidationError represents a filename validation failure.
//
// This error type is used specifically for filename validation issues,
// providing detailed information about what was invalid.
type ValidationError struct {
	// Filename is the invalid filename that was rejected
	Filename string
	// Reason explains why the filename was rejected
	Reason string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid filename %q: %s", e.Filename, e.Reason)
}
