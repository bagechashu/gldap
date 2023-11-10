package logging

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/fatih/color"
)

func InitSlogDefault(debug bool) {
	var slogOption = &slog.HandlerOptions{}
	var slogHandler *slog.Logger
	if debug {
		slogOption.AddSource = true
		slogOption.Level = slog.LevelDebug
		slogHandler = slog.New(NewPrettyTextHandler(os.Stdout, slogOption))
	} else {
		slogOption.Level = slog.LevelInfo
		slogOption.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}
			}
			return a
		}
		slogHandler = slog.New(slog.NewJSONHandler(os.Stdout, slogOption))
	}

	slog.SetDefault(slogHandler)
}

type PrettyHandler struct {
	slog.Handler
	l *log.Logger
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(timeStr, level, msg, color.WhiteString(string(b)))

	return nil
}

func NewPrettyJsonHandler(
	out io.Writer,
	opts slog.HandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, &opts),
		l:       log.New(out, "", 0),
	}

	return h
}

func NewPrettyTextHandler(
	out io.Writer,
	opts *slog.HandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewTextHandler(out, opts),
		l:       log.New(out, "", 0),
	}

	return h
}
