package github

import (
	"github.com/gofiber/fiber/v3"
	hooks "github.com/zeiss/fiber-hooks/v3"
)

type decoderImpl struct{}

var _ hooks.Decoder = (*decoderImpl)(nil)

// NewDecoder returns a new decoder implementation for the GitHub API.
func NewDecoder() hooks.Decoder {
	return &decoderImpl{}
}

// Decoder verifies the payload from the request using the GitHub API.
func (d *decoderImpl) Decode(c fiber.Ctx, secret string) (hooks.Event, error) {
	event := hooks.Event{}

	b, err := ValidatePayload(c, []byte(secret))
	if err != nil {
		return event, err
	}

	eventType := c.Get("X-GitHub-Event")
	deliveryID := c.Get("X-GitHub-Delivery")

	event.EventType = eventType
	event.DeliveryID = deliveryID
	event.Payload = b

	return event, nil
}
