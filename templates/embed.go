// Package templates embeds every file that `eva init` can install.
//
// Layout:
//
//	core/            installed in every project
//	stacks/<id>/     installed when the stack is selected
//	agents/<id>/     installed when the agent is selected
//
// Naming conventions (applied to the file name, suffixes are stripped):
//
//	*.tmpl   rendered with text/template using [[ ]] delimiters
//	*.block  content is merged into a managed block (<!-- eva:start --> ... <!-- eva:end -->)
//	         so user content outside the block is preserved
package templates

import "embed"

// FS holds the embedded template tree. The all: prefix includes dot-directories
// such as .eva/, .claude/ and .opencode/.
//
//go:embed all:core all:stacks all:agents
var FS embed.FS
