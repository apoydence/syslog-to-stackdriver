package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/logging"
	"github.com/poy/syslog-to-stackdriver/pkg/web"
)

func TestDrain(t *testing.T) {
	t.Parallel()

	spyConverter := newSpyConverter()
	spyConverter.e = logging.Entry{Timestamp: time.Unix(1, 0)}

	spyLogger := newSpyLogger()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"http://some.target",
		strings.NewReader("some-data"),
	)

	// Assert Drain is a http.Handler
	var handler http.Handler = web.NewDrain(spyConverter.Convert, spyLogger)

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("wrong: %v", recorder.Code)
	}

	if string(spyConverter.data) != "some-data" {
		t.Fatalf("wrong: %s", spyConverter.data)
	}

	if !reflect.DeepEqual(spyLogger.e, spyConverter.e) {
		t.Fatalf("wrong: %v", spyLogger.e)
	}
}

func TestDrainInvalidMethod(t *testing.T) {
	t.Parallel()

	spyConverter := newSpyConverter()
	spyConverter.e = logging.Entry{Timestamp: time.Unix(1, 0)}

	spyLogger := newSpyLogger()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPut,
		"http://some.target",
		strings.NewReader("some-data"),
	)

	// Assert Drain is a http.Handler
	var handler http.Handler = web.NewDrain(spyConverter.Convert, spyLogger)

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("wrong: %v", recorder.Code)
	}
}

func TestDrainErrorConverting(t *testing.T) {
	t.Parallel()

	spyConverter := newSpyConverter()
	spyConverter.err = errors.New("some-error")
	spyConverter.e = logging.Entry{Timestamp: time.Unix(1, 0)}

	spyLogger := newSpyLogger()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"http://some.target",
		strings.NewReader("some-data"),
	)

	// Assert Drain is a http.Handler
	var handler http.Handler = web.NewDrain(spyConverter.Convert, spyLogger)

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("wrong: %v", recorder.Code)
	}

	if reflect.DeepEqual(spyLogger.e, spyConverter.e) {
		t.Fatalf("wrong: %v", spyLogger.e)
	}
}

type spyConverter struct {
	data []byte
	e    logging.Entry
	err  error
}

func newSpyConverter() *spyConverter {
	return &spyConverter{}
}

func (s *spyConverter) Convert(data []byte) (logging.Entry, error) {
	s.data = data
	return s.e, s.err
}

type spyLogger struct {
	e logging.Entry
}

func newSpyLogger() *spyLogger {
	return &spyLogger{}
}

func (s *spyLogger) Log(e logging.Entry) {
	s.e = e
}
