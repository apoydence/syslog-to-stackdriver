package web

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/logging"
)

type Drain struct {
	c Converter
	l Logger
}

type Converter func(data []byte) (logging.Entry, error)

type Logger interface {
	Log(e logging.Entry)
}

func NewDrain(c Converter, l Logger) http.Handler {
	return &Drain{
		c: c,
		l: l,
	}
}

func (d *Drain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		d.writeError(w, errors.New("method must be POST"))
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		d.writeError(w, err)
		return
	}

	e, err := d.c(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		d.writeError(w, err)
		return
	}

	d.l.Log(e)
}

func (d *Drain) writeError(w io.Writer, err error) {
	w.Write([]byte(fmt.Sprintf(`{"error":%q}`, err)))
}
