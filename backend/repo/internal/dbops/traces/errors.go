package traces

import "errors"

// ErrEventNotFound is returned when no event exists for the given event_id.
var ErrEventNotFound = errors.New("dbops: event not found")
