// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/apex/log"
	"github.com/go-openapi/runtime"
	"github.com/kr/pretty"
	"github.com/lrstanley/discord-alertmanager/internal/alertmanager"
	"github.com/lrstanley/discord-alertmanager/internal/models"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
)

const httpRequestTimeout = 5 * time.Second

type Bot struct {
	ctx    context.Context
	config models.ConfigDiscord
	logger log.Interface
	debug  bool

	client *disgord.Client
	self   *disgord.User

	al *alertmanager.Client
}

// New creates a new bot instance. It will make a few calls to Discord to validate
// the bot config. Make sure to call Run() to start the bot.
func New(ctx context.Context, config models.ConfigDiscord, al *alertmanager.Client, debug bool) (b *Bot, err error) {
	b = &Bot{
		ctx:    ctx,
		config: config,
		logger: log.FromContext(ctx).WithField("src", "bot"),
		debug:  debug,
		al:     al,
	}

	b.client, err = disgord.NewClient(ctx, disgord.Config{
		ProjectName: "discord-alertmanager (https://github.com/lrstanley/discord-alertmanager, https://liam.sh)",
		BotToken:    b.config.Token,
		Logger:      &discordLogger{logger: b.logger},
		Presence: &disgord.UpdateStatusPayload{
			Since: nil,
			Game: []*disgord.Activity{
				{Name: "👀", Type: disgord.ActivityTypeGame},
			},
			Status: disgord.StatusOnline,
			AFK:    false,
		},
	})
	if err != nil {
		return nil, err
	}

	b.client.Gateway().BotReady(b.onReady)
	b.client.Gateway().InteractionCreate(b.onInteractionCreate)

	return b, nil
}

// Run starts the bot. It will block until the context is canceled, in which it will
// then gracefully shutdown, disconnecting from Discord.
func (b *Bot) Run(ctx context.Context) error {
	authorizationURL, err := b.client.BotAuthorizeURL(disgord.PermissionUseSlashCommands, []string{
		"bot",
		"applications.commands",
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to get bot auth url")
		return err
	}
	b.logger.WithField("url", authorizationURL).Info("please visit the following url to add the bot to your server")

	b.self, err = b.client.CurrentUser().Get()
	if err != nil {
		b.logger.WithError(err).Error("failed to get bot user")
		return err
	}

	err = b.client.Gateway().Connect()
	if err != nil {
		b.logger.WithError(err).Error("failed to connect to discord")
		return err
	}

	<-ctx.Done()
	b.logger.Info("shutting down")
	_ = b.client.Gateway().Disconnect()
	time.Sleep(500 * time.Millisecond) //nolint:gomnd

	return nil
}

// onReady is called when the bot is ready to start receiving events.
func (b *Bot) onReady() {
	b.logger.Info("updating application commands")
	if err := b.client.ApplicationCommand(b.self.ID).Global().BulkOverwrite(commands); err != nil {
		b.logger.WithError(err).Fatal("failed to update application commands")
	}
}

// onInteractionCreate is called when a user interacts with the bots slash commands.
func (b *Bot) onInteractionCreate(s disgord.Session, h *disgord.InteractionCreate) {
	fmt.Printf("% #v\n", pretty.Formatter(*h)) // TODO: replace with proper logging?

	// Static custom IDs.
	switch h.Data.CustomID {
	case "modal-add":
		b.silenceAddFromModal(s, h)
		return
	}

	// Custom IDs that have a specific prefix, and provided arguments.
	if i := strings.Index(h.Data.CustomID, "/"); i != -1 {
		id := h.Data.CustomID[:i]
		args := strings.Split(h.Data.CustomID[i+1:], "/")

		switch id {
		case "silence-remove":
			b.silenceRemoveFromCallback(s, h, args)
			return
		}
	}

	switch h.Data.Name {
	case "silence alert": // Message commands.
		b.silenceAddFromMessage(s, h)
	case "silences": // Application commands.
		switch h.Data.Options[0].Name {
		case "add":
			b.silenceAddFromCommand(s, h)
			return
		case "get":
			b.silenceGetFromCommand(s, h)
			return
		case "list":
			b.silenceListFromCommand(s, h)
			return
		case "remove":
			b.silenceRemoveFromCommand(s, h)
			return
		}
	}

	b.logger.WithFields(log.Fields{
		"guild_id":   h.GuildID,
		"channel_id": h.ChannelID,
		"user_id":    h.Member.User.ID,
		"message_id": h.ID,
		"command":    h.Data.Name,
		"custom_id":  h.Data.CustomID,
	}).Warn("unknown interaction")
}

func (b *Bot) responseError(s disgord.Session, h *disgord.InteractionCreate, title string, originalErr error) {
	b.logger.WithFields(log.Fields{
		"guild_id":   h.GuildID,
		"channel_id": h.ChannelID,
		"user_id":    h.Member.User.ID,
		"message_id": h.ID,
	}).WithError(originalErr).Error(title)

	//nolint:errorlint,revive,stylecheck
	switch terr := originalErr.(type) {
	case *silence.GetSilenceNotFound:
		originalErr = errors.New("Alertmanager was unable to find the requested silence. Please check your inputs and try again.")
	case *runtime.APIError:
		switch terr.Code {
		case http.StatusBadRequest:
			originalErr = fmt.Errorf(
				"Request was invalid, likely due to an incorrectly specified parameter. Please check your inputs and try again. (status %d)",
				terr.Code,
			)
		case http.StatusNotFound:
			originalErr = fmt.Errorf(
				"Alertmanager was unable to find the requested resource. Please check your inputs and try again. (status %d)",
				terr.Code,
			)
		case http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			originalErr = fmt.Errorf(
				"Alertmanager is currently unavailable. Please try again later. (status %d)",
				terr.Code,
			)
		}
	}

	err := s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			Flags: disgord.MessageFlagEphemeral,
			Embeds: []*disgord.Embed{{
				Type:        disgord.EmbedTypeRich,
				Color:       colorError,
				Title:       title,
				Description: originalErr.Error(),
			}},
			// Components: []*disgord.MessageComponent{{
			// 	Type: disgord.MessageComponentActionRow,
			// 	Components: []*disgord.MessageComponent{
			// 		{
			// 			Type:     disgord.MessageComponentButton,
			// 			Label:    "retry",
			// 			Style:    disgord.Primary,
			// 			CustomID: "retry",
			// 			Disabled: false,
			// 		},
			// 	},
			// }},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}
