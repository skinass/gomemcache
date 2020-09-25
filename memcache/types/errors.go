package types

import "errors"

var (
	// ErrCacheMiss means that a Get failed because the item wasn't present.
	ErrCacheMiss = errors.New("memcache: cache miss")

	// ErrCASConflict means that a CompareAndSwap call failed due to the
	// cached value being modified between the Get and the CompareAndSwap.
	// If the cached value was simply evicted rather than replaced,
	// ErrNotStored will be returned instead.
	ErrCASConflict = errors.New("memcache: compare-and-swap conflict")

	// ErrNotStored means that a conditional write operation (i.e. Add or
	// CompareAndSwap) failed because the condition was not satisfied.
	ErrNotStored = errors.New("memcache: item not stored")

	// ErrServer means that a server error occurred.
	ErrServerError = errors.New("memcache: server error")

	// ErrNoStats means that no statistics were available.
	ErrNoStats = errors.New("memcache: no statistics available")

	// ErrMalformedKey is returned when an invalid key is used.
	// Keys must be at maximum 250 bytes long and not
	// contain whitespace or control characters.
	ErrMalformedKey = errors.New("malformed: key is too long or contains invalid characters")

	// ErrNoServers is returned when no servers are configured or available.
	ErrNoServers = errors.New("memcache: no servers configured or available")

	ErrValueTooLarge  = errors.New("memcache: value to large")
	ErrInvalidArgs    = errors.New("memcache: invalid arguments")
	ErrValueNotStored = errors.New("memcache: value not stored")
	ErrNonNumeric     = errors.New("memcache: incr/decr called on non-numeric value")
	ErrAuthRequired   = errors.New("memcache: authentication required")
	ErrAuthContinue   = errors.New("memcache: authentication continue (unsupported)")
	ErrUnknownCommand = errors.New("memcache: unknown command")
	ErrOutOfMemory    = errors.New("memcache: out of memory")
	ErrUnknownError   = errors.New("memcache: unknown error from server")
)
