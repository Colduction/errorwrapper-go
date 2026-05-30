package errorwrapper

import (
	"errors"
	"strings"
)

const (
	defaultErrJoiner byte   = '.'
	defaultMsgJoiner string = ": "
)

// errWrapper holds context for building formatted error messages.
type errWrapper struct {
	err       error
	msg       string
	prefix    string
	errJoiner byte
}

// ErrorWrapper creates and wraps errors with prefix chains.
type ErrorWrapper interface {
	// NewError wraps err with the wrapper's prefix and an optional message.
	NewError(err error, msg ...string) error
	// NewErrorString creates a new error from errStr and wraps it.
	NewErrorString(errStr string, msg ...string) error
}

var _ ErrorWrapper = (*errWrapper)(nil)

// New returns an ErrorWrapper with the given joiner byte and optional prefix.
// If errJoiner is zero, the default '.' is used.
func New(errJoiner byte, prefix ...string) ErrorWrapper {
	ew := &errWrapper{errJoiner: errJoiner}
	if ew.errJoiner == 0 {
		ew.errJoiner = defaultErrJoiner
	}
	if len(prefix) >= 1 {
		ew.prefix = prefix[0]
	}
	return ew
}

// collectPrefixes walks the errWrapper chain rooted at err, returning the
// joiner-separated prefix string and the first non-errWrapper error.
func collectPrefixes(err error, joiner byte) (string, error) {
	ew, ok := err.(*errWrapper)
	if !ok {
		return "", err
	}
	if _, ok = ew.err.(*errWrapper); !ok {
		return ew.prefix, ew.err
	}
	var (
		sb    strings.Builder
		first = true
	)
	for {
		if ew.prefix != "" {
			if !first {
				sb.WriteByte(joiner)
			}
			sb.WriteString(ew.prefix)
			first = false
		}
		err = ew.err
		ew, ok = err.(*errWrapper)
		if !ok {
			break
		}
	}
	return sb.String(), err
}

// NewError wraps err with the receiver's prefix, merging any errWrapper chain.
func (ew *errWrapper) NewError(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	var tmpMsg string
	if len(msg) >= 1 {
		tmpMsg = msg[0]
	}
	var (
		mergedPrefix            string
		combinedPrefix, rootErr = collectPrefixes(err, ew.errJoiner)
	)
	switch {
	case ew.prefix == "":
		mergedPrefix = combinedPrefix
	case combinedPrefix == "":
		mergedPrefix = ew.prefix
	default:
		var sb strings.Builder
		sb.Grow(len(ew.prefix) + 1 + len(combinedPrefix))
		sb.WriteString(ew.prefix)
		sb.WriteByte(ew.errJoiner)
		sb.WriteString(combinedPrefix)
		mergedPrefix = sb.String()
	}
	return &errWrapper{
		prefix:    mergedPrefix,
		err:       rootErr,
		msg:       tmpMsg,
		errJoiner: ew.errJoiner,
	}
}

// NewErrorString creates a new error from errStr and wraps it with the receiver's prefix.
func (ew *errWrapper) NewErrorString(errStr string, msg ...string) error {
	if errStr == "" {
		return nil
	}
	var tmpMsg string
	if len(msg) >= 1 {
		tmpMsg = msg[0]
	}
	return &errWrapper{
		prefix:    ew.prefix,
		err:       errors.New(errStr),
		msg:       tmpMsg,
		errJoiner: ew.errJoiner,
	}
}

// Error formats the error as "prefix: [msg] err".
func (ew *errWrapper) Error() string {
	var errStr string
	if ew.err != nil {
		errStr = ew.err.Error()
	}
	var (
		isMsgFilled = ew.msg != ""
		total       = len(errStr)
	)
	if ew.prefix != "" {
		total += len(ew.prefix) + len(defaultMsgJoiner)
	}
	if isMsgFilled {
		total += 2 + len(ew.msg)
		if errStr != "" {
			total++
		}
	}
	var sb strings.Builder
	sb.Grow(total)
	if ew.prefix != "" {
		sb.WriteString(ew.prefix)
		sb.WriteString(defaultMsgJoiner)
	}
	if isMsgFilled {
		sb.WriteByte('[')
		sb.WriteString(ew.msg)
		sb.WriteByte(']')
		if errStr != "" {
			sb.WriteByte(' ')
		}
	}
	sb.WriteString(errStr)
	return sb.String()
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (ew *errWrapper) Unwrap() error {
	return ew.err
}
