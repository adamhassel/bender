# Plugins for Bender

The Bender bot supports plugins, written as standard Golang plugins.

Plugins can implement either "matchers" or "commands".

## Matchers

A matcher is a function that works on whatever is written in a channel. An
example can be a function that finds any web link in a message, and posts a
shortened version of it.

If you want to export one or more matchers in your plugin, you must include an
exportable variable with the names of matcher functions (which must also be
exported):

```golang
var Matchers = []string{"ExampleMatcher"}
// later...
func ExampleMatcher...
```

Matcher functions must have this signature:

```golang
func(string, *irc.Event) (string, bool)
```
The first argument is the message posted to the channel. The second argument is
an irc event (see `github.com/thoj/go-ircevent`). The return is a string
message and a bool indicating if the returned message should be considered an
action (`/me` style message).

## Commands

Commands are explicit commands, prefixed with the configured command char.
Matchers will always ignore anything that is a command.

Commands are matched with a function in a YAML file, read at config time. For
example, if your function, in code is:

```golang
func Example...
```

and the command you want to bind to this function is `command`, then your coinfig YAML file would simply be

```yaml
command:Example
```
Note that case matters. And the command function MUST be exported.

Command functions must have this signature:

```golang
func([]string, *irc.Event) (string, bool)
```
The first argument is all the arguments the command has received, minus the
command itself. The second argument is an irc event (see
`github.com/thoj/go-ircevent`). The return is a string message and a bool
indicating if the returned message should be considered an action (`/me` style
message).

