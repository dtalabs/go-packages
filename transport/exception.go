package transport

import (
	"errors"
	"net/http"
	"sync"
)

type Exception struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e Exception) Error() string {
	return e.Message
}

var (
	NotFoundException   = Exception{Message: "Not Found", Status: http.StatusNotFound}
	BadRequestException = Exception{Message: "Bad Request", Status: http.StatusBadRequest}
	InternalServerError = Exception{Message: "Internal Server Error", Status: http.StatusInternalServerError}
	UnprocessableEntity = Exception{Message: "Unprocessable Entity", Status: http.StatusUnprocessableEntity}
	Unauthorized        = Exception{Message: "Unauthorized", Status: http.StatusUnauthorized}
	TooManyRequests     = Exception{Message: "Too many attempts; try again later", Status: http.StatusTooManyRequests}
)

var mappings = make(map[error]Exception)
var mu = sync.RWMutex{}

func Register(err error, ex Exception) {
	mu.Lock()
	defer mu.Unlock()
	mappings[err] = ex
}

func ToException(err error) Exception {
	var ex Exception
	if errors.As(err, &ex) {
		return ex
	}
	mu.RLock()
	defer mu.RUnlock()
	if exception, ok := mappings[err]; ok {
		return exception
	}
	for target, exception := range mappings {
		if errors.Is(err, target) {
			return exception
		}
	}
	return InternalServerError
}
