// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"time"

	"github.com/andersfylling/disgord"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
)

func (b *Bot) silenceList(s disgord.Session, h *disgord.InteractionCreate) {
	filter, _ := optionsHasChild[string](h.Data.Options[0].Options, "filter")
	includeExpired, _ := optionsHasChild[bool](h.Data.Options[0].Options, "include-expired")
	expiredOnly, _ := optionsHasChild[bool](h.Data.Options[0].Options, "expired-only")

	params := &silence.GetSilencesParams{}
	params.SetContext(b.ctx)
	params.SetTimeout(5 * time.Second)

	if filter != "" {
		params.SetFilter([]string{filter})
	}

	silences, err := b.al.Silence.GetSilences(params, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silences", err)
		return
	}

	var embeds []*disgord.Embed

	for _, alertSilence := range silences.Payload {
		if expiredOnly && *alertSilence.Status.State != "expired" {
			continue
		} else if !expiredOnly && !includeExpired && *alertSilence.Status.State != "active" {
			continue
		}

		embeds = append(embeds, b.silenceEmbed(s, alertSilence))
	}

	if len(embeds) == 0 {
		embeds = append(embeds, &disgord.Embed{
			Type:  disgord.EmbedTypeRich,
			Color: colorInfo,
			Title: "No active silences",
		})
	}

	err = s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			Flags:           disgord.MessageFlagEphemeral,
			Embeds:          embeds,
			AllowedMentions: &disgord.AllowedMentions{Parse: []string{"users"}},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}
