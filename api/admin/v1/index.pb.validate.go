// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: admin/v1/index.proto

package admin

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

// Validate checks the field values on ActivityDetailRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ActivityDetailRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ActivityDetailRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ActivityDetailRequestMultiError, or nil if none found.
func (m *ActivityDetailRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ActivityDetailRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Code

	// no validation rules for Message

	if len(errors) > 0 {
		return ActivityDetailRequestMultiError(errors)
	}

	return nil
}

// ActivityDetailRequestMultiError is an error wrapping multiple validation
// errors returned by ActivityDetailRequest.ValidateAll() if the designated
// constraints aren't met.
type ActivityDetailRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ActivityDetailRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ActivityDetailRequestMultiError) AllErrors() []error { return m }

// ActivityDetailRequestValidationError is the validation error returned by
// ActivityDetailRequest.Validate if the designated constraints aren't met.
type ActivityDetailRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ActivityDetailRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ActivityDetailRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ActivityDetailRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ActivityDetailRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ActivityDetailRequestValidationError) ErrorName() string {
	return "ActivityDetailRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ActivityDetailRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sActivityDetailRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ActivityDetailRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ActivityDetailRequestValidationError{}

// Validate checks the field values on ActivityDetailResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ActivityDetailResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ActivityDetailResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ActivityDetailResponseMultiError, or nil if none found.
func (m *ActivityDetailResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ActivityDetailResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Code

	// no validation rules for Message

	if len(errors) > 0 {
		return ActivityDetailResponseMultiError(errors)
	}

	return nil
}

// ActivityDetailResponseMultiError is an error wrapping multiple validation
// errors returned by ActivityDetailResponse.ValidateAll() if the designated
// constraints aren't met.
type ActivityDetailResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ActivityDetailResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ActivityDetailResponseMultiError) AllErrors() []error { return m }

// ActivityDetailResponseValidationError is the validation error returned by
// ActivityDetailResponse.Validate if the designated constraints aren't met.
type ActivityDetailResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ActivityDetailResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ActivityDetailResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ActivityDetailResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ActivityDetailResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ActivityDetailResponseValidationError) ErrorName() string {
	return "ActivityDetailResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ActivityDetailResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sActivityDetailResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ActivityDetailResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ActivityDetailResponseValidationError{}
