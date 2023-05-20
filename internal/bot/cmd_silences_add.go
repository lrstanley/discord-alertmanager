// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/go-openapi/strfmt"
	"github.com/lrstanley/discord-alertmanager/internal/alertmanager"
	"github.com/lrstanley/discord-alertmanager/internal/models"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	almodels "github.com/prometheus/alertmanager/api/v2/models"
)

const defaultSilenceDuration = 2 * time.Hour

func (b *Bot) silenceAdd(s disgord.Session, h *disgord.InteractionCreate) {
	comment, _ := optionsHasChild[string](h.Data.Options[0].Options, "comment")
	filter, _ := optionsHasChild[string](h.Data.Options[0].Options, "filter")

	matchers, err := alertmanager.ParseLabels(filter, true)
	if err != nil {
		b.responseError(s, h, "Invalid filter provided", err)
		return
	}

	createParams := &silence.PostSilencesParams{}
	createParams.SetContext(b.ctx)
	createParams.SetTimeout(httpRequestTimeout)
	createParams.SetSilence(&almodels.PostableSilence{
		Silence: almodels.Silence{
			Comment:   models.Ptr(comment),
			CreatedBy: models.Ptr(fmt.Sprintf("<@%d> (%s)", h.Member.User.ID, h.Member.User.Username)),
			Matchers:  matchers,
			StartsAt:  models.Ptr(strfmt.DateTime(time.Now())),
			EndsAt:    models.Ptr(strfmt.DateTime(time.Now().Add(defaultSilenceDuration))),
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
	getParams.SetTimeout(httpRequestTimeout)
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

var (
	reAlertWebhook = regexp.MustCompile(`(?sm)^Alerts (?:Firing|Resolved):\nLabels:\n(.*?)\n(?:Annotations|Source):.*`)
	reWebhookLabel = regexp.MustCompile(`.*-\s+([^\s=]+)\s+=\s+(.+)`)
)

func (b *Bot) silenceAddFromMessage(s disgord.Session, h *disgord.InteractionCreate) {
	if h.Data.Resolved == nil || h.Data.Resolved.Messages == nil || len(h.Data.Resolved.Messages) == 0 {
		b.responseError(s, h, "No messages were provided", nil)
		return
	}

	for _, msg := range h.Data.Resolved.Messages {
		for _, embed := range msg.Embeds {
			var rawLabels string

			matches := reAlertWebhook.FindAllStringSubmatch(embed.Description, -1)

			for _, match := range matches {
				for _, matchLabel := range reWebhookLabel.FindAllStringSubmatch(match[1], -1) {
					rawLabels += fmt.Sprintf("%s=%s\n", matchLabel[1], strconv.Quote(matchLabel[2]))
				}
			}

			matchers, err := alertmanager.ParseLabels(rawLabels, false)
			if err != nil {
				uri, _ := msg.DiscordURL()
				if uri != "" {
					b.responseError(s, h, "Invalid filter within message", fmt.Errorf("message: %s %w", uri, err))
				} else {
					b.responseError(s, h, "Invalid filter within message", err)
				}
				return
			}

			if len(matchers) == 0 {
				continue
			}

			b.modalAdd(s, h, "modal-add", "Create silence", &modalAddConfig{
				Matchers: strings.Join(alertmanager.MatcherToString(matchers, false), "\n"),
				StartsAt: "TODO",
				EndsAt:   "TODO",
			})
			return //nolint:staticcheck
		}
	}

	b.responseError(s, h, "No alerts found in message", errors.New("Please use the `/silences add` command instead.")) //nolint:revive,stylecheck
}

type modalAddConfig struct {
	ID string

	Comment  string
	Matchers string
	StartsAt string
	EndsAt   string
}

func (b *Bot) modalAdd(s disgord.Session, h *disgord.InteractionCreate, customID, title string, config *modalAddConfig) {
	err := s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackModal,
		Data: &disgord.CreateInteractionResponseData{
			Title:    title,
			Flags:    disgord.MessageFlagEphemeral,
			CustomID: customID,
			Components: []*disgord.MessageComponent{
				{
					Type: disgord.MessageComponentActionRow,
					Components: []*disgord.MessageComponent{{
						Type:        disgord.MessageComponentTextInput,
						Style:       disgord.TextInputStyleShort,
						Required:    true,
						CustomID:    "comment",
						Label:       "Silence comment",
						Placeholder: "Why are you silencing this alert?",
						Value:       config.Comment,
					}},
				},
				{
					Type: disgord.MessageComponentActionRow,
					Components: []*disgord.MessageComponent{{
						Type:        disgord.MessageComponentTextInput,
						Style:       disgord.TextInputStyleParagraph,
						Required:    true,
						CustomID:    "matcher",
						Label:       "Silence matcher (multiline/comma-separated)",
						Placeholder: "key=\"value\"\nkey2!=\"value2\"\nkey3=~\"value[34]\"\netc...",
						Value:       config.Matchers,
					}},
				},
				{
					Type: disgord.MessageComponentActionRow,
					Components: []*disgord.MessageComponent{{
						Type:        disgord.MessageComponentTextInput,
						Style:       disgord.TextInputStyleShort,
						Required:    true,
						CustomID:    "startsAt",
						Label:       "Starts at",
						Placeholder: "TODO",
						Value:       config.StartsAt,
					}},
				},
				{
					Type: disgord.MessageComponentActionRow,
					Components: []*disgord.MessageComponent{{
						Type:        disgord.MessageComponentTextInput,
						Style:       disgord.TextInputStyleShort,
						Required:    true,
						CustomID:    "endsAt",
						Label:       "Ends at",
						Placeholder: "TODO",
						Value:       config.EndsAt,
					}},
				},
			},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}
