// Package plugins defines support for plugins.
//
// Each IRC command should have an associated symbol (function name) in the plugin. This function should have this
// signature:
// func Example(args []string, e *irc.Event) (reply string, action bool)
//
// Each plugin must have a config file which defines IRC commands and the functions, matching the signature above, that
// implements them:
//
// ```
// command1: function1
// command2: function2
// ```
package plugins
