package responder

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"syscall"
	"unicode"

	"github.com/nickwells/twrap.mod/twrap"
	"golang.org/x/term"
)

// Responder describes the methods offered to get a response
type Responder interface {
	GetResponse() (rune, error)
	GetResponseOrDie() rune
	GetResponseIndent(int, int) (rune, error)
	GetResponseIndentOrDie(int, int) rune
}

const (
	helpRune      = '?'
	errExitStatus = 1
)

// R holds the details needed to collect and validate a response
type R struct {
	prompt string

	validResps map[rune]string
	hasDflt    bool
	dflt       rune

	maxReprompts int
	limitPrompts bool

	fd  int
	rdr *bufio.Reader

	indent      int
	indentFirst int
}

// RespOptFunc is a function which can be passed to the New function
// to set optional parts of the R
type RespOptFunc func(*R) error

// SetDefault sets the default value for a Responder
func SetDefault(d rune) RespOptFunc {
	return func(r *R) error {
		if _, ok := r.validResps[d]; !ok {
			return fmt.Errorf(
				"SetDefault: the default response (%c) is not"+
					" in the list of valid responses",
				d)
		}

		r.dflt = d
		r.hasDflt = true

		return nil
	}
}

// SetMaxReprompts sets the maximum number of times that the user
// will be reprompted for a valid response before reporting an error. The
// value must be greater than 0
func SetMaxReprompts(maximum int) RespOptFunc {
	return func(r *R) error {
		if maximum <= 0 {
			return fmt.Errorf(
				"SetMaxReprompts: the maximum number of"+
					" reprompts (%d) must be greater than 0",
				maximum)
		}

		r.maxReprompts = maximum
		r.limitPrompts = true

		return nil
	}
}

// SetIndents sets the indents for the first and subsequent lines of output
func SetIndents(indentFirst, indent int) RespOptFunc {
	return func(r *R) error {
		if indent < 0 {
			return fmt.Errorf(
				"SetIndents: the indent (%d) must be"+
					" greater than or equal to 0",
				indent)
		}

		if indentFirst < 0 {
			return fmt.Errorf(
				"SetIndents: the first indent (%d) must be"+
					" greater than or equal to 0",
				indentFirst)
		}

		r.indent = indent
		r.indentFirst = indentFirst

		return nil
	}
}

// NewOrPanic creates a new responder and panics if there are any errors
func NewOrPanic(
	prompt string,
	responses map[rune]string,
	opts ...RespOptFunc,
) *R {
	r, err := New(prompt, responses, opts...)
	if err != nil {
		panic(err)
	}

	return r
}

