package logger

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(text string) Level {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	level  Level
	fields map[string]string
}

func New(out io.Writer, level Level) *Logger {
	if out == nil {
		out = os.Stderr
	}
	return &Logger{out: out, level: level, fields: make(map[string]string)}
}

func Default() *Logger { return New(os.Stderr, LevelInfo) }

func (l *Logger) With(kv ...string) *Logger {
	child := &Logger{out: l.out, level: l.level, fields: make(map[string]string, len(l.fields)+len(kv)/2)}
	for k, v := range l.fields {
		child.fields[k] = v
	}
	for i := 0; i+1 < len(kv); i += 2 {
		child.fields[kv[i]] = kv[i+1]
	}
	return child
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) Debug(msg string, kv ...string) { l.log(LevelDebug, msg, kv) }
func (l *Logger) Info(msg string, kv ...string)  { l.log(LevelInfo, msg, kv) }
func (l *Logger) Warn(msg string, kv ...string)  { l.log(LevelWarn, msg, kv) }
func (l *Logger) Error(msg string, kv ...string) { l.log(LevelError, msg, kv) }

func (l *Logger) log(level Level, msg string, kv []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}
	merged := make(map[string]string, len(l.fields)+len(kv)/2)
	for k, v := range l.fields {
		merged[k] = v
	}
	for i := 0; i+1 < len(kv); i += 2 {
		merged[kv[i]] = kv[i+1]
	}
	var sb strings.Builder
	sb.WriteString(time.Now().UTC().Format(time.RFC3339))
	sb.WriteByte(' ')
	sb.WriteString(level.String())
	sb.WriteByte(' ')
	sb.WriteString(msg)
	if len(merged) > 0 {
		keys := make([]string, 0, len(merged))
		for k := range merged {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteByte(' ')
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(quote(merged[k]))
		}
	}
	sb.WriteByte('\n')
	_, _ = io.WriteString(l.out, sb.String())
}

func quote(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, " \t\n\"") {
		return fmt.Sprintf("%q", v)
	}
	return v
}
