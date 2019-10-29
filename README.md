# cli.mod
Various packages to help when writing command line interfaces

## cli/responder
This provides a means of prompting the user and getting a response. The
programmer can specify the allowed responses and optionally a default
value. The package will check that the user has entered a valid response and
will return an error if not. The user can request help on the allowed values
through a standard user interface.
