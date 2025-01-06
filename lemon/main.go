package lemon

import (
	"fmt"
	"regexp"
	"strings"
)

var Version string
var Usage = fmt.Sprintf(`Usage: lemonade [options]... SUB_COMMAND [arg]
Sub Commands:
  open [URL]                  Open URL by browser
  copy [text]                 Copy text.
  paste                       Paste text.
  server                      Start lemonade server.
  tmux [see --tmux-* options] Send keys via tmux.

Options:
  --port=2489                 TCP port number
  --line-ending               Convert Line Ending (CR/CRLF)
  --allow="0.0.0.0/0,::/0"    Allow IP Range                [Server only]
  --host="localhost"          Destination hostname          [Client only]
  --no-fallback-messages      Do not show fallback messages [Client only]
  --trans-loopback=true       Translate loopback address    [open subcommand only]
  --trans-localfile=true      Translate local file path     [open subcommand only]
  --log-level=1               Log level                     [4 = Critical, 0 = Debug]
  --help                      Show this
	--tmux-target               Target tmux target pane       [potential format/values- session:window.pane, window.pane or pane]
	--tmux-send-keys            Keys to send
	--tmux-select-target-pane   Select target pane            [selects the pane specified by --tmux-target after sending keys]

Version:
  %s`, Version)

func ConvertLineEnding(text, option string) string {
	switch option {
	case "lf", "LF":
		text = strings.Replace(text, "\r\n", "\n", -1)
		return strings.Replace(text, "\r", "\n", -1)
	case "crlf", "CRLF":
		text = regexp.MustCompile(`\r(.)|\r$`).ReplaceAllString(text, "\r\n$1")
		text = regexp.MustCompile(`([^\r])\n|^\n`).ReplaceAllString(text, "$1\r\n")
		return text
	default:
		return text
	}
}
