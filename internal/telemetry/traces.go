package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// EventType represents the name of trace events used for tracking
// different stages of message processing and operations.
type EventType string

const (
	// SendACKEvent represents the acknowledgment sending event.
	SendACKEvent = "send_ack"
	// HandleMentionsEvent represents the event for handling bot mentions.
	HandleMentionsEvent = "handle_mentions"
	// NonThreadPostEphemeralEvent represents posting ephemeral messages outside threads.
	NonThreadPostEphemeralEvent = "non_thread_post_ephemeral"
	// ProcessThreadEvent represents the thread processing event.
	ProcessThreadEvent = "process_thread"
	// GetConversationRepliesEvent represents fetching conversation replies.
	GetConversationRepliesEvent = "get_conversation_replies"
	// SummarizeThreadEvent represents the thread summarization event.
	SummarizeThreadEvent = "summarize_thread"
	// UploadFileV2Event represents the file upload event using v2 API.
	UploadFileV2Event = "upload_file_v2"
)

// StartEvent adds a start event marker to the given trace span with a stack trace.
// It appends "_start" suffix to the event name for tracking the beginning of an operation.
func StartEvent(t trace.Span, evt EventType) {
	t.AddEvent(string(evt) + "_start")
}

// EndEvent adds an end event marker to the given trace span with a stack trace.
// It appends "_end" suffix to the event name for tracking the completion of an operation.
func EndEvent(t trace.Span, evt EventType) {
	t.AddEvent(string(evt) + "_end")
}

// WrapErrorWithTrace records an error in the trace span and optionally wraps it with context.
//
// Parameters:
//   - t: The trace span to record the error in
//   - call: Optional context string describing the call site (e.g., "GetConversationReplies")
//   - err: The error to record and potentially wrap
//
// Behavior:
//   - Records the original error in the span with a stack trace
//   - Sets the span status to error
//   - If call is non-empty, wraps the error with the call context using fmt.Errorf
//   - If call is empty, returns the original unwrapped error
//
// Returns the wrapped error (or original error if call is empty).
func WrapErrorWithTrace(t trace.Span, call string, err error) error {
	rErr := err
	if call != "" {
		rErr = fmt.Errorf("%s: %w", call, err)
	}

	t.RecordError(err, trace.WithStackTrace(true))
	t.SetStatus(codes.Error, rErr.Error())

	return rErr
}
