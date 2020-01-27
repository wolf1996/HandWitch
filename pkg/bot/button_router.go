package bot

import (
	"fmt"
)

type (
	buttonHandler = func(msg string) (processingState, error)

	buttonRouters = []buttonHandler
)

func applyMessageRouters(msg string, routers buttonRouters) (processingState, error) {
	for _, handler := range routers {
		state, err := handler(msg)
		if err == nil {
			return state, err
		}
	}
	return nil, fmt.Errorf("can't apply any handler")
}
