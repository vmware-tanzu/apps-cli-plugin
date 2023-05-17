// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/AlecAivazis/survey/v2/terminal"
)

// PromptConfig is the configuration for a prompt.
type PromptConfig struct {
	// Message to display to user.
	Message string

	// Options for user to choose from
	Options []string

	// Default option.
	Default string

	// Sensitive information.
	Sensitive bool

	// Help for the prompt.
	Help string
}

// Run the prompt.
func (p *PromptConfig) Run(response interface{}, opts ...PromptOpt) error {
	return Prompt(p, response, opts...)
}

// Prompt for input, reads input value, without trimming any characters (may include leading/tailing spaces)
func Prompt(p *PromptConfig, response interface{}, opts ...PromptOpt) error {
	prompt := translatePromptConfig(p)
	options := defaultPromptOptions()
	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return err
		}
	}

	surveyOpts := translatePromptOpts(options)
	return survey.AskOne(prompt, response, surveyOpts...)
}

func translatePromptConfig(p *PromptConfig) survey.Prompt {
	if p.Sensitive {
		return &survey.Password{
			Message: p.Message,
			Help:    p.Help,
		}
	}
	if len(p.Options) != 0 {
		return &survey.Select{
			Message: p.Message,
			Options: p.Options,
			Default: p.Default,
			Help:    p.Help,
		}
	}
	return &survey.Input{
		Message: p.Message,
		Default: p.Default,
		Help:    p.Help,
	}
}

func defaultPromptOptions() *PromptOptions {
	return &PromptOptions{
		Stdio: terminal.Stdio{
			In:  os.Stdin,
			Out: os.Stdout,
			Err: os.Stderr,
		},
		Icons: survey.IconSet{
			Error: survey.Icon{
				Text:   "X",
				Format: "red",
			},
			Help: survey.Icon{
				Text:   "?",
				Format: "cyan",
			},
			Question: survey.Icon{
				Text:   "?",
				Format: "cyan+b",
			},
			MarkedOption: survey.Icon{
				Text:   "[x]",
				Format: "green",
			},
			UnmarkedOption: survey.Icon{
				Text:   "[ ]",
				Format: "default+hb",
			},
			SelectFocus: survey.Icon{
				Text:   ">",
				Format: "cyan+b",
			},
		},
	}
}

// PromptOptions are options for prompting.
type PromptOptions struct {
	// Standard in/out/error
	Stdio terminal.Stdio
	Icons survey.IconSet

	// Validators on user inputs
	Validators []survey.Validator
}

// PromptOpt is an option for prompts
type PromptOpt func(*PromptOptions) error

func translatePromptOpts(options *PromptOptions) (surveyOpts []survey.AskOpt) {
	surveyOpts = append(surveyOpts, survey.WithStdio(options.Stdio.In, options.Stdio.Out, options.Stdio.Err), survey.WithIcons(func(icons *survey.IconSet) {
		icons.Error = options.Icons.Error
		icons.Question = options.Icons.Question
		icons.Help = options.Icons.Help
		icons.MarkedOption = options.Icons.MarkedOption
		icons.UnmarkedOption = options.Icons.UnmarkedOption
		icons.SelectFocus = options.Icons.SelectFocus
	}))

	// Add validators
	for _, v := range options.Validators {
		surveyOpts = append(surveyOpts, survey.WithValidator(v))
	}

	return
}

// WithStdio specifies the standard input, output and error. By default, these are os.Stdin,
// os.Stdout, and os.Stderr.
func WithStdio(in terminal.FileReader, out terminal.FileWriter, err io.Writer) PromptOpt {
	return func(options *PromptOptions) error {
		options.Stdio.In = in
		options.Stdio.Out = out
		options.Stdio.Err = err
		return nil
	}
}

// WithValidator specifies a validator to use while prompting the user
func WithValidator(v survey.Validator) PromptOpt {
	return func(options *PromptOptions) error {
		// add the provided validator to the list
		options.Validators = append(options.Validators, v)

		// nothing went wrong
		return nil
	}
}

// Custom validators to be used in prompts

// NoUpperCase checks if a given input string contains uppercase characters and returns an error if it does. This function does not panic if the input value is not a string.
func NoUpperCase(v interface{}) error {
	s, ok := v.(string)

	if ok && strings.ToLower(s) != s {
		return errors.New("value contains uppercase characters")
	}

	return nil
}

// NoOnlySpaces checks if a given input string contains only empty spaces and returns an error if it does. This function does not panic if the input value is not a string.
func NoOnlySpaces(v interface{}) error {
	s, ok := v.(string)

	if ok && strings.TrimSpace(s) == "" {
		return errors.New("value contains only spaces")
	}

	return nil
}
