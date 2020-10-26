package runner

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cskr/pubsub"
	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const amx = "gamemodes/ScavengeSurvive.amx"

func RunAPI(ctx context.Context, ps *pubsub.PubSub, restartTime time.Duration) {
	rtr := chi.NewRouter()

	update := atomic.NewBool(false)

	rtr.Get("/update", func(w http.ResponseWriter, r *http.Request) {
		if update.Load() {
			zap.L().Info("sending restart signal", zap.Duration("time", restartTime))
			w.Write([]byte(fmt.Sprintf("update %d", int(restartTime.Seconds())))) //nolint:errcheck
			ps.Pub(restartTime, "server_update")
		}

		update.Store(false)
	})

	go http.ListenAndServe(":7788", rtr) //nolint:errcheck

	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = w.Add(amx)
	if err != nil {
		panic(err)
	}
	for range w.Events {
		info, err := os.Stat(amx)
		if err != nil {
			continue
		}
		if info.Size() > 0 {
			update.Store(true)
		}
	}
}
