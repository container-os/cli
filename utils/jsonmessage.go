package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type JSONError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *JSONError) Error() string {
	return e.Message
}

type JSONProgress struct {
	Current int   `json:"current,omitempty"`
	Total   int   `json:"total,omitempty"`
	Start   int64 `json:"start,omitempty"`
}

func (p *JSONProgress) String() string {
	if p.Current == 0 && p.Total == 0 {
		return ""
	}
	current := HumanSize(int64(p.Current))
	if p.Total == 0 {
		return fmt.Sprintf("%8v/?", current)
	}
	total := HumanSize(int64(p.Total))
	percentage := int(float64(p.Current)/float64(p.Total)*100) / 2

	fromStart := time.Now().UTC().Sub(time.Unix(int64(p.Start), 0))
	perEntry := fromStart / time.Duration(p.Current)
	left := time.Duration(p.Total-p.Current) * perEntry
	left = (left / time.Second) * time.Second
	return fmt.Sprintf("[%s>%s] %8v/%v %s", strings.Repeat("=", percentage), strings.Repeat(" ", 50-percentage), current, total, left.String())
}

type JSONMessage struct {
	Status          string        `json:"status,omitempty"`
	Progress        *JSONProgress `json:"progressDetail,omitempty"`
	ProgressMessage string        `json:"progress,omitempty"` //deprecated
	ID              string        `json:"id,omitempty"`
	From            string        `json:"from,omitempty"`
	Time            int64         `json:"time,omitempty"`
	Error           *JSONError    `json:"errorDetail,omitempty"`
	ErrorMessage    string        `json:"error,omitempty"` //deprecated
}

func (jm *JSONMessage) Display(out io.Writer, isTerminal bool) error {
	if jm.Error != nil {
		if jm.Error.Code == 401 {
			return fmt.Errorf("Authentication is required.")
		}
		return jm.Error
	}
	endl := ""
	if isTerminal {
		// <ESC>[2K = erase entire current line
		fmt.Fprintf(out, "%c[2K\r", 27)
		endl = "\r"
	}
	if jm.Time != 0 {
		fmt.Fprintf(out, "[%s] ", time.Unix(jm.Time, 0))
	}
	if jm.ID != "" {
		fmt.Fprintf(out, "%s: ", jm.ID)
	}
	if jm.From != "" {
		fmt.Fprintf(out, "(from %s) ", jm.From)
	}
	if jm.Progress != nil {
		fmt.Fprintf(out, "%s %s%s", jm.Status, jm.Progress.String(), endl)
	} else if jm.ProgressMessage != "" { //deprecated
		fmt.Fprintf(out, "%s %s%s", jm.Status, jm.ProgressMessage, endl)
	} else {
		fmt.Fprintf(out, "%s%s\n", jm.Status, endl)
	}
	return nil
}

func DisplayJSONMessagesStream(in io.Reader, out io.Writer, isTerminal bool) error {
	dec := json.NewDecoder(in)
	ids := make(map[string]int)
	diff := 0
	for {
		jm := JSONMessage{}
		if err := dec.Decode(&jm); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if (jm.Progress != nil || jm.ProgressMessage != "") && jm.ID != "" {
			line, ok := ids[jm.ID]
			if !ok {
				line = len(ids)
				ids[jm.ID] = line
				fmt.Fprintf(out, "\n")
				diff = 0
			} else {
				diff = len(ids) - line
			}
			if isTerminal {
				// <ESC>[{diff}A = move cursor up diff rows
				fmt.Fprintf(out, "%c[%dA", 27, diff)
			}
		}
		err := jm.Display(out, isTerminal)
		if jm.ID != "" {
			if isTerminal {
				// <ESC>[{diff}B = move cursor down diff rows
				fmt.Fprintf(out, "%c[%dB", 27, diff)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
