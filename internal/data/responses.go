package data

// Struct to return error messages
type ErrorResponseMessage struct {
	ErrType string `json:"type"`    // The type of error
	Message string `json:"message"` // The error message in details
}

// Struct to return helpful messages
type SuccessResponseMessage struct {
	Title       string `json:"title"`       // Title of a message
	Description string `json:"description"` // Detailed description of success message
}

// A slice of SuccessResponseMessage
type MessageBox []SuccessResponseMessage

// A slice of ErrorResponseMessage
type ErrorBox []ErrorResponseMessage

// Add a new message to successlist
func (m *MessageBox) Add(obj SuccessResponseMessage) {
	*m = append(*m, obj)
}

// A message response
func MessageResponse(title, description string) SuccessResponseMessage {
	return SuccessResponseMessage{
		Title:       title,
		Description: description,
	}
}

// Add a new error into errobox
func (e *ErrorBox) Add(obj ErrorResponseMessage) {
	*e = append(*e, obj)
}

// Return an response for internal Server Error
func InternalServerErrorResponse(msg string) ErrorResponseMessage {
	return ErrorResponseMessage{
		ErrType: "Internal Server Error",
		Message: msg}

}

// Response for 404 Resource not found
func ResourceNotFoundResponse(msg string) ErrorResponseMessage {

	return ErrorResponseMessage{
		ErrType: "Resource Not Found",
		Message: msg}

}

// Response for 400 Bad request
func BadRequestResponse(msg string) ErrorResponseMessage {
	return ErrorResponseMessage{
		ErrType: "Bad Request",
		Message: msg}
}

// A custom error response
func CustomErrorResponse(errorType string, msg string) ErrorResponseMessage {
	return ErrorResponseMessage{
		ErrType: errorType,
		Message: msg,
	}
}

// Response for Invalid credentials
func InvalidCredentialsResponse(msg string) ErrorResponseMessage {
	return ErrorResponseMessage{
		ErrType: "Authentication Error",
		Message: msg,
	}
}

// Response for Authorization Error
func AuthorizationErrorResponse(msg string) ErrorResponseMessage {
	return ErrorResponseMessage{
		ErrType: "Authorization Error",
		Message: msg,
	}
}

// Response for Account Error
func AccountErrorResponse(msg string) ErrorResponseMessage {
	return ErrorResponseMessage{
		ErrType: "Account Error",
		Message: msg,
	}
}
