package bot

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type helpCommand struct {
	ctx      context.Context
	tg       telegram
	handProc core.HandProcessor
	log      *log.Entry
}

func newHelpCommand(ctx context.Context, handProc core.HandProcessor, tg telegram, log *log.Entry) comand {
	return &helpCommand{
		ctx:      ctx,
		handProc: handProc,
		tg:       tg,
		log:      log,
	}
}

func (proc *helpCommand) Process(messageArguments string) error {
	var respWriter strings.Builder
	err := proc.handProc.WriteHelp(&respWriter)
	if err != nil {
		return err
	}
	return proc.tg.Send(proc.ctx, respWriter.String())
}
