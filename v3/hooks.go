// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package hooks

import (
	"github.com/gofiber/fiber/v3"
)

// Event is a struct that holds the event data.
type Event struct {
	// EventType is the type of the event.
	EventType string `json:"event_type"`
	// DeliveryID is the ID of the delivery.
	DeliveryID string `json:"delivery_id"`
	// Payload is the payload of the event.
	Payload []byte `json:"payload"`
}

// Decoder is the function to decode the event payload.
type Decoder func(c fiber.Ctx, secret string) (Event, error)

// Dispatcher is the function to dispatch the event to the registered handlers.
type Dispatcher func() chan<- Event

// New creates a new handler to manage the session.
func New(config ...Config) fiber.Handler {
	cfg := configDefault(config...)

	return func(c fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		chn := cfg.Dispatcher()
		defer close(chn)

		event, err := cfg.Decoder(c, cfg.SigningSecret)
		if err != nil {
			return err
		}

		chn <- event

		return c.Next()
	}
}

// Config caputes the configuration for running the goth middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c fiber.Ctx) bool
	// SigningSecret is the secret used to sign the session cookie.
	SigningSecret string
	// Decoder is the function to decode the event payload.
	Decoder Decoder
	// Dispatcher is the function to dispatch the event to the registered handlers.
	Dispatcher Dispatcher
}

// ConfigDefault is the default config.
var ConfigDefault = Config{
	Decoder:    DefaultDecoder,
	Dispatcher: DefaultDispatcher,
}

func configDefault(config ...Config) Config {
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}

	if cfg.Decoder == nil {
		cfg.Decoder = ConfigDefault.Decoder
	}

	if cfg.Dispatcher == nil {
		cfg.Dispatcher = ConfigDefault.Dispatcher
	}

	if cfg.SigningSecret == "" {
		cfg.SigningSecret = ConfigDefault.SigningSecret
	}

	return cfg
}
