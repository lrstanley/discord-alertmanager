// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.
package main

import (
	"context"

	"github.com/apex/log"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lrstanley/clix"
	"github.com/lrstanley/discord-alertmanager/internal/alertmanager"
	"github.com/lrstanley/discord-alertmanager/internal/bot"
	"github.com/lrstanley/discord-alertmanager/internal/models"
)

var (
	version = "master"
	commit  = "latest"
	date    = "-"

	logger log.Interface
	cli    = &clix.CLI[models.Flags]{
		Links: clix.GithubLinks("github.com/lrstanley/discord-alertmanager", "master", "https://liam.sh"),
		VersionInfo: &clix.VersionInfo[models.Flags]{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
)

func main() {
	cli.LoggerConfig.Pretty = true
	cli.Parse()
	logger = cli.Logger

	ctx := log.NewContext(context.Background(), logger)

	al, err := alertmanager.NewClient(cli.Flags.Alertmanager, cli.Debug)
	if err != nil {
		logger.WithError(err).Fatal("error creating alertmanager client")
	}

	b, err := bot.New(ctx, cli.Flags.Discord, al, cli.Debug)
	if err != nil {
		logger.WithError(err).Fatal("error creating bot")
	}

	if err := clix.RunCtx(ctx, b.Run); err != nil {
		logger.WithError(err).Fatal("error running bot")
	}
}
