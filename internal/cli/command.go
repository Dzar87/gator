package cli

import (
	"context"
	"fmt"

	"github.com/Dzar87/gator/internal/state"
)

type Command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(context.Context, *state.State, Command) error
}

func (c *commands) Run(ctx context.Context, s *state.State, cmd Command) error {
	f, ok := c.handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return f(ctx, s, cmd)
}

func (c *commands) register(
	name string, f func(context.Context, *state.State, Command) error,
) {
	c.handlers[name] = f
}

func NewCommands() *commands {
	c := &commands{
		handlers: make(map[string]func(context.Context, *state.State, Command) error),
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	return c
}
