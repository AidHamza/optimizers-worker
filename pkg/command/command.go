package command

import (
	"bytes"
	"context"
	"time"
	"strings"
	"os/exec"

	"github.com/AidHamza/optimizers-worker/pkg/log"
)

const COMMAND_TIMEOUT time.Duration = 10

type Handler interface {
	RunCommand(string, []string, []byte) ([]byte, error)
	EnableDebug(bool)
}

type handler struct {
	command *exec.Cmd
	ctx context.Context
	cancel context.CancelFunc
	debug bool
}

func(h *handler) RunCommand(cmd string, args []string, imgBuf []byte) ([]byte, error) {
	var cmdOut bytes.Buffer

	h.ctx, h.cancel = context.WithTimeout(context.Background(), COMMAND_TIMEOUT * time.Second)
	defer h.cancel()
	h.command = exec.CommandContext(h.ctx, cmd, args...)
	h.command.Stdin = strings.NewReader(string(imgBuf))
	h.command.Stdout = &cmdOut

	if h.debug == true {
		log.Logger.Error("Executing", "CMD", strings.Join(h.command.Args, " "))
	}

	if err := h.command.Run(); err != nil {
        log.Logger.Error("Cannot start optimizers command, Non-zero exit code", "CMD_START_FAILED", err.Error(), "CMD", cmd, "ARGS", args)
        return []byte{0}, err
    }

	if h.ctx.Err() == context.DeadlineExceeded {
		log.Logger.Error("Command Timeout reached", "CMD_TIMEOUT_REACHED", h.ctx.Err(), "CMD", cmd, "ARGS", args)
		return []byte{0}, h.ctx.Err()
	}

	return cmdOut.Bytes(), nil
}

func(h *handler) EnableDebug(isDebug bool) {
	h.debug = isDebug
}

func NewHandler() Handler {
	return &handler{}
}