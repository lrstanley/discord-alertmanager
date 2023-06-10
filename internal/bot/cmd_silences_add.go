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

const defaultSilenceDuration = 4 * time.Hour

var (
	reAlertWebhook = regexp.MustCompile(`(?sm)^Alerts (?:Firing|Resolved):\nLabels:\n(.*?)\n(?:Annotations|Source):.*`)
	reWebhookLabel = regexp.MustCompile(`.*-\s+([^\s=]+)\s+=\s+(.+)`)
)

// parseSilenceTime parses a silence time from a string. It will first attempt to
// parse it as a duration, then as a RFC3339 timestamp, otherwise returning an error.
func parseSilenceTime(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, errors.New("no time provided")
	}

	if strings.EqualFold(input, "now") {
		return time.Now().Local(), nil
	}

	d, err := time.ParseDuration(strings.ToLower(input))
	if err == nil {
		return time.Now().Local().Add(d), nil
	}

	t, err := time.Parse(time.RFC3339, input)
	if err == nil {
		return t, nil
	}

	return time.Time{}, errors.New("unable to parse time")
}

type addConfig struct {
	id string // Only used when editing.

	comment  string
	matchers string
	startsAt string
	endsAt   string

	matchersParsed []*almodels.Matcher
	startsAtParsed time.Time
	endsAtParsed   time.Time
}

func (m *addConfig) validate() (err error) {
	if m.comment == "" {
		return errors.New("comment is required")
	}

	if m.matchers == "" {
		return errors.New("matchers are required")
	}

	m.matchersParsed, err = alertmanager.ParseLabels(m.matchers, true)
	if err != nil {
		return fmt.Errorf("invalid filter/matchers provided: %w", err)
	}

	if m.startsAt == "" {
		m.startsAt = "now"
	}

	m.startsAtParsed, err = parseSilenceTime(m.startsAt)
	if err != nil {
		return fmt.Errorf("invalid startsAt provided: %w", err)
	}

	if m.endsAt == "" {
		m.endsAt = time.Now().Local().Add(defaultSilenceDuration).Format(time.RFC3339)
	}

	m.endsAtParsed, err = parseSilenceTime(m.endsAt)
	if err != nil {
		return fmt.Errorf("invalid endsAt provided: %w", err)
	}

	return nil
}

func (b *Bot) addOrUpdateSilence(s disgord.Session, h *disgord.InteractionCreate, config *addConfig) (ok bool) { //nolint:unparam
	if err := config.validate(); err != nil {
		b.responseError(s, h, "Invalid silence configuration provided", err)
		return false
	}

	createParams := &silence.PostSilencesParams{}
	createParams.SetContext(b.ctx)
	createParams.SetTimeout(httpRequestTimeout)
	createParams.SetSilence(&almodels.PostableSilence{
		ID: config.id,
		Silence: almodels.Silence{
			Comment:   models.Ptr(config.comment),
			CreatedBy: models.Ptr(fmt.Sprintf("<@%d> (%s)", h.Member.User.ID, h.Member.User.Username)),
			Matchers:  config.matchersParsed,
			StartsAt:  models.Ptr(strfmt.DateTime(config.startsAtParsed)),
			EndsAt:    models.Ptr(strfmt.DateTime(config.endsAtParsed)),
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
	if config.id == "" {
		silenceEmbed.Title = fmt.Sprintf("Silence created: %s", *resp.Payload.ID)
	} else {
		silenceEmbed.Title = fmt.Sprintf("Silence updated: %s", *resp.Payload.ID)
		silenceEmbed.Description = fmt.Sprintf("replaces silence: [%s](%s)\n", config.id, b.al.SilenceURL(config.id)) + silenceEmbed.Description
	}

	err = s.SendInteractionResponse(b.ctx, h, &disgord.CreateInteractionResponse{
		Type: disgord.InteractionCallbackChannelMessageWithSource,
		Data: &disgord.CreateInteractionResponseData{
			AllowedMentions: &disgord.AllowedMentions{Parse: []string{"users"}},
			Embeds:          []*disgord.Embed{silenceEmbed},
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
	config.comment, _ = optionsHasChild[string](h.Data.Options, "comment")
	config.matchers, _ = optionsHasChild[string](h.Data.Options, "filter")
	config.startsAt, _ = optionsHasChild[string](h.Data.Options, "at")
	config.endsAt, _ = optionsHasChild[string](h.Data.Options, "until")

	if config.comment == "" || config.matchers == "" {
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
				matchers: strings.Join(alertmanager.MatcherToString(matchers, false), "\n"),
				startsAt: "now",
				endsAt:   time.Now().Local().Add(defaultSilenceDuration).Format(time.RFC3339),
			})
			return //nolint:staticcheck
		}
	}

	b.responseError(s, h, "No alerts found in message", errors.New("Please use the `/silences add` command instead.")) //nolint:revive,stylecheck
}

func (b *Bot) silenceAddFromModalCallback(s disgord.Session, h *disgord.InteractionCreate, _ string, _ []string) {
	config := &addConfig{}
	config.comment, _ = componentsHasChild[string](h.Data.Components, "comment")
	config.matchers, _ = componentsHasChild[string](h.Data.Components, "matcher")
	config.startsAt, _ = componentsHasChild[string](h.Data.Components, "startsAt")
	config.endsAt, _ = componentsHasChild[string](h.Data.Components, "endsAt")

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
						Value:       config.comment,
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
						Value:       config.matchers,
					}},
				},
				{
					Type: disgord.MessageComponentActionRow,
					Components: []*disgord.MessageComponent{{
						Type:        disgord.MessageComponentTextInput,
						Style:       disgord.TextInputStyleShort,
						Required:    true,
						CustomID:    "startsAt",
						Label:       "Starts at (RFC3339 or 1h30m, -1h30m, etc)",
						Placeholder: "RFC3339 or 1h30m, -1h30m, etc",
						Value:       config.startsAt,
					}},
				},
				{
					Type: disgord.MessageComponentActionRow,
					Components: []*disgord.MessageComponent{{
						Type:        disgord.MessageComponentTextInput,
						Style:       disgord.TextInputStyleShort,
						Required:    true,
						CustomID:    "endsAt",
						Label:       "Ends at (RFC3339 or 1h30m, -1h30m, etc)",
						Placeholder: "RFC3339 or 1h30m, -1h30m, etc",
						Value:       config.endsAt,
					}},
				},
			},
		},
	})
	if err != nil {
		b.logger.WithError(err).Error("failed to respond to interaction")
	}
}
