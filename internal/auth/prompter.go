package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// HuhPrompter implements Prompter using charmbracelet/huh for interactive forms.
type HuhPrompter struct{}

// NewTerminalPrompter creates a new HuhPrompter for interactive terminal prompts.
func NewTerminalPrompter() *HuhPrompter {
	return &HuhPrompter{}
}

// Print outputs text to the user.
func (p *HuhPrompter) Print(message string) {
	fmt.Println(message)
}

// PromptSecret prompts for secret input with masked display.
func (p *HuhPrompter) PromptSecret(prompt string) (string, error) {
	var value string

	err := huh.NewInput().
		Title(prompt).
		EchoMode(huh.EchoModePassword).
		Value(&value).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", errors.New("canceled by user")
		}
		return "", fmt.Errorf("prompt input: %w", err)
	}

	return strings.TrimSpace(value), nil
}

// PromptChoice prompts user to select from options and returns the 0-based index.
func (p *HuhPrompter) PromptChoice(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return 0, errors.New("no options provided")
	}

	// Build huh options with display labels and index values
	huhOptions := make([]huh.Option[int], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, i)
	}

	var selected int

	err := huh.NewSelect[int]().
		Title(prompt).
		Options(huhOptions...).
		Value(&selected).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return 0, errors.New("canceled by user")
		}
		return 0, fmt.Errorf("prompt choice: %w", err)
	}

	return selected, nil
}
