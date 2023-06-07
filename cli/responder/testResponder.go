package responder

import "os"

// FixedResponse always returns the given response. This is expected to be
// useful for testing. Note that there are no checks made of the Response and
// so it is possible to have a response that cannot be made from the standard
// Responder such as an uppercase value or a whitespace character.
type FixedResponse struct {
	Response rune
	Err      error
}

// GetResponse returns the fixed responses
func (fr FixedResponse) GetResponse() (rune, error) {
	return fr.Response, fr.Err
}

// GetResponseOrDie returns the fixed responses. It will exit if the Err
// field is not nil.
func (fr FixedResponse) GetResponseOrDie() rune {
	if fr.Err != nil {
		os.Exit(errExitStatus)
	}
	return fr.Response
}

// GetResponseIndent returns the fixed responses
func (fr FixedResponse) GetResponseIndent(_, _ int) (rune, error) {
	return fr.Response, fr.Err
}

// GetResponseIndentOrDie returns the fixed responses. It will exit if the
// Err field is not nil.
func (fr FixedResponse) GetResponseIndentOrDie(_, _ int) rune {
	return fr.GetResponseOrDie()
}
