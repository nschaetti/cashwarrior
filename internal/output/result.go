package output

// Error describes a command failure in a machine-readable result.
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Result is the format-independent result returned by a command.
// Data is intentionally left as a DTO supplied by the command layer.
type Result struct {
	Success bool   `json:"success"`
	Type    string `json:"type,omitempty"`
	Data    any    `json:"data"`
	Count   int    `json:"count"`
	Error   *Error `json:"error,omitempty"`
}

// SuccessResult creates a successful command result.
func SuccessResult(resultType string, data any, count int) Result {
	return Result{Success: true, Type: resultType, Data: data, Count: count}
}

// FailureResult creates a failed command result.
func FailureResult(resultType string, commandError Error) Result {
	return Result{Success: false, Type: resultType, Error: &commandError}
}
