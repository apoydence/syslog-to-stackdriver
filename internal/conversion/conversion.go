package conversion

import (
	"cloud.google.com/go/logging"
	"code.cloudfoundry.org/rfc5424"
)

func Convert(data []byte) (logging.Entry, error) {
	var msg rfc5424.Message
	if err := msg.UnmarshalBinary(data); err != nil {
		return logging.Entry{}, err
	}

	return logging.Entry{
		Timestamp: msg.Timestamp,
		Severity:  priorityToSeverity(msg.Priority),
		Payload: map[string]string{
			"host_name":  msg.Hostname,
			"app_name":   msg.AppName,
			"process_id": msg.ProcessID,
			"message_id": msg.MessageID,
			"message":    string(msg.Message),
		},
	}, nil
}

func priorityToSeverity(p rfc5424.Priority) logging.Severity {
	switch p {
	case rfc5424.Emergency:
		return logging.Emergency
	case rfc5424.Alert:
		return logging.Alert
	case rfc5424.Crit:
		return logging.Critical
	case rfc5424.Error:
		return logging.Error
	case rfc5424.Warning:
		return logging.Warning
	case rfc5424.Notice:
		return logging.Notice
	case rfc5424.Info:
		return logging.Info
	case rfc5424.Debug:
		return logging.Debug
	default:
		return logging.Default
	}
}
