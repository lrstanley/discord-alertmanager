// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package alertmanager

import (
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/lrstanley/discord-alertmanager/internal/models"
	almodels "github.com/prometheus/alertmanager/api/v2/models"
	"golang.org/x/exp/slices"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "StringSingle", Pattern: `'(?:\\.|[^'])*'`},
		{Name: "StringDouble", Pattern: `"(?:\\.|[^"])*"`},
		{Name: "Equality", Pattern: `(!=|=~|!~|=)`},
		{Name: "Separator", Pattern: `[ \t\n\r,]+`},
	})
	parser = participle.MustBuild[ParseResults](
		participle.Lexer(lex),
		participle.Elide("Separator"),
		participle.Unquote("StringSingle", "StringDouble"),
		participle.UseLookahead(2),
	)
	excludeLabelNames = []string{"alertstate"}
)

// EBNF equivalent:
//
//	ParseResults = (LabelEntry* | (LabelEntry ("," LabelEntry)+)) .
//	LabelEntry = <ident> <equality> (<stringsingle> | <stringdouble> | <ident>) .
//
// This parser allows grabbing out labels in all sorts of ways, reducing the chance of someone
// accidentally passing in invalid syntax. Example:
//
//	foo="bar", bar='bar'
//	foo=~"bar" foo!~"bar"
//	foo!="^foo\"test\"bar[^baz]+$"
//	foo=bar123

type ParseResults struct {
	Entries []*LabelEntry `parser:"( @@* | @@ ( ',' @@ )+ )"`
}

type LabelEntry struct {
	Name    string `parser:"@Ident"`
	Matcher string `parser:"@Equality"`
	Value   string `parser:"( @StringSingle | @StringDouble | @Ident )"`
}

func (l *LabelEntry) IsEqual() *bool {
	return models.Ptr(strings.HasPrefix(l.Matcher, "="))
}

func (l *LabelEntry) IsRegex() *bool {
	return models.Ptr(strings.HasSuffix(l.Matcher, "~"))
}

func (l *LabelEntry) String() string {
	return l.Name + l.Matcher + strconv.Quote(l.Value)
}

// ParseLabels parses a string of labels into a list of matchers. If allowDuplicates
// is true, then duplicate matchers will be allowed, otherwise the last seen value
// will be used.
func ParseLabels(input string, allowDuplicates bool) (matchers []*almodels.Matcher, err error) {
	var ast *ParseResults

	ast, err = parser.ParseString("", input)
	if err != nil {
		return nil, err
	}

	for _, entry := range ast.Entries {
		matcher := &almodels.Matcher{
			Name:    &entry.Name,
			Value:   &entry.Value,
			IsEqual: entry.IsEqual(),
			IsRegex: entry.IsRegex(),
		}

		// Check to see if the matcher is already in the list. If it is, overwrite
		// using the last seen value.
		var found bool
		for _, m := range matchers {
			if m.Name != matcher.Name {
				continue
			}

			m.Value = matcher.Value
			m.IsEqual = matcher.IsEqual
			m.IsRegex = matcher.IsRegex

			found = true
			break
		}

		if !found || allowDuplicates {
			matchers = append(matchers, matcher)
		}
	}

	return matchers, nil
}

// Convert matcher key-value pairs. If pad is true, then the output will be
// padded to be aligned with each other (keys and equal/regex/negation).
func MatcherToString(matchers []*almodels.Matcher, pad bool) (out []string) {
	var nameLen, equalLen int

	for _, m := range matchers {
		if slices.Contains(excludeLabelNames, *m.Name) {
			continue
		}

		if len(*m.Name) > nameLen {
			nameLen = len(*m.Name)
		}

		if !*m.IsEqual || *m.IsRegex {
			equalLen = 2
		} else if equalLen == 0 {
			equalLen = 1
		}
	}

	for _, m := range matchers {
		if slices.Contains(excludeLabelNames, *m.Name) {
			continue
		}

		s := *m.Name

		if pad {
			s += strings.Repeat(" ", nameLen-len(*m.Name)) + " "
		}

		if *m.IsEqual {
			s += "="
		} else {
			s += "!"
		}

		equal := 1
		if *m.IsRegex {
			s += "~"
			equal = 2
		} else if !*m.IsEqual {
			s += "="
			equal = 2
		}

		if pad {
			s += strings.Repeat(" ", (equalLen+1)-equal)
		}

		s += strconv.Quote(*m.Value)
		out = append(out, s)
	}

	return out
}
