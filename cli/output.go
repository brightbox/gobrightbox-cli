package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

func tabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
}

func tabWriterRight() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', tabwriter.AlignRight)
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}

func drawShow(w io.Writer, d map[string]interface{}, order []string) {
	for _, key := range order {
		fmt.Fprint(w, key, ": \t", d[key])
		w.Write([]byte{'\n'})
	}
}

func PrettyPrintJson(body []byte) *bytes.Buffer {
	var out bytes.Buffer
	json.Indent(&out, body, "", "\t")
	return &out
}
