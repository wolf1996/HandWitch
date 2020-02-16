package bot

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type helpCommand struct {
	ctx     context.Context
	tg      telegram
	urlProc core.URLProcessor
	log     *log.Entry
}

func newHelpCommand(ctx context.Context, urlProc core.URLProcessor, tg telegram, log *log.Entry) comand {
	return &helpCommand{
		ctx:     ctx,
		urlProc: urlProc,
		tg:      tg,
		log:     log,
	}
}

func (proc *helpCommand) processCommonHelp() error {
	var respWriter strings.Builder
	err := proc.urlProc.WriteBriefHelp(&respWriter)
	if err != nil {
		return err
	}
	return proc.tg.Send(proc.ctx, respWriter.String())
}

func (proc *helpCommand) processArgs(messageArguments string) error {
	var respWriter strings.Builder

	name, err := getHandNameFromArguments(messageArguments)
	if err != nil {
		return fmt.Errorf("Failed to parse hand name from arguments %w", err)
	}

	handProc, err := proc.urlProc.GetHand(name)
	if err != nil {
		return fmt.Errorf("failed to get hand processor by name %s, %w", name, err)
	}

	err = handProc.WriteHelp(&respWriter)
	if err != nil {
		return err
	}
	return proc.tg.Send(proc.ctx, respWriter.String())
}

func (proc *helpCommand) Process(messageArguments string) error {
	if len(messageArguments) == 0 {
		return proc.processCommonHelp()
	}
	return proc.processArgs(messageArguments)
}

func newStartCommand(ctx context.Context, urlProc core.URLProcessor, tg telegram, log *log.Entry) comand {
	// no special actions on sart yet
	return &helpCommand{
		ctx:     ctx,
		urlProc: urlProc,
		tg:      tg,
		log:     log,
	}
}
