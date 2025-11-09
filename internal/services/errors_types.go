package services

import "errors"

type commandType string

// CommandSummarize is the command that tells handleMentions to run slackMessageProcessor's message handler.
const CommandSummarize commandType = "summarize"

var (
	// ErrInvalidCommandType returned by handleMentions in case of an unimplemented CommandType occures.
	ErrInvalidCommandType = errors.New("invalid command type")

	errIgnoredInvalidAPI   = errors.New("ignored invalid evets api data")
	errHandleEvent         = errors.New("failed to handle event")
	errNotImplementedEvent = errors.New("not implemented events api event received")
)
