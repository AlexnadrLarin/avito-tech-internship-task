package handlers

type Validator interface {
	Validate(s interface{}) error
}
