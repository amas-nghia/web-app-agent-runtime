package adapters

import (
	"fmt"

	"web-app-agent-runtime/internal/adapters/mock"
	"web-app-agent-runtime/internal/adapters/opencode"
)

func New(name string) (Adapter, error) {
	switch name {
	case "", "mock":
		return mock.New(), nil
	case "opencode":
		return opencode.New(), nil
	default:
		return nil, fmt.Errorf("adapters: unknown engine %q", name)
	}
}
