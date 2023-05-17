// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package models

type Flags struct {
	Discord      ConfigDiscord      `group:"Discord Options" namespace:"discord" env-namespace:"DISCORD"`
	Alertmanager ConfigAlertmanager `group:"Alertmanager Options" namespace:"alertmanager" env-namespace:"ALERTMANAGER"`
}

type ConfigDiscord struct {
	Token string `long:"token" env:"TOKEN" required:"true" description:"Discord bot token"`
}

type ConfigAlertmanager struct {
	URL      string `long:"url" env:"URL" required:"true" description:"Alertmanager URL"`
	Username string `long:"username" env:"USERNAME" description:"Alertmanager username (if configured)"`
	Password string `long:"password" env:"PASSWORD" description:"Alertmanager password (if configured)"`
}
