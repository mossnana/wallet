package responses

type ErrorResponse struct {
	StatusCode int      `json:"status_code"`
	Reasons    []string `json:"reasons"`
}

func NewErrorResponse(statusCode int, reasons ...string) ErrorResponse {
	return ErrorResponse{
		StatusCode: statusCode,
		Reasons:    reasons,
	}
}
