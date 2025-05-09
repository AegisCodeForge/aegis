package routes

type RouteErrorType int

const (
	NOT_FOUND RouteErrorType = 1
	OTHER_ERROR RouteErrorType = 2
)

type RouteError struct {
	ErrorType RouteErrorType
	ErrorMsg string
}

func (re RouteError) Error() string {
	return re.ErrorMsg
}

func IsRouteError(e error) bool {
	_, ok := e.(*RouteError)
	return ok
}

func NewRouteError(t RouteErrorType, msg string) *RouteError {
	return &RouteError{
		ErrorType: t,
		ErrorMsg: msg,
	}
}

