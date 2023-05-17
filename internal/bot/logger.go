// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"fmt"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/apex/log"
)

var _ disgord.Logger = (*discordLogger)(nil)

// discordLogger is a wrapper for apex/log, that can be used with the disgord.Logger
// interface.
type discordLogger struct {
	logger log.Interface
}

func (l *discordLogger) wrap(v ...any) string {
	var out []string
	for _, i := range v {
		out = append(out, fmt.Sprint(i))
	}
	return strings.Join(out, " ")
}

func (l *discordLogger) Debug(v ...any) {
	l.logger.Debug(l.wrap(v...))
}

func (l *discordLogger) Info(v ...any) {
	l.logger.Info(l.wrap(v...))
}

func (l *discordLogger) Error(v ...any) {
	l.logger.Error(l.wrap(v...))
}
