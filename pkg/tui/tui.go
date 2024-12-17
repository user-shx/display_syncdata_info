package tui

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

func PrintTable(rows [][]string, header bool) {
	// Print the table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	if header {
		addRow(t, rows[0], true)
		border := make([]string, len(rows[0]))
		for i := range border {
			border[i] = strings.Repeat("-", len(rows[0][i]))
		}
		addRow(t, border, false)
		rows = rows[1:]
	}
	for _, row := range rows {
		addRow(t, row, false)
	}

	t.SetStyle(table.Style{
		Name: "tiup",
		Box: table.BoxStyle{
			BottomLeft:       "",
			BottomRight:      "",
			BottomSeparator:  "",
			Left:             "|",
			LeftSeparator:    "|",
			MiddleHorizontal: "-",
			MiddleSeparator:  "  ",
			MiddleVertical:   "  ",
			PaddingLeft:      "",
			PaddingRight:     "",
			Right:            "",
			RightSeparator:   "",
			TopLeft:          "",
			TopRight:         "",
			TopSeparator:     "",
			UnfinishedRow:    "",
		},
		Format: table.FormatOptions{
			Header: text.FormatDefault,
		},
		Options: table.Options{
			SeparateColumns: true,
			// DoNotFillSpaceWhenEndOfLine: true,
		},
	})
	t.Render()
}

func addRow(t table.Writer, rawLine []string, header bool) {
	// Convert []string to []any
	row := make(table.Row, len(rawLine))
	for i, v := range rawLine {
		row[i] = v
	}

	// Add line to the table
	if header {
		t.AppendHeader(row)
	} else {
		t.AppendRow(row)
	}
}
