// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/go-openapi/strfmt"
	"github.com/lrstanley/discord-alertmanager/internal/alertmanager"
	"github.com/lrstanley/discord-alertmanager/internal/models"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
)

func (b *Bot) silenceEditFromCallback(s disgord.Session, h *disgord.InteractionCreate, _ string, args []string) {
	if len(args) < 1 {
		return
	}

	getParams := &silence.GetSilenceParams{}
	getParams.SetContext(b.ctx)
	getParams.SetTimeout(httpRequestTimeout)
	getParams.SetSilenceID(strfmt.UUID(args[0]))
	resp, err := b.al.Silence.GetSilence(getParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silence", err)
		return
	}

	b.modalAdd(s, h, fmt.Sprintf("modal-edit/%s/%d", args[0], h.Message.ID), "Update silence", &addConfig{
		id:       args[0],
		comment:  *resp.Payload.Comment,
		matchers: strings.Join(alertmanager.MatcherToString(resp.Payload.Matchers, false), "\n"),
		startsAt: time.Until(time.Time(*resp.Payload.StartsAt)).Round(time.Minute).String(),
		endsAt:   time.Until(time.Time(*resp.Payload.EndsAt)).Round(time.Minute).String(),
	})
}

func (b *Bot) silenceEditFromCommand(s disgord.Session, h *disgord.InteractionCreate) {
	id, _ := optionsHasChild[string](h.Data.Options, "id")

	getParams := &silence.GetSilenceParams{}
	getParams.SetContext(b.ctx)
	getParams.SetTimeout(httpRequestTimeout)
	getParams.SetSilenceID(strfmt.UUID(id))
	resp, err := b.al.Silence.GetSilence(getParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silence", err)
		return
	}

	config := &addConfig{
		id:       id,
		comment:  *resp.Payload.Comment,
		matchers: strings.Join(alertmanager.MatcherToString(resp.Payload.Matchers, false), "\n"),
		startsAt: time.Until(time.Time(*resp.Payload.StartsAt)).Round(time.Minute).String(),
		endsAt:   time.Until(time.Time(*resp.Payload.EndsAt)).Round(time.Minute).String(),
	}

	wantsModal := true

	if v, ok := optionsHasChild[string](h.Data.Options, "comment"); ok {
		config.comment = v
		wantsModal = false
	}

	if v, ok := optionsHasChild[string](h.Data.Options, "filter"); ok {
		config.matchers = v
		wantsModal = false
	}

	if v, ok := optionsHasChild[string](h.Data.Options, "at"); ok {
		config.startsAt = v
		wantsModal = false
	}

	if v, ok := optionsHasChild[string](h.Data.Options, "until"); ok {
		config.endsAt = v
		wantsModal = false
	}

	// If they only provided the ID via the command, then we want to show the
	// modal, otherwise we can just update the silence with the fields they provided.
	if wantsModal {
		b.modalAdd(s, h, fmt.Sprintf("modal-edit/%s", id), "Update silence", config)
		return
	}

	_ = b.addOrUpdateSilence(s, h, config)
}

func (b *Bot) silenceEditFromModalCallback(s disgord.Session, h *disgord.InteractionCreate, _ string, args []string) {
	if len(args) < 1 {
		return
	}

	var messageID disgord.Snowflake
	if len(args) > 1 {
		messageID = disgord.ParseSnowflakeString(args[1])
	}

	config := &addConfig{}
	config.id = args[0]
	config.comment, _ = componentsHasChild[string](h.Data.Components, "comment")
	config.matchers, _ = componentsHasChild[string](h.Data.Components, "matcher")
	config.startsAt, _ = componentsHasChild[string](h.Data.Components, "startsAt")
	config.endsAt, _ = componentsHasChild[string](h.Data.Components, "endsAt")

	if b.addOrUpdateSilence(s, h, config) && !messageID.IsZero() {
		// Remove all of the buttons from the previous message.
		_, err := b.client.Channel(h.ChannelID).Message(messageID).Update(&disgord.UpdateMessage{
			Components: models.Ptr([]*disgord.MessageComponent{}),
		})
		if err != nil {
			b.logger.WithError(err).Error("failed to update message")
		}
	}
}
