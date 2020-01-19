package bot

import (
	"context"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type helpCommand struct {
	ctx      context.Context
	tg       telegram
	handProc core.HandProcessor
	log      *log.Entry
}

func NewHelpCommand(ctx context.Context, handProc core.HandProcessor, tg telegram, log *log.Entry) comand {
	return &helpCommand{
		ctx:      ctx,
		handProc: handProc,
		tg:       tg,
		log:      log,
	}
}

func (proc *helpCommand) Process(messageArguments string, writer io.Writer) error {
	return proc.handProc.WriteHelp(writer)
}
