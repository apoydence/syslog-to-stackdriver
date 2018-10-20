package conversion_test

import (
	"testing"
	"time"

	"cloud.google.com/go/logging"
	"code.cloudfoundry.org/rfc5424"
	"github.com/poy/syslog-to-stackdriver/pkg/conversion"
)

func TestConversion(t *testing.T) {
	t.Parallel()

	msg := rfc5424.Message{
		Priority:  rfc5424.Error,
		Timestamp: time.Unix(1234, 0),
		Hostname:  "some-host",
		AppName:   "some-app",
		ProcessID: "some-process",
		MessageID: "some-message-id",

		// Trim whitespace
		Message: []byte(" some-message\n"),
	}

	data, err := msg.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	e, err := conversion.Convert(data)
	if err != nil {
		t.Fatal(err)
	}

	if e.Timestamp.UnixNano() != int64(1234*time.Second) {
		t.Fatalf("wrong: %v", e.Timestamp.UnixNano())
	}

	if e.Severity != logging.Error {
		t.Fatalf("wrong: %v", e.Severity)
	}

	m, ok := e.Payload.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string payload: %T", e.Payload)
	}

	assertValue := func(k, v string) {
		if m[k] != v {
			t.Fatalf("expected %s to equal %s", m[k], v)
		}
	}

	assertValue("host_name", "some-host")
	assertValue("app_name", "some-app")
	assertValue("process_id", "some-process")
	assertValue("message_id", "some-message-id")
	assertValue("message", "some-message")
}

func TestConversionSeverity(t *testing.T) {
	t.Parallel()

	assert := func(p rfc5424.Priority, expected logging.Severity) {
		t.Helper()
		msg := rfc5424.Message{
			Priority:  p,
			Timestamp: time.Unix(0, 1234),
			Hostname:  "some-host",
			AppName:   "some-app",
			ProcessID: "some-process",
			MessageID: "some-message-id",
			Message:   []byte("some-message"),
		}

		data, err := msg.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		e, err := conversion.Convert(data)
		if err != nil {
			t.Fatal(err)
		}

		if e.Severity != expected {
			t.Fatalf("expected %s to equal %s", e.Severity, expected)
		}
	}

	assert(rfc5424.Emergency, logging.Emergency)
	assert(rfc5424.Alert, logging.Alert)
	assert(rfc5424.Crit, logging.Critical)
	assert(rfc5424.Error, logging.Error)
	assert(rfc5424.Warning, logging.Warning)
	assert(rfc5424.Notice, logging.Notice)
	assert(rfc5424.Info, logging.Info)
	assert(rfc5424.Debug, logging.Debug)
	assert(999, logging.Default)
}
