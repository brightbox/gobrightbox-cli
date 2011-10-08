package main

import (
	"tabwriter"
	"os"
)

func DrawTable(headers []string, data[][]string) {
	out := tabwriter.NewWriter(os.Stdout, 4, 2, 2, ' ', tabwriter.StripEscape)
	out.Write([]uint8(" "))
	for i, header := range headers {
		out.Write([]uint8(header))
		if i < len(headers) {
			out.Write([]uint8("\t"))
		}
	}
	out.Write([]uint8("\n"))
	out.Write([]byte("\t\t\t\t\t\t\t\n"))
	for _, row := range data {
		out.Write([]uint8(" "))
		for i := 0; i < len(headers); i++ {
			if i < len(row) {
				out.Write([]uint8(row[i]))
			}
			out.Write([]uint8("\t"))
		}
		out.Write([]uint8("\n"))
	}
	out.Flush()
}