package logging

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var levelMap = map[string]zerolog.Level{
	"fatal": zerolog.FatalLevel,
	"error": zerolog.ErrorLevel,
	"warn":  zerolog.WarnLevel,
	"info":  zerolog.InfoLevel,
	"debug": zerolog.DebugLevel,
}

var Logger zerolog.Logger

func init() {

	// https://github.com/akamai/cli
	AKAMAI_LOG := strings.ToLower(os.Getenv("AKAMAI_LOG"))
	if AKAMAI_LOG == "" {
		AKAMAI_LOG = "info"
	}

	// The logger logs to STDERR on purpose, because at the end of every execution, the program will write the hosts file value to STDOUT.
	// Logging to STDERR therefore allow stream redirections such as pipes, with no undesired output.
	Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger().Level(levelMap[AKAMAI_LOG])

}
