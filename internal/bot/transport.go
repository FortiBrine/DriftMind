package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/FortiBrine/DriftMind/internal/config"
	"github.com/mymmrac/telego"
)

type transport struct {
	logger *slog.Logger
	server *http.Server
}

func newTransport(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	bot *telego.Bot,
) (t *transport, updates <-chan telego.Update, err error) {
	if cfg.WebhookURL == "" {
		if err = bot.DeleteWebhook(ctx, new(telego.DeleteWebhookParams{
			DropPendingUpdates: true,
		})); err != nil {
			return nil, nil, fmt.Errorf("deleting webhook: %w", err)
		}

		if updates, err = bot.UpdatesViaLongPolling(ctx, nil); err != nil {
			return nil, nil, fmt.Errorf("setting up long polling: %w", err)
		}

		return new(transport{logger: logger}), updates, nil
	}

	secretToken := bot.SecretToken()

	if err = bot.SetWebhook(ctx, new(telego.SetWebhookParams{
		URL:         cfg.WebhookURL,
		SecretToken: secretToken,
	})); err != nil {
		return nil, nil, fmt.Errorf("setting webhook: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	server := new(http.Server{
		Addr:    cfg.HttpAddress,
		Handler: mux,
	})

	if updates, err = bot.UpdatesViaWebhook(ctx, telego.WebhookHTTPServer(server, "/webhook", secretToken)); err != nil {
		return nil, nil, fmt.Errorf("setting up webhook: %w", err)
	}

	return new(transport{logger: logger, server: server}), updates, nil
}

func (t *transport) Start() {
	if t.server == nil {
		return
	}

	go func() {
		if err := t.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.logger.Error("error starting HTTP server", "error", err)
		}
	}()
}

func (t *transport) Stop(ctx context.Context) error {
	if t.server == nil {
		return nil
	}
	if err := t.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down HTTP server: %w", err)
	}
	return nil
}
