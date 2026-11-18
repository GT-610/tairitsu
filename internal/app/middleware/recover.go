package middleware

import (
	"errors"

	"github.com/GT-610/tempest-zt/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

// errRecoveredPanic is what a recovered panic is reported as downstream.
//
// It is a fixed value rather than a wrap of the panic payload on purpose: the
// global error handler echoes a *fiber.Error's message to the caller, and a
// panicking handler can panic with anything, including a value that would read
// like a client-facing explanation. Nothing from the panic reaches the response.
var errRecoveredPanic = errors.New("recovered from a panic while serving the request")

// Recover keeps a panicking handler from taking the process with it.
//
// Fiber does not install a recoverer by default, so before this a single panic
// anywhere in a handler — a bad index on a decoded controller response, say —
// terminated the listener and dropped every connected user. Here the panic
// becomes an ordinary error, which the global error handler turns into its usual
// 500 body, and the panic value and stack are recorded in the log where they are
// useful and not reachable by the caller.
//
// It must be registered after ErrorHandler so that the error it returns travels
// back out through that handler.
func Recover() fiber.Handler {
	return recover.New(recover.Config{
		PanicHandler: func(c fiber.Ctx, recovered any) error {
			logger.Error("recovered from a panic while serving a request",
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Any("panic", recovered),
				zap.Stack("stack"))

			return errRecoveredPanic
		},
	})
}
