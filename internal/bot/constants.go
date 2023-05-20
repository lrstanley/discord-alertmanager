// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"github.com/andersfylling/disgord"
	"github.com/lrstanley/discord-alertmanager/internal/models"
)

const (
	colorError   = 0xe50d0d
	colorSuccess = 0x26a25a
	colorInfo    = 0x5865F2 // 0x0a8d8d
	colorExpired = 0x99AAB5
	colorWarning = 0xFEE75C
)

// nolint:gomnd
var commands = []*disgord.CreateApplicationCommand{
	{
		Name:                     "silence alert",
		Type:                     disgord.ApplicationCommandMessage,
		DMPermission:             models.Ptr(false),
		DefaultMemberPermissions: models.Ptr(disgord.PermissionBit(0)),
	},
	{
		Name:        "silences",
		Description: "Manage alert silences",
		// Disallow DMs, which might bypass Discords default permission checks.
		DMPermission: models.Ptr(false),
		// Require admin by default, and let the owner add roles/channels as necessary.
		DefaultMemberPermissions: models.Ptr(disgord.PermissionBit(0)),
		Options: []*disgord.ApplicationCommandOption{
			// {
			// 	Name:        "test_option",
			// 	Type:        disgord.OptionTypeString,
			// 	Description: "testing options",
			// 	Choices: []*disgord.ApplicationCommandOptionChoice{
			// 		{
			// 			Name:  "test_choice",
			// 			Value: "test_val",
			// 		},
			// 	},
			// },
			{
				Name:        "add",
				Description: "Add a new silence",
				Type:        disgord.OptionTypeSubCommand,
				Options: []*disgord.ApplicationCommandOption{
					{
						Name:        "comment",
						Description: "Comment or description to go along with the silence",
						Type:        disgord.OptionTypeString,
						Required:    true,
						MinLength:   4,
					},
					{
						Name:        "filter",
						Description: "Filter silences by label-value pairs. e.g. alertname=\"foo\",bar=\"baz\"",
						Type:        disgord.OptionTypeString,
						Required:    true,
						MinLength:   4,
					},
					{
						Name:        "at",
						Description: "Time at which the silence should start -- defaults to now",
						Type:        disgord.OptionTypeString,
						Required:    false,
					},
					{
						Name:        "until",
						Description: "Time at which the silence should end -- defaults to 2 hours from now",
						Type:        disgord.OptionTypeString,
						Required:    false,
					},
				},
			},
			{
				Name:        "get",
				Description: "Get an existing silence",
				Type:        disgord.OptionTypeSubCommand,
				Options: []*disgord.ApplicationCommandOption{
					{
						Name:        "id",
						Description: "Silence ID to get",
						Type:        disgord.OptionTypeString,
						Required:    true,
						MinLength:   36,
						MaxLength:   36,
					},
				},
			},
			{
				Name:        "list",
				Description: "Lists all existing silences",
				Type:        disgord.OptionTypeSubCommand,
				Options: []*disgord.ApplicationCommandOption{
					{
						Name:        "filter",
						Description: "Filter silences by label-value pairs. e.g. alertname=\"foo\",bar=\"baz\"",
						Type:        disgord.OptionTypeString,
						Required:    false,
						MinLength:   4,
					},
					{
						Name:        "include-expired",
						Description: "Include expired silences in returned results",
						Type:        disgord.OptionTypeBoolean,
						Required:    false,
					},
					{
						Name:        "expired-only",
						Description: "Include only expired silences in returned results",
						Type:        disgord.OptionTypeBoolean,
						Required:    false,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove an existing silence",
				Type:        disgord.OptionTypeSubCommand,
				Options: []*disgord.ApplicationCommandOption{
					{
						Name:        "id",
						Description: "Silence ID to remove",
						Type:        disgord.OptionTypeString,
						Required:    true,
						MinLength:   36,
						MaxLength:   36,
					},
				},
			},
		},
	},
}
