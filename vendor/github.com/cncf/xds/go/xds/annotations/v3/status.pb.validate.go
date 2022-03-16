// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: xds/annotations/v3/status.proto

package v3

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
)

// Validate checks the field values on FileStatusAnnotation with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *FileStatusAnnotation) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for WorkInProgress

	return nil
}

// FileStatusAnnotationValidationError is the validation error returned by
// FileStatusAnnotation.Validate if the designated constraints aren't met.
type FileStatusAnnotationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e FileStatusAnnotationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e FileStatusAnnotationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e FileStatusAnnotationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e FileStatusAnnotationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e FileStatusAnnotationValidationError) ErrorName() string {
	return "FileStatusAnnotationValidationError"
}

// Error satisfies the builtin error interface
func (e FileStatusAnnotationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sFileStatusAnnotation.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = FileStatusAnnotationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = FileStatusAnnotationValidationError{}

// Validate checks the field values on MessageStatusAnnotation with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *MessageStatusAnnotation) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for WorkInProgress

	return nil
}

// MessageStatusAnnotationValidationError is the validation error returned by
// MessageStatusAnnotation.Validate if the designated constraints aren't met.
type MessageStatusAnnotationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MessageStatusAnnotationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MessageStatusAnnotationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MessageStatusAnnotationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MessageStatusAnnotationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MessageStatusAnnotationValidationError) ErrorName() string {
	return "MessageStatusAnnotationValidationError"
}

// Error satisfies the builtin error interface
func (e MessageStatusAnnotationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMessageStatusAnnotation.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MessageStatusAnnotationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MessageStatusAnnotationValidationError{}

// Validate checks the field values on FieldStatusAnnotation with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *FieldStatusAnnotation) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for WorkInProgress

	return nil
}

// FieldStatusAnnotationValidationError is the validation error returned by
// FieldStatusAnnotation.Validate if the designated constraints aren't met.
type FieldStatusAnnotationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e FieldStatusAnnotationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e FieldStatusAnnotationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e FieldStatusAnnotationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e FieldStatusAnnotationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e FieldStatusAnnotationValidationError) ErrorName() string {
	return "FieldStatusAnnotationValidationError"
}

// Error satisfies the builtin error interface
func (e FieldStatusAnnotationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sFieldStatusAnnotation.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = FieldStatusAnnotationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = FieldStatusAnnotationValidationError{}

// Validate checks the field values on StatusAnnotation with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *StatusAnnotation) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for WorkInProgress

	// no validation rules for PackageVersionStatus

	return nil
}

// StatusAnnotationValidationError is the validation error returned by
// StatusAnnotation.Validate if the designated constraints aren't met.
type StatusAnnotationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e StatusAnnotationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e StatusAnnotationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e StatusAnnotationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e StatusAnnotationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e StatusAnnotationValidationError) ErrorName() string { return "StatusAnnotationValidationError" }

// Error satisfies the builtin error interface
func (e StatusAnnotationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sStatusAnnotation.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = StatusAnnotationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = StatusAnnotationValidationError{}
