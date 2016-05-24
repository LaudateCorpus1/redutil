package pubsub

import (
	"fmt"
	"strconv"
	"strings"
)

// EventType is used to distinguish between pattern and plain text events.
type EventType int

const (
	PlainEvent EventType = iota
	PatternEvent
)

// SubCommand returns the command issued to subscribe to the event in Redis.
func (e EventType) SubCommand() string {
	switch e {
	case PlainEvent:
		return "SUBSCRIBE"
	case PatternEvent:
		return "PSUBSCRIBE"
	default:
		panic("unknown event type")
	}
}

// UnsubCommand returns the command issued
// o unsubscribe from the event in Redis.
func (e EventType) UnsubCommand() string {
	switch e {
	case PlainEvent:
		return "UNSUBSCRIBE"
	case PatternEvent:
		return "PUNSUBSCRIBE"
	default:
		panic("unknown event type")
	}
}

// Fields are concatenated into events which can
// be listened to over liveloading.
type Field struct {
	valid bool
	alias string
	value string
}

// As sets the alias of the field in the event list. You may then call
// Event.Find(alias) to look up the value of the field in the event.
func (f Field) As(alias string) Field {
	return Field{valid: f.valid, value: f.value, alias: alias}
}

// IsZero returns true if the field is empty. A call to Event.Find() or
// Event.Get() with a non-existent alias or index will return such a struct.
func (f Field) IsZero() bool { return !f.valid }

// String returns the field value as a string.
func (f Field) String() string { return f.value }

// String returns the field value as a byte slice.
func (f Field) Bytes() []byte { return []byte(f.value) }

// Int attempts to parse and return the field value as an integer.
func (f Field) Int() (int, error) {
	x, err := strconv.ParseInt(f.value, 10, 32)
	return int(x), err
}

// Uint64 attempts to parse and return the field value as a uint64.
func (f Field) Uint64() (uint64, error) { return strconv.ParseUint(f.value, 10, 64) }

// Int64 attempts to parse and return the field value as a int64.
func (f Field) Int64() (int64, error) { return strconv.ParseInt(f.value, 10, 64) }

// String creates and returns a Field containing a string.
func String(str string) Field { return Field{valid: true, value: str} }

// Int creates and returns a Field containing an integer.
func Int(x int) Field { return Field{valid: true, value: strconv.Itoa(x)} }

// Star returns a field containing the Kleene star `*` for pattern subscription.
func Star() Field { return Field{valid: true, value: "*"} }

// An Event is passed to an Emitter to manage which
// events a Listener is subscribed to.
type Event struct {
	fields []Field
	kind   EventType
}

// Len returns the number of fields contained in the event.
func (e Event) Len() int { return len(e.fields) }

// Get returns the value of a field at index `i` within the event. If the
// field does not exist, an empty struct will be returned.
func (e Event) Get(i int) Field {
	if len(e.fields) <= i {
		return Field{valid: false}
	}

	return e.fields[i]
}

// Find looks up a field value by its alias. This is most useful in pattern
// subscriptions where might use Find to look up a parameterized property.
// If the alias does not exist, an empty struct will be returned.
func (e Event) Find(alias string) Field {
	for _, field := range e.fields {
		if field.alias == alias {
			return field
		}
	}

	return Field{valid: false}
}

// Name returns name of the event, formed by a concatenation of all the
// event fields.
func (e Event) Name() string {
	strs := make([]string, len(e.fields))
	for i, field := range e.fields {
		strs[i] = field.value
	}

	return strings.Join(strs, "")
}

// Returns the type of the event.
func (e Event) Type() EventType {
	return e.kind
}

// toFieldFromString attempts to convert v from a string or byte slice into
// a Field, if it isn't already one. It panics if v is none of the above.
func toFieldFromString(v interface{}) Field {
	switch t := v.(type) {
	case string:
		return String(t)
	case []byte:
		return String(string(t))
	case Field:
		return t
	default:
		panic(fmt.Sprintf("Expected string or field when creating an event, got %T", v))
	}
}

// NewEvent creates and returns a new event based off the series of fields.
// This translates to a Redis SUBSCRIBE call.
func NewEvent(name interface{}, fields ...Field) Event {
	return Event{
		fields: append([]Field{toFieldFromString(name)}, fields...),
		kind:   PlainEvent,
	}
}

// NewPatternEvent creates and returns a new event pattern off the series
// of fields. This translates to a Redis PSUBSCRIBE call.
func NewPatternEvent(name interface{}, fields ...Field) Event {
	return Event{
		fields: append([]Field{toFieldFromString(name)}, fields...),
		kind:   PatternEvent,
	}
}
