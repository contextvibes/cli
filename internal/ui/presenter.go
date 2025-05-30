// internal/ui/presenter.go

package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color" // Import the color library
)

// Presenter handles structured writing to standard output and standard error,
// and reading standardized user input, mimicking the Pulumi CLI style.
type Presenter struct {
	// Interface fields for flexibility and testing
	outW io.Writer
	errW io.Writer
	inR  io.Reader

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
// Color support is automatically detected and disabled if the terminal doesn't support it
// or if the NO_COLOR environment variable is set.
// If outW, errW, or inR are nil, they default to os.Stdout, os.Stderr, and os.Stdin respectively.
func NewPresenter(outW, errW io.Writer, inR io.Reader) *Presenter {
	// *** CORRECTED VARIABLE DECLARATIONS AND ASSIGNMENTS ***
	// Declare local variables with the correct INTERFACE types
	var out io.Writer = os.Stdout
	var err io.Writer = os.Stderr
	var in io.Reader = os.Stdin

	// Assign parameters ONLY if they are not nil, overwriting defaults
	if outW != nil {
		out = outW
	}
	if errW != nil {
		err = errW
	}
	if inR != nil {
		in = inR
	}
	// *********************************************************

	// Initialize and return the struct, assigning interface values to interface fields
	return &Presenter{
		outW: out, // Assign io.Writer to io.Writer field
		errW: err, // Assign io.Writer to io.Writer field
		inR:  in,  // Assign io.Reader to io.Reader field

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

// --- Output Stream Getters ---

// Out returns the configured output writer (typically os.Stdout).
func (p *Presenter) Out() io.Writer {
	return p.outW
}

// Err returns the configured error writer (typically os.Stderr).
func (p *Presenter) Err() io.Writer {
	return p.errW
}

// --- Output Formatting Methods ---
// These methods correctly use p.outW and p.errW which are io.Writer interfaces

func (p *Presenter) Header(format string, a ...any) { p.headerColor.Fprintf(p.outW, format+"\n", a...) }
func (p *Presenter) Summary(format string, a ...any) {
	p.summaryColor.Fprint(p.outW, "SUMMARY:\n")
	fmt.Fprintf(p.outW, "  "+format+"\n", a...)
	p.Newline()
}
func (p *Presenter) Step(format string, a ...any) {
	p.stepColor.Fprintf(p.outW, "- "+format+"\n", a...)
}
func (p *Presenter) Info(format string, a ...any) {
	p.infoColor.Fprintf(p.outW, "~ "+format+"\n", a...)
}
func (p *Presenter) InfoPrefixOnly() { p.infoColor.Fprint(p.outW, "~ ") }
func (p *Presenter) Success(format string, a ...any) {
	p.successColor.Fprintf(p.outW, "+ "+format+"\n", a...)
}
func (p *Presenter) Error(format string, a ...any) {
	p.errorColor.Fprintf(p.errW, "! "+format+"\n", a...)
}
func (p *Presenter) Warning(format string, a ...any) {
	p.warningColor.Fprintf(p.errW, "~ "+format+"\n", a...)
}
func (p *Presenter) Advice(format string, a ...any) {
	p.warningColor.Fprintf(p.outW, "~ "+format+"\n", a...)
}
func (p *Presenter) Detail(format string, a ...any) {
	p.detailColor.Fprintf(p.outW, "  "+format+"\n", a...)
}
func (p *Presenter) Highlight(text string) string { return p.boldColor.Sprint(text) }
func (p *Presenter) Newline()                     { fmt.Fprintln(p.outW) }
func (p *Presenter) Separator() {
	color.New(color.Faint).Fprintln(p.outW, "----------------------------------------")
}

// --- Input Methods ---
// These methods correctly use p.inR which is an io.Reader interface

func (p *Presenter) PromptForInput(prompt string) (string, error) {
	reader := bufio.NewReader(p.inR) // Use interface field
	prompt = strings.TrimSpace(prompt)
	if !strings.HasSuffix(prompt, ":") {
		prompt += ":"
	}
	prompt += " "
	p.promptColor.Fprint(p.errW, prompt) // Write prompt to error stream
	input, err := reader.ReadString('\n')
	if err != nil {
		p.errorColor.Fprintf(p.errW, "\n! Error reading input: %v\n", err)
		return "", fmt.Errorf("reading input failed: %w", err)
	}
	return strings.TrimSpace(input), nil
}

func (p *Presenter) PromptForConfirmation(prompt string) (bool, error) {
	reader := bufio.NewReader(p.inR) // Use interface field
	prompt = strings.TrimSpace(prompt)
	if !strings.HasSuffix(prompt, "?") {
		prompt += "?"
	}
	fullPrompt := prompt + " [y/N]: "
	for {
		p.promptColor.Fprint(p.errW, fullPrompt) // Write prompt to error stream
		input, err := reader.ReadString('\n')
		if err != nil {
			p.errorColor.Fprintf(p.errW, "\n! Error reading confirmation: %v\n", err)
			return false, fmt.Errorf("reading confirmation failed: %w", err)
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" || input == "" {
			return false, nil
		}
		p.warningColor.Fprintf(p.errW, "~ Invalid input. Please enter 'y' or 'n'.\n") // Write warning to error stream
	}
}
