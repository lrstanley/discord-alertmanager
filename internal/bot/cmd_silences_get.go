// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"fmt"
	"regexp"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/apex/log"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	almodels "github.com/prometheus/alertmanager/api/v2/models"
)

func (b *Bot) silenceGet(s disgord.Session, h *disgord.InteractionCreate) {
	id, _ := optionsHasChild[string](h.Data.Options[0].Options, "id")

	getParams := &silence.GetSilenceParams{}
	getParams.SetContext(b.ctx)
	getParams.SetTimeout(5 * time.Second)
	getParams.SetSilenceID(strfmt.UUID(id))
	resp, err := b.al.Silence.GetSilence(getParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silence", err)
		return
	}

	err = s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			Flags:  disgord.MessageFlagEphemeral,
			Embeds: []*disgord.Embed{b.silenceEmbed(s, resp.Payload)},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}

var reDiscordUsername = regexp.MustCompile(`<@!?(\d+)>\s+?\(([^)]+)\)`)

func (b *Bot) silenceEmbed(s disgord.Session, alertSilence *almodels.GettableSilence) *disgord.Embed {
	fields := []*disgord.EmbedField{}

	// All key-value label pairs.
	var description string
	for _, matcher := range alertSilence.Matchers {
		description += fmt.Sprintf("%s = %s\n", *matcher.Name, *matcher.Value)
	}

	fields = append(fields, &disgord.EmbedField{
		Name:   ":memo: Comment",
		Value:  *alertSilence.Comment,
		Inline: false,
	})

	// Timestamps, using Discords timestamp formatting, which automatically handles
	// humanization for us, while also showing the full time in a tooltip.
	if alertSilence.StartsAt != nil {
		fields = append(fields, &disgord.EmbedField{
			Name:   ":watch: Starts",
			Value:  fmt.Sprintf("<t:%d:R>", time.Time(*alertSilence.StartsAt).Unix()),
			Inline: true,
		})
	}
	if alertSilence.EndsAt != nil {
		fields = append(fields, &disgord.EmbedField{
			Name:   ":watch: Ends",
			Value:  fmt.Sprintf("<t:%d:R>", time.Time(*alertSilence.EndsAt).Unix()),
			Inline: true,
		})
	}

	// Adjustments if it's expired.
	titlePrefix := "Silence"
	color := colorInfo
	if *alertSilence.Status.State != "active" {
		titlePrefix = "Expired"
		color = colorExpired
	}

	// Handle Discord usernames, if a Discord user created the silence.
	author := *alertSilence.CreatedBy
	authorID := *alertSilence.CreatedBy
	var iconURL string

	if match := reDiscordUsername.FindAllStringSubmatch(author, -1); len(match) > 0 && len(match[0]) == 3 {
		author = match[0][2]
		authorID = fmt.Sprintf("<@%s>", match[0][1])
		if id := disgord.ParseSnowflakeString(match[0][1]); !id.IsZero() {
			user, err := s.User(id).Get()
			if err == nil {
				iconURL, err = user.AvatarURL(64, true)
			}
			if err != nil {
				b.logger.WithFields(log.Fields{
					"author":    author,
					"author_id": match[0][1],
				}).Warn("failed to fetch user information")
			}
		}
	}

	fields = append(fields, &disgord.EmbedField{
		Name:   ":pencil2: Created by",
		Value:  authorID,
		Inline: true,
	})

	return &disgord.Embed{
		Type:        disgord.EmbedTypeRich,
		Color:       color,
		Title:       fmt.Sprintf("%s: %s", titlePrefix, *alertSilence.ID),
		Description: "```\n" + description + "```",
		URL:         fmt.Sprintf("%s/#/silences/%s", b.al.URL(), *alertSilence.ID),
		Fields:      fields,
		Timestamp:   disgord.Time{Time: time.Time(*alertSilence.UpdatedAt)},
		Footer: &disgord.EmbedFooter{
			Text:    fmt.Sprintf("Created by %s", author),
			IconURL: iconURL,
		},
	}
}
