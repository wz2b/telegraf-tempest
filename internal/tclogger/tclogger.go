package tclogger

import (
	"bytes"
	"fmt"
	lineProtocol "github.com/influxdata/line-protocol"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type TelegrafCompatibleLogger struct {
	OutputAsComment     bool
	OutputAsMeasurement bool
	MeasurementName     string

	Writer     io.Writer
	lineBuffer string
}

func (w *TelegrafCompatibleLogger) Write(p []byte) (int, error) {

	count := 0
	if w.OutputAsComment {
		count += w.formatComment(p)
	}

	if w.OutputAsMeasurement {
		c, err := w.formatMeasurement(p)
		if err != nil {
			count += c
		}
	}

	return count, nil
}

func (w *TelegrafCompatibleLogger) formatComment(bytesToWrite []byte) int {
	count := 0

	stringToWrite := string(bytesToWrite)

	/*
	 * If there is any buffer left over from last go-around that gets processed first
	 */
	if len(w.lineBuffer) > 0 {
		stringToWrite = w.lineBuffer + stringToWrite
		w.lineBuffer = ""
	}

	lastNewlineIndex := strings.LastIndexByte(stringToWrite, '\n')
	if lastNewlineIndex != len(stringToWrite)-1 {
		/*
		 * The last line is not complete
		 */
		w.lineBuffer = stringToWrite[lastNewlineIndex+1:]
		stringToWrite = stringToWrite[:lastNewlineIndex]
	}

	/*
	 * Check whether or not the last line is
	 */

	lines := strings.Split(stringToWrite, "\n")

	for _, line := range lines {
		/* Throw away any ampty lines */
		if len(line) > 0 {
			c1, _ := fmt.Fprint(w.Writer, "# ")
			c2, _ := fmt.Fprintln(w.Writer, line)
			count += c1 + c2
		}
	}

	return count
}

func (w *TelegrafCompatibleLogger) formatMeasurement(bytesToWrite []byte) (int, error) {
	stringToWrite := string(bytesToWrite)

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	message := strings.ReplaceAll(stringToWrite, "\n", " ")

	tags["application"] = "vue2"
	fields["log"] = message

	metric, err := lineProtocol.New("log", tags, fields, time.Now())
	if err != nil {
		fmt.Fprintf(w.Writer, "# Error encoding log to metric: %s\n", err)
		return 0, err
	}

	buf := &bytes.Buffer{}
	serializer := lineProtocol.NewEncoder(buf)
	serializer.SetMaxLineBytes(4096)
	serializer.SetFieldTypeSupport(lineProtocol.UintSupport)
	serializer.Encode(metric)

	strbuf := buf.String()
	if len(strbuf) > 0 {
		fmt.Fprintln(w.Writer, strbuf)
		return len(strbuf), nil
	} else {
		return 0, nil
	}
}

/*
 * Create a new instance with default settings
 */
func Create() *TelegrafCompatibleLogger {
	return &TelegrafCompatibleLogger{
		OutputAsComment:     true,
		OutputAsMeasurement: false,
		MeasurementName:     "log",
		Writer: os.Stderr,
	}
}

/*
 * Install this tclogger as the 'writer' for the default logger
 */
func (l *TelegrafCompatibleLogger) Install() *TelegrafCompatibleLogger {
	log.SetOutput(l)
	return l
}

/*
 * Start this teclogger, which is the same as calling Install() except
 * it also announces itself
 */
func (l *TelegrafCompatibleLogger) Start() *TelegrafCompatibleLogger {
	l.Install()
	log.Printf("Telegraf compatible logger started at %s\n",
		time.Now().Format(time.RFC822))
	return l
}