// New creates a responder and verifies that it is correct
func New(
	prompt string,
	responses map[rune]string,
	opts ...RespOptFunc,
) (*R, error) {
	r := &R{
		prompt: prompt,
		fd:     syscall.Stdin,
		rdr:    bufio.NewReader(os.Stdin),
	}

	if len(responses) <= 1 {
		return nil,
			fmt.Errorf("too few allowed responses - there must be at least 2")
	}

	for v := range responses {
		if unicode.IsUpper(v) {
			return nil,
				fmt.Errorf(
					"only lowercase responses are allowed - '%c' is uppercase",
					v)
		}

		if unicode.IsSpace(v) {
			return nil,
				fmt.Errorf(
					"a whitespace character is not an allowed response" +
						" - it is used to select the default response")
		}

		if v == helpRune {
			return nil,
				fmt.Errorf(
					"'%c' is not an allowed response"+
						" - it is used to request help",
					helpRune)
		}
	}

	r.validResps = responses

	for _, o := range opts {
		err := o(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// PrintValidResponses prints the valid response runes separated
// by a slash.
//
// A rune matching the default is shown in brackets (like so: [y]).
func (r R) PrintValidResponses() {
	fmt.Print("(")

	responses := r.getSortedValidResponses()

	sep := ""

	if r.hasDflt {
		fmt.Printf("[%c]", r.dflt)

		sep = "/"
	}

	for _, c := range responses {
		if r.hasDflt && c == r.dflt {
			continue
		}

		fmt.Printf("%s%c", sep, c)

		sep = "/"
	}

	fmt.Printf("%s%c): ", sep, helpRune)
}

// PrintPrompt prints the prompt and any valid responses.
func (r R) PrintPrompt() {
	fmt.Print(r.prompt)
	fmt.Print("? ")

	r.PrintValidResponses()
}

// getSortedValidResponses gets the valid responses in lexicographic order
func (r R) getSortedValidResponses() []rune {
	keys := make([]rune, 0, len(r.validResps))

	for k := range r.validResps {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

// PrintHelp prints the help message.
func (r R) PrintHelp() {
	r.PrintHelpIndent(r.indent)
}

// PrintHelpIndent prints the help message.
func (r R) PrintHelpIndent(indent int) {
	twc := twrap.NewTWConfOrPanic()

	twc.Println() //nolint: errcheck
	twc.Wrap("Enter one of:", indent)

	keys := r.getSortedValidResponses()

	const charFmt = "%c  "

	const listIndent = 4

	if r.hasDflt {
		twc.WrapPrefixed(
			fmt.Sprintf(charFmt, r.dflt),
			fmt.Sprintf("%s (this is the default)", r.validResps[r.dflt]),
			indent+listIndent)
	}

	for _, k := range keys {
		if r.hasDflt && r.dflt == k {
			continue
		}

		twc.WrapPrefixed(
			fmt.Sprintf(charFmt, k),
			r.validResps[k],
			indent+listIndent)
	}

	twc.WrapPrefixed(
		fmt.Sprintf(charFmt, helpRune),
		"to show this message\n",
		indent+listIndent)
	twc.Wrap("to select the default either enter the character or whitespace"+
		" (a space, tab or return character)",
		indent)
}

// GetResponseOrDie calls GetResponse to get the response but if there is an
// error it will print it and exit with status 1.
func (r R) GetResponseOrDie() rune {
	return r.GetResponseIndentOrDie(r.indentFirst, r.indent)
}

// GetResponseIndentOrDie calls GetResponseIndent to get the response but if
// there is an error it will print it and exit with status 1.
func (r R) GetResponseIndentOrDie(first, second int) rune {
	resp, err := r.GetResponseIndent(first, second)
	if err != nil {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr,
			strings.Repeat(" ", r.indent)+"    "+err.Error())
		os.Exit(errExitStatus)
	}

	return resp
}

// GetResponse will print the prompt and read a single rune from standard
// input. It will check that the rune is a valid response. If it is not in
// the set of valid responses it will print an error message and reprompt. It
// will do this maxReprompts times before returning an error.
//
// If an error is detected the response returned will be the unicode
// ReplacementChar.
func (r R) GetResponse() (response rune, err error) {
	return r.GetResponseIndent(r.indentFirst, r.indent)
}

// GetResponseIndent behaves as GetResponse but the indents are taken from
// the parameters rather than the responder.
func (r R) GetResponseIndent(first, second int) (response rune, err error) {
	i := 0

	prefix := strings.Repeat(" ", first)
	secondPrefix := strings.Repeat(" ", second)

	for {
		fmt.Print(prefix)

		prefix = secondPrefix

		r.PrintPrompt()

		response, err = r.getResp()
		if response == helpRune {
			r.PrintHelpIndent(second)
			continue
		}

		i++

		if err == nil || err == io.EOF {
			return
		}

		if r.limitPrompts && i > r.maxReprompts {
			return
		}

		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, prefix+"    "+err.Error())
	}
}

// getRune gets the response and performs any mappings and display of help
func (r R) getRune() (rune, error) {
	state, err := term.MakeRaw(r.fd)
	if err == nil {
		defer term.Restore(r.fd, state) //nolint: errcheck
	}

	resp, _, err := r.rdr.ReadRune()

	return resp, err
}

// getResp gets the response and performs any mappings and display of help
func (r R) getResp() (rune, error) {
	resp, err := r.getRune()
	if err != nil {
		return unicode.ReplacementChar, err
	}

	if r.hasDflt && unicode.IsSpace(resp) {
		resp = r.dflt
	} else if resp != helpRune {
		resp = unicode.ToLower(resp)

		if _, ok := r.validResps[resp]; !ok {
			return unicode.ReplacementChar,
				fmt.Errorf("Bad response: %c", resp)
		}
	}

	return resp, nil
}
