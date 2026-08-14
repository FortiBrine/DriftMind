package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/FortiBrine/DriftMind/internal/config"
	"github.com/FortiBrine/DriftMind/internal/middleware"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type Bot struct {
	logger    *slog.Logger
	bh        *th.BotHandler
	transport *transport
}

func New(
	ctx context.Context, cfg config.Config,
	logger *slog.Logger,
) (b *Bot, err error) {
	telegoBot, err := telego.NewBot(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("creating bot: %w", err)
	}

	tr, updates, err := newTransport(ctx, cfg, logger, telegoBot)
	if err != nil {
		return nil, fmt.Errorf("setting up transport: %w", err)
	}
	defer func() {
		if err != nil {
			tr.Stop(ctx)
		}
	}()

	bh, err := th.NewBotHandler(telegoBot, updates, th.WithErrorHandler(
		middleware.NewCustomErrorHandler(logger),
	))
	if err != nil {
		return nil, fmt.Errorf("creating bot handler: %w", err)
	}
	defer func() {
		if err != nil {
			bh.Stop()
		}
	}()
	bh.Use(th.PanicRecoveryHandler(func(recovered any) error {
		return fmt.Errorf("panic in handler: %v", recovered)
	}))
	RegisterRoutes(
		bh,
		logger,
	)

	b = new(Bot{
		logger:    logger,
		bh:        bh,
		transport: tr,
	})

	return
}

func (b *Bot) Start(ctx context.Context) error {
	b.transport.Start()

	done := make(chan error, 1)
	go func() {
		done <- b.bh.Start()
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-done:
		if err != nil {
			return fmt.Errorf("starting bot handler: %w", err)
		}
		return nil
	}
}

func (b *Bot) Close(ctx context.Context) error {
	if err := b.transport.Stop(ctx); err != nil {
		return fmt.Errorf("stopping transport: %w", err)
	}

	if err := b.bh.StopWithContext(ctx); err != nil {
		return fmt.Errorf("stopping bot handler: %w", err)
	}

	return nil
}
