package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
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

func drawShow(w io.Writer, d []interface{}) {
	for i := 0; i+1 < len(d); i += 2 {
		fmt.Fprint(w, d[i], ": \t", d[i+1])
		w.Write([]byte{'\n'})
	}
}

func PrettyPrintJson(body []byte) *bytes.Buffer {
	var out bytes.Buffer
	json.Indent(&out, body, "", "\t")
	return &out
}

func collectById(resources interface{}) string {
	return collectByField(resources, "Id")
}

func collectByField(resources interface{}, name string) string {
	val := reflect.ValueOf(resources)
	if val.Kind() != reflect.Slice {
		return ""
	}
	var ids = make([]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		rval := val.Index(i)
		if rval.Kind() != reflect.Struct {
			continue
		}
		if !rval.FieldByName(name).IsValid() {
			continue
		}
		ids[i] = rval.FieldByName(name).String()
	}
	return strings.Join(ids, ",")
}
