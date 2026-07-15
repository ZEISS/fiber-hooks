package hooks

import (
	"github.com/gofiber/fiber/v3"
)

// NoopEvent is a no-op event that is used when no event is available.
var NoopEvent = Event{}

type noopDecoder struct{}

// Decode implements the Decoder interface for the noopDecoder.
func (d *noopDecoder) Decode(c fiber.Ctx, secret string) (Event, error) {
	return NoopEvent, nil
}

type noopDispatcher struct{}

var _ Dispatcher = (*noopDispatcher)(nil)

// Dispatch implements the Dispatcher interface for the noopDispatcher.
func (d *noopDispatcher) Dispatch(event Event) error {
	// no-op
	return nil
}
