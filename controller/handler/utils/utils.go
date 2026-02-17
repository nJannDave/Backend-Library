package utils

import (
	"stmnplibrary/dto"
	"strings"

	"github.com/go-playground/validator/v10"
)

func getMsgType(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "fields must be filled in"
	case "email":
		return "incorrect email format"
	case "number":
		return "must be a number"
	case "min":
		return "exceeds the minimum limit: " + err.Param()
	case "max":
		return "exceeds the maximum limit: " + err.Param()
	case "gt":
		return "len category id must > 0"
	case "unique":
		return "id category must be unique"
	default:
		return "unknown validation"
	}
}

func GetData(fn func() error, message string) *dto.Response {
	const status = "false / error"
	const jserr = "json format is wrong"
	if err := fn(); err != nil {
		errMsg := &dto.Response{
			Status:  status,
			Message: message,
		}
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, i := range ve {
				errMsg.Error.Binding = append(errMsg.Error.Binding, dto.Binding{
					Field:  i.Field(),
					Reason: getMsgType(i),
				})
			}
			return errMsg
		}
		errMsg.Error.Error = jserr
		return errMsg
	}
	return nil
}

func ErrorMsg(resMsg string, msg []string) *dto.Response {
	const status = "false / failed"
	errBS := &dto.Response{
		Status:  status,
		Message: resMsg,
	}
	for _, v := range msg {
		errBS.Error.Service = append(errBS.Error.Service, dto.Service{
			Reason: v,
		})
	}
	return errBS
}

func ValidateErr(err error, resMsg string, errMsg string) (int, *dto.Response) {
	const status = "false / failed"
	if strings.Contains(err.Error(), "internal server error: ") || strings.Contains(err.Error(), "operation" ) {
		return 500, &dto.Response{
			Status:  status,
			Message: resMsg,
			Error:   dto.Errors{Error: "an error occurred"},
		}
	}
	if strings.Contains(err.Error(), "no data found") {
		return 404, &dto.Response{
			Status:  status,
			Message: resMsg,
			Error:   dto.Errors{Error: err.Error()},
		}
	}
	if strings.Contains(err.Error(), "no data affected") {
		return 400, &dto.Response{
			Status:  status,
			Message: resMsg,
			Error:   dto.Errors{Error: errMsg},
		}
	}
	if strings.Contains(err.Error(), "duplicate request") {
		return 409, &dto.Response{
			Status: status,
			Message: resMsg,
			Error: dto.Errors{Error: err.Error()},
		}
	}
	return 400, &dto.Response{
		Status:  status,
		Message: resMsg,
		Error:   dto.Errors{Error: err.Error()},
	}
}
