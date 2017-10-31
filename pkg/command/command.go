package command

import (
	"context"
	"fmt"
	"time"
	"os/exec"
)

const COMMAND_TIMEOUT time.Duration = 10

type Handler interface {
	RunCommand(string, []string)
}

type handler struct {
	command *exec.Cmd
	ctx context.Context
	cancel context.CancelFunc
}

func(h *handler) RunCommand(cmd string, args []string) {
	h.ctx, h.cancel = context.WithTimeout(context.Background(), COMMAND_TIMEOUT * time.Second)
	defer h.cancel()
	h.command = exec.CommandContext(h.ctx, cmd, args...)

	out, err := h.command.Output()
	if h.ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Command timed out")
	}

	fmt.Println("Output:", string(out))
	if err != nil {
		fmt.Println("Non-zero exit code:", err)
	}
}

func NewHandler() Handler {
	return &handler{}
}