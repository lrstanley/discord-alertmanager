// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"fmt"

	"github.com/andersfylling/disgord"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
)

func (b *Bot) silenceRemove(s disgord.Session, h *disgord.InteractionCreate) {
	id, _ := optionsHasChild[string](h.Data.Options[0].Options, "id")

	// First get the silence, so we can show it in the response to make it clear
	// to others in the same channel what was removed.
	getParams := &silence.GetSilenceParams{}
	getParams.SetContext(b.ctx)
	getParams.SetTimeout(httpRequestTimeout)
	getParams.SetSilenceID(strfmt.UUID(id))
	resp, err := b.al.Silence.GetSilence(getParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silence", err)
		return
	}

	deleteParams := &silence.DeleteSilenceParams{}
	deleteParams.SetContext(b.ctx)
	deleteParams.SetTimeout(httpRequestTimeout)
	deleteParams.SetSilenceID(strfmt.UUID(id))

	_, err = b.al.Silence.DeleteSilence(deleteParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while deleting silence", err)
		return
	}

	silenceEmbed := b.silenceEmbed(s, resp.Payload)
	silenceEmbed.Color = colorError
	silenceEmbed.Title = "Silence removed"
	silenceEmbed.URL = ""
	silenceEmbed.Description = fmt.Sprintf("Silence `%s` has been removed.\n%s", id, silenceEmbed.Description)

	err = s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			Embeds: []*disgord.Embed{silenceEmbed},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}
