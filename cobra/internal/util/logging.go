package util

import (
	"context"
	"fmt"
	"time"
)

type VerboseLogger struct {
	IsVerbose bool
}

func NewVerboseLogger(ctx context.Context) *VerboseLogger {
	v, ok := ctx.Value("verbose").(bool)
	return &VerboseLogger{
		IsVerbose: ok && v,
	}
}

func (l *VerboseLogger) Print(format string, args ...interface{}) {
	if l.IsVerbose {
		fmt.Printf(format, args...)
	}
}

func (l *VerboseLogger) LogOperation(operation string, fn func() error) error {
	if !l.IsVerbose {
		return fn()
	}

	l.Print("Starting %s at %v\n", operation, time.Now().Format(time.RFC3339))
	start := time.Now()
	err := fn()
	elapsed := time.Since(start)

	if err != nil {
		l.Print("%s failed after %v: %v\n", operation, elapsed, err)
	} else {
		l.Print("%s completed in %v\n", operation, elapsed)
	}

	return err
}
