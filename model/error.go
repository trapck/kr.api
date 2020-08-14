package model

//GenericError is model of an generic error
type GenericError struct {
	Code    int    `json:"code"`
	Debug   string `json:"debug"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
	Request string `json:"request"`
	Status  string `json:"status"`
}

//GenericErrorWrap is model for generic erorr http wrap
type GenericErrorWrap struct {
	Error GenericError
}

//NewGenericErrorWrap returns simplest error object configuration based on error
func NewGenericErrorWrap(code int, e error) GenericErrorWrap {
	return GenericErrorWrap{GenericError{Code: code, Message: e.Error()}}
}
