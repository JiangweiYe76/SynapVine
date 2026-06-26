package repository

import "errors"

// ErrNotFound is returned by repository methods when the requested row does
// not exist. Callers translate this into a 404.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a uniqueness constraint is violated
// (e.g. provider name already taken).
var ErrDuplicate = errors.New("duplicate")
