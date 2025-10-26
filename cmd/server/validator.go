package server

import (
	"errors"
	"fmt"
	"net/http"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func BindAndValidateIncomingRequestBody(ginCtx *gin.Context, requestBody any) bool {
	err := ginCtx.ShouldBindJSON(requestBody)
	if err != nil {
		// handle the errors captured by go validator
		var validationErrors validator.ValidationErrors

		if errors.As(err, &validationErrors) {
			errorMap := make(map[string]string)

			for _, fieldError := range validationErrors {
				// convert struct field name to snake_case to match the JSON field name
				jsonFieldName := toSnakeCase(fieldError.StructField())
				// generate a user-friendly error message for the JSON field name
				errorMap[jsonFieldName] = messageForTag(jsonFieldName, fieldError)
			}

			sendErrorResponse(ginCtx, http.StatusBadRequest, errorMap)
			return false
		}

		// any other errors that are not captured by go validator
		// eg: malformed JSON body
		sendErrorResponse(ginCtx, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func messageForTag(jsonFieldName string, fieldError validator.FieldError) string {
	fieldParam := fieldError.Param()
	fieldValue := fieldError.Value()

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is a required field", jsonFieldName)

	case "min":
		return fmt.Sprintf("%s must be at least %s characters", jsonFieldName, fieldParam)

	case "gt":
		return fmt.Sprintf("%s must be at greater than %s", jsonFieldName, fieldParam)

	case "email":
		return fmt.Sprintf("%v is not a valid email", fieldValue)
	}

	// fallback to default
	return fieldError.Error()
}

func toSnakeCase(str string) string {
	var result []rune
	runes := []rune(str)

	for i, r := range runes {
		if i > 0 && unicode.IsLower(runes[i-1]) && unicode.IsUpper(r) {
			result = append(result, '_')
		}

		if i > 0 && unicode.IsUpper(runes[i-1]) && unicode.IsUpper(r) &&
			(i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
			result = append(result, '_')
		}

		result = append(result, unicode.ToLower(r))
	}

	return string(result)
}
