package main

import "net/http"

// logError is generic helper for logging error message.
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// errorResponse is generic helper for sending JSON-formatted error message
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{
		"error": message,
	}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorResponse will be used to send a 500 Internal Server Error status code with JSON formatted
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	app.errorResponse(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

// notFoundResponse will be used to send a 404 Not Found status code with JSON formatted
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

// methodNotAllowedResponse will be used to send a 405 Method Not Allowed status code with JSON formatted
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}

// badRequestResponse will be used to send a 400 Bad Request status code with JSON formatted
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse will be used to send a 422 Unprocessable Entity status code with JSON formatted
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// editConflictResponse will be used to send a 409 Conflict status code with JSON formatted
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusConflict, http.StatusText(http.StatusConflict))
}

// rateLimitExceededResponse will be used to send a 429 Too Many Requests status code with JSON formatted
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
}

// invalidCredentialsResponse will be used to send a 401 Unauthorized status code with JSON formatted.
// This error helper is used when credential that provided by client is invalid
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// invalidCredentialsResponse will be used to send a 401 Unauthorized status code with JSON formatted.
// This error helper is used when authentication token that provided by client is invalid.
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	// Including a "WWW-Authenticate: Bearer" header here to help inform or remind the client
	// that we expect to authenticate using a bearer token.
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authnetication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// authenticationRequiredResponse will be used to send a 401 Unauthorized status code with JSON formatted.
// This error helper is used when the client is not authenticated yet
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// inactiveAccountResponse will be used to send a 403 Forbidden status code with JSON formatted.
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

// inactiveAccountResponse will be used to send a 403 Forbidden status code with JSON formatted.
func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
