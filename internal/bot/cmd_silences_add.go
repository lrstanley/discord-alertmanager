// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"fmt"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/go-openapi/strfmt"
	"github.com/lrstanley/discord-alertmanager/internal/alertmanager"
	"github.com/lrstanley/discord-alertmanager/internal/models"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	almodels "github.com/prometheus/alertmanager/api/v2/models"
)

func (b *Bot) silenceAdd(s disgord.Session, h *disgord.InteractionCreate) {
	comment, _ := optionsHasChild[string](h.Data.Options[0].Options, "comment")
	filter, _ := optionsHasChild[string](h.Data.Options[0].Options, "filter")

	matchers, err := alertmanager.ParseLabels(filter)
	if err != nil {
		b.responseError(s, h, "Invalid filter provided", err)
		return
	}

	createParams := &silence.PostSilencesParams{}
	createParams.SetContext(b.ctx)
	createParams.SetTimeout(5 * time.Second)
	createParams.SetSilence(&almodels.PostableSilence{
		Silence: almodels.Silence{
			Comment:   models.Ptr(comment),
			CreatedBy: models.Ptr(fmt.Sprintf("<@%d> (%s)", h.Member.User.ID, h.Member.User.Username)),
			Matchers:  matchers,
			StartsAt:  models.Ptr(strfmt.DateTime(time.Now())),
			EndsAt:    models.Ptr(strfmt.DateTime(time.Now().Add(2 * time.Hour))),
		},
	})

	createResp, err := b.al.Silence.PostSilences(createParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while creating silence", err)
		return
	}

	// Assuming there were no issues, refetch to get status info.

	getParams := &silence.GetSilenceParams{}
	getParams.SetContext(b.ctx)
	getParams.SetTimeout(5 * time.Second)
	getParams.SetSilenceID(strfmt.UUID(createResp.Payload.SilenceID))
	resp, err := b.al.Silence.GetSilence(getParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silence", err)
		return
	}

	silenceEmbed := b.silenceEmbed(s, resp.Payload)
	silenceEmbed.Color = colorSuccess
	silenceEmbed.Title = fmt.Sprintf("Silence created: %s", *resp.Payload.ID)

	err = s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			Embeds:          []*disgord.Embed{silenceEmbed},
			AllowedMentions: &disgord.AllowedMentions{Parse: []string{"users"}},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}
