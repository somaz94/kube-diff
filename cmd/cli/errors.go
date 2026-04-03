package cli

import "errors"

// ErrChangesDetected is returned when diff results contain changes.
// Callers should handle this by exiting with code 1.
var ErrChangesDetected = errors.New("changes detected")
