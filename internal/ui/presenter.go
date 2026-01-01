// Package ui provides the presentation layer for the CLI, handling output formatting and user interaction.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"

	//nolint:depguard // isatty is standard for terminal detection.
	"github.com/mattn/go-isatty"
)

// Presenter handles structured writing to standard output and standard error,
// and reading standardized user input, mimicking the Pulumi CLI style.
type Presenter struct {
	outW io.Writer
	errW io.Writer

	// Color instances (initialized in New)
	successColor *color.Color
	errorColor   *color.Color
	warningColor *color.Color
	infoColor    *color.Color
	stepColor    *color.Color
	detailColor  *color.Color
	promptColor  *color.Color
	headerColor  *color.Color
	boldColor    *color.Color
	summaryColor *color.Color
}

// NewPresenter creates a new Console instance with Pulumi-like color support.
func NewPresenter(outW, errW io.Writer) *Presenter {
	var out io.Writer = os.Stdout
	if outW != nil {
		out = outW
	}

	var err io.Writer = os.Stderr
	if errW != nil {
		err = errW
	}

	return &Presenter{
		outW: out,
		errW: err,

		// Initialize all color fields
		successColor: color.New(color.FgGreen, color.Bold),
		errorColor:   color.New(color.FgRed, color.Bold),
		warningColor: color.New(color.FgYellow),
		infoColor:    color.New(color.FgBlue),
		stepColor:    color.New(color.FgWhite),
		detailColor:  color.New(color.Faint),
		promptColor:  color.New(color.FgCyan),
		headerColor:  color.New(color.Bold, color.Underline),
		boldColor:    color.New(color.Bold),
		summaryColor: color.New(color.FgCyan, color.Bold),
	}
}

// Out returns the configured output writer (typically os.Stdout).
func (p *Presenter) Out() io.Writer {
	return p.outW
}

// Err returns the configured error writer (typically os.Stderr).
func (p *Presenter) Err() io.Writer {
	return p.errW
}

// --- Output Formatting Methods ---

// Header prints a header.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Header(format string, a ...any) {
	_, _ = p.headerColor.Fprintf(p.outW, format+"\n", a...)
}

// Summary prints a summary.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Summary(format string, a ...any) {
	_, _ = p.summaryColor.Fprint(p.outW, "SUMMARY:\n")

	_, _ = fmt.Fprintf(p.outW, "  "+format+"\n", a...)
	p.Newline()
}

// Step prints a step.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Step(format string, a ...any) {
	_, _ = p.stepColor.Fprintf(p.outW, "- "+format+"\n", a...)
}

// Info prints info.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Info(format string, a ...any) {
	_, _ = p.infoColor.Fprintf(p.outW, "~ "+format+"\n", a...)
}

// InfoPrefixOnly prints the info prefix without a newline.
func (p *Presenter) InfoPrefixOnly() { _, _ = p.infoColor.Fprint(p.outW, "~ ") }

// Success prints success.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Success(format string, a ...any) {
	_, _ = p.successColor.Fprintf(p.outW, "+ "+format+"\n", a...)
}

// Error prints an error.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Error(format string, a ...any) {
	_, _ = p.errorColor.Fprintf(p.errW, "! "+format+"\n", a...)
}

// Warning prints a warning.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Warning(format string, a ...any) {
	_, _ = p.warningColor.Fprintf(p.errW, "~ "+format+"\n", a...)
}

// Advice prints advice.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Advice(format string, a ...any) {
	_, _ = p.warningColor.Fprintf(p.outW, "~ "+format+"\n", a...)
}

// Detail prints detail.
//
//nolint:goprintffuncname // Noun-based naming is intended.
func (p *Presenter) Detail(format string, a ...any) {
	_, _ = p.detailColor.Fprintf(p.outW, "  "+format+"\n", a...)
}

// Highlight highlights text.
func (p *Presenter) Highlight(text string) string { return p.boldColor.Sprint(text) }

// Newline prints a newline.
func (p *Presenter) Newline() { _, _ = fmt.Fprintln(p.outW) }

// Separator prints a separator.
func (p *Presenter) Separator() {
	_, _ = color.New(color.Faint).Fprintln(p.outW, "----------------------------------------")
}

// --- Input Methods ---

// PromptForInput prompts the user for input.
func (p *Presenter) PromptForInput(prompt string) (string, error) {
	interactiveReader, cleanup, err := p.getInteractiveReader()
	if err != nil {
		p.Error("Could not get an interactive terminal for prompting: %v", err)

		return "", err
	}
	defer cleanup()

	reader := bufio.NewReader(interactiveReader)

	prompt = strings.TrimSpace(prompt)
	if !strings.HasSuffix(prompt, ":") {
		prompt += ":"
	}

	prompt += " "

	_, _ = p.promptColor.Fprint(p.errW, prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		_, _ = p.errorColor.Fprintf(p.errW, "\n! Error reading input: %v\n", err)

		return "", fmt.Errorf("reading input failed: %w", err)
	}

	return strings.TrimSpace(input), nil
}

// PromptForConfirmation prompts the user for confirmation.
func (p *Presenter) PromptForConfirmation(prompt string) (bool, error) {
	interactiveReader, cleanup, err := p.getInteractiveReader()
	if err != nil {
		p.Error("Could not get an interactive terminal for prompting: %v", err)

		return false, err
	}
	defer cleanup()

	reader := bufio.NewReader(interactiveReader)

	prompt = strings.TrimSpace(prompt)
	if !strings.HasSuffix(prompt, "?") {
		prompt += "?"
	}

	fullPrompt := prompt + " [y/N]: "
	for {
		_, _ = p.promptColor.Fprint(p.errW, fullPrompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			_, _ = p.errorColor.Fprintf(p.errW, "\n! Error reading confirmation: %v\n", err)

			return false, fmt.Errorf("reading confirmation failed: %w", err)
		}

		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true, nil
		}

		if input == "n" || input == "no" || input == "" {
			return false, nil
		}

		_, _ = p.warningColor.Fprintf(
			p.errW,
			"~ Invalid input. Please enter 'y' or 'n'.\n",
		)
	}
}

// PromptForSelect displays a single-choice list to the user and returns the selected option.
func (p *Presenter) PromptForSelect(title string, options []string) (string, error) {
	var choice string

	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	form := huh.NewSelect[string]().
		Title(title).
		Options(huhOptions...).
		Value(&choice)

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("selection failed: %w", err)
	}

	return choice, nil
}

// PromptForMultiSelect displays a multi-choice list to the user and returns the selected options.
func (p *Presenter) PromptForMultiSelect(title string, options []string) ([]string, error) {
	var choices []string

	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	form := huh.NewMultiSelect[string]().
		Title(title).
		Options(huhOptions...).
		Value(&choices)

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("multi-selection failed: %w", err)
	}

	return choices, nil
}

// getInteractiveReader intelligently selects the correct input for user prompts.
// If stdin is a pipe, it opens /dev/tty for interactive input. Otherwise, it uses stdin.
// The returned cleanup function MUST be called by the caller to close /dev/tty if it was opened.
//

func (p *Presenter) getInteractiveReader() (io.Reader, func(), error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, ttyErr := os.Open("/dev/tty")
		if ttyErr != nil {
			return nil, func() {}, fmt.Errorf(
				"stdin is a pipe and could not open /dev/tty for interactive prompt: %w",
				ttyErr,
			)
		}

		return tty, func() { _ = tty.Close() }, nil
	}

	return os.Stdin, func() {}, nil
}
