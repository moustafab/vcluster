package log

import (
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/klog"
	"strings"
)

type WithDepth interface {
	WithDepth(depth int) logr.Logger
}

func NewLog(level int) logr.Logger {
	return &log{
		level: level,
		depth: 1,
	}
}

type log struct {
	current  int
	level    int
	prefixes []string
	depth    int
}

func (l *log) WithDepth(depth int) logr.Logger {
	return &log{
		level:    l.level,
		current:  l.current,
		prefixes: l.prefixes,
		depth:    depth,
	}
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (l *log) Info(msg string, keysAndValues ...interface{}) {
	klog.InfoDepth(l.depth, l.formatMsg(msg, keysAndValues...))
}

// Enabled tests whether this InfoLogger is enabled.  For example,
// commandline flags might be used to set the logging verbosity and disable
// some info logs.
func (l *log) Enabled() bool {
	return true
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func (l *log) Error(err error, msg string, keysAndValues ...interface{}) {
	newKeysAndValues := []interface{}{err}
	newKeysAndValues = append(newKeysAndValues, keysAndValues...)
	klog.ErrorDepth(l.depth, l.formatMsg(msg, newKeysAndValues...))
}

// V returns an InfoLogger value for a specific verbosity level.  A higher
// verbosity level means a log message is less important.  It's illegal to
// pass a log level less than zero.
func (l *log) V(level int) logr.Logger {
	if level < l.level {
		return &silent{}
	}

	prefixes := []string{}
	prefixes = append(prefixes, l.prefixes...)
	return &log{
		level:    l.level,
		current:  level,
		prefixes: prefixes,
		depth:    l.depth,
	}
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (l *log) WithValues(keysAndValues ...interface{}) logr.Logger {
	prefixes := []string{}
	prefixes = append(prefixes, l.prefixes...)
	prefixes = append(prefixes, formatKeysAndValues(keysAndValues...))

	return &log{
		level:    l.level,
		current:  l.current,
		prefixes: prefixes,
		depth:    l.depth,
	}
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly reccomended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (l *log) WithName(name string) logr.Logger {
	if name == "" {
		return &log{
			level:    l.level,
			current:  l.current,
			prefixes: l.prefixes,
			depth:    l.depth,
		}
	}

	prefixes := []string{}
	prefixes = append(prefixes, l.prefixes...)
	prefixes = append(prefixes, name)

	return &log{
		level:    l.level,
		current:  l.current,
		prefixes: prefixes,
		depth:    l.depth,
	}
}

func (l *log) formatMsg(msg string, keysAndValues ...interface{}) string {
	prefixes := strings.Join(l.prefixes, ": ")
	addString := formatKeysAndValues(keysAndValues...)

	retString := msg
	if prefixes != "" {
		retString = prefixes + ": " + retString
	}
	if addString != "" {
		retString += " " + addString
	}
	// if l.current != 0 {
	//	retString = "(" + strconv.Itoa(l.current) + ") " + retString
	// }
	return retString
}

func formatKeysAndValues(keysAndValues ...interface{}) string {
	args := []string{}
	for _, kv := range keysAndValues {
		switch t := kv.(type) {
		case string:
			args = append(args, t)
		case error:
			args = append(args, t.Error())
		default:
			args = append(args, fmt.Sprintf("%#v", kv))
		}
	}

	return strings.Join(args, " ")
}

type silent struct{}

func (s *silent) Info(msg string, keysAndValues ...interface{})             {}
func (s *silent) Enabled() bool                                             { return false }
func (s *silent) Error(err error, msg string, keysAndValues ...interface{}) {}
func (s *silent) V(level int) logr.Logger                                   { return s }
func (s *silent) WithValues(keysAndValues ...interface{}) logr.Logger       { return s }
func (s *silent) WithName(name string) logr.Logger                          { return s }
