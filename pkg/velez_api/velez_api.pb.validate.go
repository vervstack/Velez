// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: grpc/velez_api.proto

package velez_api

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
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
	_ = sort.Sort
)

// Validate checks the field values on Version with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Version) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Version with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in VersionMultiError, or nil if none found.
func (m *Version) ValidateAll() error {
	return m.validate(true)
}

func (m *Version) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return VersionMultiError(errors)
	}

	return nil
}

// VersionMultiError is an error wrapping multiple validation errors returned
// by Version.ValidateAll() if the designated constraints aren't met.
type VersionMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m VersionMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m VersionMultiError) AllErrors() []error { return m }

// VersionValidationError is the validation error returned by Version.Validate
// if the designated constraints aren't met.
type VersionValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VersionValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VersionValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VersionValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VersionValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VersionValidationError) ErrorName() string { return "VersionValidationError" }

// Error satisfies the builtin error interface
func (e VersionValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVersion.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VersionValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VersionValidationError{}

// Validate checks the field values on Version_Request with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *Version_Request) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Version_Request with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// Version_RequestMultiError, or nil if none found.
func (m *Version_Request) ValidateAll() error {
	return m.validate(true)
}

func (m *Version_Request) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return Version_RequestMultiError(errors)
	}

	return nil
}

// Version_RequestMultiError is an error wrapping multiple validation errors
// returned by Version_Request.ValidateAll() if the designated constraints
// aren't met.
type Version_RequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m Version_RequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m Version_RequestMultiError) AllErrors() []error { return m }

// Version_RequestValidationError is the validation error returned by
// Version_Request.Validate if the designated constraints aren't met.
type Version_RequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Version_RequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Version_RequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Version_RequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Version_RequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Version_RequestValidationError) ErrorName() string { return "Version_RequestValidationError" }

// Error satisfies the builtin error interface
func (e Version_RequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVersion_Request.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Version_RequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Version_RequestValidationError{}

// Validate checks the field values on Version_Response with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *Version_Response) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Version_Response with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// Version_ResponseMultiError, or nil if none found.
func (m *Version_Response) ValidateAll() error {
	return m.validate(true)
}

func (m *Version_Response) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Version

	if len(errors) > 0 {
		return Version_ResponseMultiError(errors)
	}

	return nil
}

// Version_ResponseMultiError is an error wrapping multiple validation errors
// returned by Version_Response.ValidateAll() if the designated constraints
// aren't met.
type Version_ResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m Version_ResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m Version_ResponseMultiError) AllErrors() []error { return m }

// Version_ResponseValidationError is the validation error returned by
// Version_Response.Validate if the designated constraints aren't met.
type Version_ResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Version_ResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Version_ResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Version_ResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Version_ResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Version_ResponseValidationError) ErrorName() string { return "Version_ResponseValidationError" }

// Error satisfies the builtin error interface
func (e Version_ResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVersion_Response.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Version_ResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Version_ResponseValidationError{}
