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

var (
	reAlertWebhook = regexp.MustCompile(`(?sm)^Alerts (?:Firing|Resolved):\nLabels:\n(.*?)\n(?:Annotations|Source):.*`)
	reWebhookLabel = regexp.MustCompile(`.*-\s+([^\s=]+)\s+=\s+(.+)`)
)

type addConfig struct {
	ID string // Only used when editing.

	Comment  string
	Matchers string
	StartsAt string
	EndsAt   string
}

func (m *addConfig) validate() error {
	if m.Comment == "" {
		return errors.New("comment is required")
	}

	if m.Matchers == "" {
		return errors.New("matchers are required")
	}

	if m.StartsAt == "" {
		m.StartsAt = "TODO"
		// return errors.New("startsAt is required")
	}

	if m.EndsAt == "" {
		m.EndsAt = "TODO"
		// return errors.New("endsAt is required")
	}

	return nil
}

func (b *Bot) addOrUpdateSilence(s disgord.Session, h *disgord.InteractionCreate, config *addConfig) (ok bool) { //nolint:unparam
	if err := config.validate(); err != nil {
		b.responseError(s, h, "Invalid silence configuration provided", err)
		return false
	}

	matchers, err := alertmanager.ParseLabels(config.Matchers, true)
	if err != nil {
		b.responseError(s, h, "Invalid filter/matchers provided", err)
		return false
	}

	createParams := &silence.PostSilencesParams{}
	createParams.SetContext(b.ctx)
	createParams.SetTimeout(httpRequestTimeout)
	createParams.SetSilence(&almodels.PostableSilence{
		ID: config.ID,
		Silence: almodels.Silence{
			Comment:   models.Ptr(config.Comment),
			CreatedBy: models.Ptr(fmt.Sprintf("<@%d> (%s)", h.Member.User.ID, h.Member.User.Username)),
			Matchers:  matchers,
			StartsAt:  models.Ptr(strfmt.DateTime(time.Now())), // TODO: swap out to actual input time.
			EndsAt:    models.Ptr(strfmt.DateTime(time.Now().Add(defaultSilenceDuration))),
		},
	})

	createResp, err := b.al.Silence.PostSilences(createParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while creating/updating silence", err)
		return false
	}

	// Assuming there were no issues, refetch to get status info.

	getParams := &silence.GetSilenceParams{}
	getParams.SetContext(b.ctx)
	getParams.SetTimeout(httpRequestTimeout)
	getParams.SetSilenceID(strfmt.UUID(createResp.Payload.SilenceID))
	resp, err := b.al.Silence.GetSilence(getParams, b.al.HandleAuth)
	if err != nil {
		b.responseError(s, h, "An error occurred while fetching silence", err)
		return false
	}

	silenceEmbed := b.silenceEmbed(s, resp.Payload)
	silenceEmbed.Color = colorSuccess
	if config.ID == "" {
		silenceEmbed.Title = fmt.Sprintf("Silence created: %s", *resp.Payload.ID)
	} else {
		silenceEmbed.Title = fmt.Sprintf("Silence updated: %s", *resp.Payload.ID)
		// TODO: may have to remove [%s](%s) depending on if Discord continues to support markdown format.
		silenceEmbed.Description = fmt.Sprintf("replaces silence: [%s](%s)\n", config.ID, b.al.SilenceURL(config.ID)) + silenceEmbed.Description
	}

	err = s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			AllowedMentions: &disgord.AllowedMentions{Parse: []string{"users"}},
			Embeds:          []*disgord.Embed{silenceEmbed},
			Components: []*disgord.MessageComponent{{
				Type: disgord.MessageComponentActionRow,
				Components: []*disgord.MessageComponent{
					{
						Type:     disgord.MessageComponentButton,
						Label:    "edit",
						Style:    disgord.Primary,
						CustomID: fmt.Sprintf("silence-edit/%s", createResp.Payload.SilenceID),
						Disabled: false,
					},
					{
						Type:     disgord.MessageComponentButton,
						Label:    "remove",
						Style:    disgord.Danger,
						CustomID: fmt.Sprintf("silence-remove/%s", createResp.Payload.SilenceID),
						Disabled: false,
					},
				},
			}},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
		return false
	}

	return true
}

func (b *Bot) silenceAddFromCommand(s disgord.Session, h *disgord.InteractionCreate) {
	config := &addConfig{}
	config.Comment, _ = optionsHasChild[string](h.Data.Options, "comment")
	config.Matchers, _ = optionsHasChild[string](h.Data.Options, "filter")
	config.StartsAt, _ = optionsHasChild[string](h.Data.Options, "at")
	config.EndsAt, _ = optionsHasChild[string](h.Data.Options, "until")

	if config.Comment == "" || config.Matchers == "" {
		b.modalAdd(s, h, "modal-add", "Create silence", config)
		return
	}

	_ = b.addOrUpdateSilence(s, h, config)
}

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
				if errors.Is(err, alertmanager.ErrNoLabelsProvided) {
					continue
				}

				uri, _ := msg.DiscordURL()
				if uri != "" {
					b.responseError(s, h, "Invalid filter within message", fmt.Errorf("message: %s %w", uri, err))
				} else {
					b.responseError(s, h, "Invalid filter within message", err)
				}
				return
			}

			b.modalAdd(s, h, "modal-add", "Create silence", &addConfig{
				Matchers: strings.Join(alertmanager.MatcherToString(matchers, false), "\n"),
			})
			return //nolint:staticcheck
		}
	}

	b.responseError(s, h, "No alerts found in message", errors.New("Please use the `/silences add` command instead.")) //nolint:revive,stylecheck
}

func (b *Bot) silenceAddFromModalCallback(s disgord.Session, h *disgord.InteractionCreate, _ string, _ []string) {
	config := &addConfig{}
	config.Comment, _ = componentsHasChild[string](h.Data.Components, "comment")
	config.Matchers, _ = componentsHasChild[string](h.Data.Components, "matcher")
	config.StartsAt, _ = componentsHasChild[string](h.Data.Components, "startsAt")
	config.EndsAt, _ = componentsHasChild[string](h.Data.Components, "endsAt")

	_ = b.addOrUpdateSilence(s, h, config)
}

func (b *Bot) modalAdd(s disgord.Session, h *disgord.InteractionCreate, customID, title string, config *addConfig) {
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
