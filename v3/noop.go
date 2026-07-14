package hooks

import (
	"github.com/gofiber/fiber/v3"
)

// DefaultDecoder is a no-op decoder that always returns an empty event and no error.
func DefaultDecoder(c fiber.Ctx, secret string) (Event, error) {
	event := Event{}
	return event, nil
}

// DefaultDispatcher is a no-op dispatcher that does nothing.
func DefaultDispatcher() chan<- Event {
	ch := make(chan Event)

	go func() {
		for range ch {
		}
	}()

	return ch
}
