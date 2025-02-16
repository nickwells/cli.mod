package responder_test

import (
	"fmt"

	"github.com/nickwells/cli.mod/cli/responder"
)

func Example() {
	r := responder.NewOrPanic(
		"Question",
		map[rune]string{
			'y': "to show differences",
			'n': "to skip this file",
			'q': "to quit",
		},
		responder.SetDefault('y'),
	)

	for {
		response := r.GetResponseOrDie()

		fmt.Println()

		switch response {
		case 'y':
			fmt.Println("Yes!")
		case 'n':
			fmt.Println("No.")
		case 'q':
			break
		}
	}
}

// This example shows the text that will be printed to prompt the user to
// respond
func ExampleR_PrintPrompt_noDefault() {
	r := responder.NewOrPanic(
		"Delete File",
		map[rune]string{
			'y': "delete the file",
			'n': "leave the file alone",
		},
	)
	r.PrintPrompt()
	// Output:
	// Delete File? (n/y/?):
}

// This example shows the text that will be printed to prompt the user to
// respond. In this example a default response is given
func ExampleR_PrintPrompt_withDefault() {
	r := responder.NewOrPanic(
		"Delete File",
		map[rune]string{
			'y': "delete the file",
			'n': "leave the file alone",
		},
		responder.SetDefault('y'),
	)
	r.PrintPrompt()
	// Output:
	// Delete File? ([y]/n/?):
}
