package ratelimiter

import "errors"

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported rate limiting algorithm")
)
