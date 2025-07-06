package ratelimiter

import "errors"

// Errors used throughout the codebase
var (
	ErrInvalidLimit            = errors.New("limit must be greater than zero")
	ErrInvalidInterval         = errors.New("interval must be greater than zero")
	ErrInvalidThrottleDuration = errors.New("throttle duration must be greater than zero")
	ErrTokenFactoryNotDefined  = errors.New("token factory must be defined")
)
