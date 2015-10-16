package cli

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"
)

type FieldOutput struct {
	Writer *tabwriter.Writer
}

type RowFieldOutput struct {
	FieldOutput
	FieldOrder []string
}

type ShowFieldOutput struct {
	FieldOutput
	FieldOrder []string
}

func (fo *FieldOutput) Flush() {
	fo.Writer.Flush()
}

func (fo *RowFieldOutput) Setup(fieldorder []string) {
	fo.Writer = tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	fo.FieldOrder = fieldorder
}

func (fo *ShowFieldOutput) Setup(fieldorder []string) {
	fo.Writer = tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', tabwriter.AlignRight)
	fo.FieldOrder = fieldorder
}

func (fo *RowFieldOutput) SendHeader() {
	for i, f := range fo.FieldOrder {
		fmt.Fprint(fo.Writer, strings.ToUpper(f))
		if i+1 < len(fo.FieldOrder) {
			fo.Writer.Write([]byte{'\t'})
		} else {
			fo.Writer.Write([]byte{'\n'})
		}
	}
}
func (fo *RowFieldOutput) Write(fields map[string]string) error {
	for i, f := range fo.FieldOrder {
		v, ok := fields[f]
		if ok == false {
			return fmt.Errorf("No field named '%s' available for display", f)
		}
		fmt.Fprint(fo.Writer, v)
		if i+1 < len(fo.FieldOrder) {
			fo.Writer.Write([]byte{'\t'})
		} else {
			fo.Writer.Write([]byte{'\n'})
		}
	}
	return nil
}
func (fo *ShowFieldOutput) Write(fields map[string]string) error {
	var order = fo.FieldOrder
	if len(order) == 1 && order[0] == "all" {
		order = []string{}
		for k := range fields {
			order = append(order, k)
		}
	}
	for _, f := range order {
		v, ok := fields[f]
		if ok == false {
			return fmt.Errorf("No field named '%s' available for display", f)
		}
		fmt.Fprint(fo.Writer, f, ": \t", v)
		fo.Writer.Write([]byte{'\n'})
	}
	return nil
}

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

func formatTime(t *time.Time) string {
	if t != nil {
		return t.String()
	} else {
		return ""
	}
}
func formatBool(t bool) string {
	return fmt.Sprintf("%v", t)
}
func formatInt(t int) string {
	return fmt.Sprintf("%d", t)
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
