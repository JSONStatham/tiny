package response

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

type ErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func Error(msg string, errs ...map[string]string) ErrorResponse {
	if len(errs) == 0 {
		return ErrorResponse{
			Message: msg,
		}
	}

	return ErrorResponse{
		Message: msg,
		Errors:  errs[0],
	}
}
