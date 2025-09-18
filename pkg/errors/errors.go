package errors

import (
	"github.com/cockroachdb/errors"
)

// Re-export commonly used functions from cockroachdb/errors
var (
	// Creation
	New            = errors.New
	Newf           = errors.Newf
	Errorf         = errors.Errorf
	WithMessage    = errors.WithMessage
	WithMessagef   = errors.WithMessagef
	WithDetail     = errors.WithDetail
	WithDetailf    = errors.WithDetailf
	WithHint       = errors.WithHint
	WithHintf      = errors.WithHintf
	WithStack      = errors.WithStack
	WithStackDepth = errors.WithStackDepth

	// Wrapping
	Wrap          = errors.Wrap
	Wrapf         = errors.Wrapf
	WrapWithDepth = errors.WrapWithDepth

	// Inspection
	Is     = errors.Is
	As     = errors.As
	Unwrap = errors.Unwrap
	Cause  = errors.Cause

	// Formatting
	Redact      = errors.Redact
	ReportError = errors.ReportError

	// Assertions
	AssertionFailedf                 = errors.AssertionFailedf
	AssertionFailedWithDepthf        = errors.AssertionFailedWithDepthf
	NewAssertionErrorWithWrappedErrf = errors.NewAssertionErrorWithWrappedErrf

	// Handled errors
	Handled                    = errors.Handled
	HandledWithMessage         = errors.HandledWithMessage
	HandledInDomain            = errors.HandledInDomain
	HandledInDomainWithMessage = errors.HandledInDomainWithMessage

	// Safe formatting
	Safe = errors.Safe
)

// Domain-specific error types
type (
	// NotFoundError indicates a resource was not found
	NotFoundError struct {
		Resource string
	}

	// ValidationError indicates invalid input
	ValidationError struct {
		Field   string
		Message string
	}

	// TestExecutionError indicates a test execution failure
	TestExecutionError struct {
		Package string
		Test    string
		Reason  string
	}
)

// Error implementations
func (e *NotFoundError) Error() string {
	return errors.Newf("%s not found", e.Resource).Error()
}

func (e *ValidationError) Error() string {
	return errors.Newf("validation failed for %s: %s", e.Field, e.Message).Error()
}

func (e *TestExecutionError) Error() string {
	return errors.Newf("test execution failed: package=%s test=%s reason=%s",
		e.Package, e.Test, e.Reason).Error()
}

// Helper constructors
func NotFound(resource string) error {
	return &NotFoundError{Resource: resource}
}

func Invalid(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

func TestFailed(pkg, test, reason string) error {
	return &TestExecutionError{Package: pkg, Test: test, Reason: reason}
}

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

// IsValidation checks if an error is a ValidationError
func IsValidation(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// IsTestExecution checks if an error is a TestExecutionError
func IsTestExecution(err error) bool {
	var tee *TestExecutionError
	return errors.As(err, &tee)
}
