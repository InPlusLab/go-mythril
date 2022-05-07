package ethereum

import (
	"sync"
	"time"
)

type TimeHandler struct {
	StartTime     int
	ExecutionTime int
}

var timeHandler *TimeHandler
var once sync.Once

// Singleton
func GetTimeHandlerInstance() *TimeHandler {
	once.Do(func() {
		timeHandler = &TimeHandler{
			StartTime:     0,
			ExecutionTime: 0,
		}
	})
	return timeHandler
}

func (handler *TimeHandler) StartExecution(executionTime int) {
	handler.StartTime = int(time.Now().Unix() * 1000)
	handler.ExecutionTime = executionTime * 1000
}

func (handler *TimeHandler) TimeRemaining() int {
	return handler.ExecutionTime - (int(time.Now().Unix()*1000) - handler.StartTime)
}
