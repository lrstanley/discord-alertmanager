// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/andersfylling/disgord"
)

var reSilenceID = regexp.MustCompile(`\b[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}\b`)

// interactionHasSilence searches through the interaction's resolved messages and
// embeds for a silence ID. If found, it returns the ID, otherwise an empty string.
func interactionHasSilence(data *disgord.ApplicationCommandInteractionData) (id string) {
	fn := func(inputs ...string) string {
		for _, input := range inputs {
			match := reSilenceID.FindString(input)
			if match == "" {
				continue
			}

			return match
		}

		return ""
	}

	for _, msg := range data.Resolved.Messages {
		if id = fn(msg.Content); id != "" {
			return id
		}

		for _, embed := range msg.Embeds {
			if id = fn(embed.Title, embed.Description, embed.URL); id != "" {
				return id
			}

			for _, field := range embed.Fields {
				if id = fn(field.Name, field.Value); id != "" {
					return id
				}
			}
		}

		// Don't check components for now.
	}

	return ""
}

// optionsHasChild recursively searches through application command options (and it's children)
// for an option with the given name. If found, it returns the option's value and true.
func optionsHasChild[T any](options []*disgord.ApplicationCommandDataOption, name string) (v T, ok bool) {
	// First recurse through top-level options.
	for _, opt := range options {
		if !strings.EqualFold(opt.Name, name) {
			continue
		}

		return opt.Value.(T), true
	}

	// If any options have children options, recurse through those as well.
	for _, opt := range options {
		if len(opt.Options) < 1 {
			continue
		}

		v, ok = optionsHasChild[T](opt.Options, name)
		if ok {
			return v, ok
		}
	}

	return v, false
}

// componentHasChild recursively searches through message components (and it's children)
// for a component with the given ID. If found, it returns the component's value and true.
func componentsHasChild[T string | bool | int64 | float64](components []*disgord.MessageComponent, id string) (v T, ok bool) {
	// First recurse through top-level components.
	for _, comp := range components {
		if !strings.EqualFold(comp.CustomID, id) {
			continue
		}

		switch any(v).(type) {
		case string:
			return any(comp.Value).(T), true
		case bool:
			result, err := strconv.ParseBool(comp.Value)
			if err != nil {
				return v, false
			}

			return any(result).(T), true
		case int64:
			result, err := strconv.ParseInt(comp.Value, 10, 64)
			if err != nil {
				return v, false
			}

			return any(result).(T), true
		case float64:
			result, err := strconv.ParseFloat(comp.Value, 64)
			if err != nil {
				return v, false
			}

			return any(result).(T), true
		default:
			panic(errors.New("unknown type"))
		}
	}

	// If any components have children components, recurse through those as well.
	for _, comp := range components {
		if len(comp.Components) < 1 {
			continue
		}

		v, ok = componentsHasChild[T](comp.Components, id)
		if ok {
			return v, ok
		}
	}

	return v, false
}
