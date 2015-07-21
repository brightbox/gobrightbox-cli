package cli

import (
	"bytes"
	"os"
	"io"
	"fmt"
	"text/tabwriter"
	"encoding/json"
)


func tabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
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

func PrettyPrintJson(body []byte) *bytes.Buffer {
	var out bytes.Buffer
	json.Indent(&out, body, "", "\t")
	return &out
}
