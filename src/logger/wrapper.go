package logger

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
)

func WrapProcess(executable string, arg ...string) {
	fdlLogger := NewLogger("Logs wrapper")
	defer handlePanic(fdlLogger)

	r, w, err := os.Pipe()
	if err != nil {
		fdlLogger.Fatal().Err(err).Msg("Could not create pipe for logs")
		os.Exit(1)
	}

	cmd := exec.Command(executable, arg...)
	cmd.Stderr = w

	if err = cmd.Start(); err != nil {
		fdlLogger.Fatal().Err(err).Msg("Could not launch main process")
		os.Exit(1)
	}
	exitCodeCh := make(chan int)
	logsCh := make(chan []byte)

	go waitForCommandToExit(cmd, fdlLogger, exitCodeCh)
	go collectLogs(r, fdlLogger, logsCh)

	panicLogsBuilder := strings.Builder{}
	foundPanic := false
	for {
		select {
		case exitCode := <-exitCodeCh:
			handleExit(exitCode, panicLogsBuilder.String(), fdlLogger)
		case logsLineBytes := <-logsCh:
			foundPanic = handleLogLine(logsLineBytes, foundPanic, &panicLogsBuilder, fdlLogger)
		}
	}
}

func waitForCommandToExit(cmd *exec.Cmd, fdlLogger zerolog.Logger, exitCodeCh chan<- int) {
	defer handlePanic(fdlLogger)
	err := cmd.Wait()
	if err == nil {
		exitCodeCh <- 0
		return
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		exitCodeCh <- 1
		return
	}
	exitCodeCh <- exitErr.ExitCode()
}

func collectLogs(r *os.File, fdlLogger zerolog.Logger, logsCh chan<- []byte) {
	defer handlePanic(fdlLogger)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		logsCh <- line
	}
	if err := scanner.Err(); err != nil {
		fdlLogger.Fatal().Err(err).Msg("Error scanning piped main process's Stderr")
		os.Exit(1)
	}
}

func handleExit(exitCode int, panicLogs string, fdlLogger zerolog.Logger) {
	if exitCode == 0 {
		fdlLogger.Info().Msg("Exited with code 0")
	} else {
		fdlLogger.
			Fatal().
			Err(errors.New(panicLogs)).
			Msgf("Panicked and exited with code: %d", exitCode)
	}
	os.Exit(exitCode)
}

func handleLogLine(logsLineBytes []byte, foundPanic bool, builder *strings.Builder, fdlLogger zerolog.Logger) bool {
	logsLine := string(logsLineBytes)
	if !foundPanic && strings.HasPrefix(logsLine, "panic") {
		foundPanic = true
	}
	switch {
	case len(logsLineBytes) == 0:
		return foundPanic
	case foundPanic:
		builder.WriteString(fmt.Sprintf("%s\n", logsLine))
	case isJSON(logsLineBytes):
		println(logsLine)
	default:
		fdlLogger.Error().Msgf("Got log line that is not JSON formatted: '%s'", logsLine)
	}
	return foundPanic
}

func handlePanic(fdlLogger zerolog.Logger) {
	r := recover()
	if r == nil {
		return
	}
	fdlLogger.Fatal().
		Caller().
		Str("error", fmt.Sprint(r)).
		Str("stack_trace", string(debug.Stack())).
		Msg("Program panicked and exited")
}

func isJSON(b []byte) bool {
	var js json.RawMessage
	err := json.Unmarshal(b, &js)
	return err == nil && js != nil
}
