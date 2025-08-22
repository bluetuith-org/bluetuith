// Package fctx facilitates storing simple string based key-value data into
// contexts and then wrapping error values with that data so top-level error
// handlers have access to the data from the entire call chain.
//
// You can call `WithMeta` as many times as you like during a chain of function
// calls to decorate that call chain with metadata such as user IDs, request IDs
// and other business domain information. Then, when an error occurs, you wrap
// the error with a contextual error which contains the key-value data that was
// stored in the `context.Context` value. Then when your error is handled, you
// can easily extract this metadata for logging or error message purposes.
package fctx

import (
	"context"
	"errors"
)

type contextKey struct{}

// withContext implements the error interface and stores a simple table of data.
type withContext struct {
	underlying error
	meta       map[string]string
}

func (e *withContext) Error() string  { return "<fctx>" }
func (e *withContext) Cause() error   { return e.underlying }
func (e *withContext) Unwrap() error  { return e.underlying }
func (e *withContext) String() string { return e.Error() }

// WithMeta wraps a context with some arbitrary string based key-value metadata.
//
// You can do this at any point. If you're writing a HTTP server, the best place
// for this is middleware. You can store the user's ID or the path parameters or
// any information that's available. You can also continue to add metadata to a
// context at any point during a call chain as more metadata becomes available.
//
// Metadata is passed in as a simple argument list of key-value pairs. This must
// be an even number of arguments. It's easier to see if you format your code as
// a list of pairs, like this:
//
//	WithMeta(ctx,
//		"user_id", userID,
//		"post_id", postID,
//	)
func WithMeta(ctx context.Context, kv ...string) context.Context {
	if ctx == nil {
		return nil
	}

	// overwrite any existing context metadata
	return context.WithValue(ctx, contextKey{}, createMeta(ctx, kv...))
}

// Wrap wraps an error with the metadata stored in the context using `WithMeta`.
// You can also pass in additional key-value strings for some extra information.
//
// When errors occurs at a service boundary (such as a call to another package)
// you should wrap those errors with the available context value like this:
//
//	user, err := database.GetUser(ctx, userID)
//	if err != nil {
//		return nil, fctx.Wrap(err, ctx, "role", "admin")
//	}
//
// This library aims to be simple so there is no stack trace collection or
// additional message parameter. If you need this functionality, use pkg/errors.
//
//	user, err := database.GetUser(ctx, userID)
//	if err != nil {
//		return nil, fctx.Wrap(errors.Wrap(err, "failed to get user data"),
//			ctx,
//			"role", "admin")
//	}
func Wrap(err error, ctx context.Context, kv ...string) error {
	if err == nil || ctx == nil {
		return err
	}

	return &withContext{err, createMeta(ctx, kv...)}
}

func createMeta(ctx context.Context, kv ...string) map[string]string {
	meta := make(map[string]string)

	if parent, ok := ctx.Value(contextKey{}).(map[string]string); ok {
		// make a copy to avoid mutating parent context meta via map reference.
		for k, v := range parent {
			meta[k] = v
		}
	}

	l := len(kv)
	if l%2 != 0 {
		l -= 1 // don't error on odd number of args
	}

	for i := 0; i < l; i += 2 {
		k := kv[i]
		v := kv[i+1]

		meta[k] = v
	}

	return meta
}

// With implements the Fault Wrapper interface.
func With(ctx context.Context, kv ...string) func(error) error {
	return func(err error) error {
		return Wrap(err, ctx, kv...)
	}
}

// Unwrap pulls out any contextual metadata stored within an error as a simple
// string to string map. This data can then be used in your logger of choice, or
// be serialised to an RPC response of some kind. Below are some examples.
//
//	func HandleError(err error) {
//		metadata := fctx.Unwrap(err)
//		logger.Log("request error", metadata)
//	}
//
// If you use the Echo HTTP library, the error handler is a great use-case:
//
//	router.HTTPErrorHandler = func(err error, c echo.Context) {
//		ec := fctx.Unwrap(err)
//
//		l.Info("request error",
//		  zap.String("error", err.Error()),
//		  zap.Any("metadata", ec),
//		)
//
//		c.JSON(500, map[string]any{
//		  "error": err.Error(),
//		  "meta": ec,
//		})
//	}
func Unwrap(err error) map[string]string {
	values := map[string]string{}

	for err != nil {
		if f, ok := err.(*withContext); ok {
			if m := f.meta; m != nil {
				for k, v := range m {
					values[k] = v
				}
			}
		}

		err = errors.Unwrap(err)
	}

	if len(values) == 0 {
		return nil
	}

	return values
}

// GetMeta unwraps the stored metadata from a context value (if any) so you can
// use it in non-error situations such as structured logging. For example, Zap:
//
//	zap.L().Info("post created",
//		zap.String("key", "value"),
//		zap.Any("meta", fctx.GetMeta(ctx)),
//	)
//
// This will log your context metadata inside a `meta` keyed object. If you want
// your metadata on the top level of log entries, you can write a log specific
// glue layer to output a list of fields, for example, with Zap:
//
//	func Z(ctx context.Context) []zap.Field {
//		fields := []zap.Field{}
//		for k, v := range GetMeta(ctx) {
//			fields = append(fields, zap.String(k, v))
//		}
//		return fields
//	}
//
// Usage:
//
//	zap.L().Info("post created",
//		zap.String("key", "value"),
//		Z(ctx)...,
//	)
//
// Which will flatten out the KV metadata from the context into the log entry.
func GetMeta(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}

	meta, ok := ctx.Value(contextKey{}).(map[string]string)
	if !ok {
		return nil
	}

	return meta
}
