// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bot

import (
	"strings"

	"github.com/andersfylling/disgord"
)

func optionsHasChild[T any](options []*disgord.ApplicationCommandDataOption, name string) (v T, ok bool) {
	for _, opt := range options {
		if strings.EqualFold(opt.Name, name) {
			return opt.Value.(T), true
		}
	}
	return v, false
}
