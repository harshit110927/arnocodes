package course

import "errors"

var ErrTopicLocked = errors.New("topic locked")
var ErrDiagnosticNotCompleted = errors.New("diagnostic not completed")
