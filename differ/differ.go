package differ

import (
	"bytes"
	"drift/checker"
	"html/template"
)

type DiffType int

const (
	EQUAL DiffType = iota
	LEFTNEW
	RIGHTNEW
	DIFFERENT
)

type DiffLine struct {
	T     DiffType
	Left  checker.Pair
	Right checker.Pair
}

type DiffResult struct {
	Left  string
	Right string
	Diffs []DiffLine
}

// Diff checks differences between two slices of key-value pairs.
func Diff(x, y []checker.Pair) (dr DiffResult, err error) {
	var i, j int
	var diffs []DiffLine
	xn, yn := len(x), len(y)
	for i < xn && j < yn {
		if x[i].Key == y[j].Key {
			if x[i].Value == y[j].Value {
				diffs = append(diffs, DiffLine{T: EQUAL, Left: x[i], Right: y[j]})
			} else {
				diffs = append(diffs, DiffLine{T: DIFFERENT, Left: x[i], Right: y[j]})
			}
			i++
			j++
		} else {
			if x[i].Key < y[j].Key {
				diffs = append(diffs, DiffLine{T: LEFTNEW, Left: x[i]})
				i++
			} else {
				diffs = append(diffs, DiffLine{T: RIGHTNEW, Right: y[j]})
				j++
			}
		}
	}
	// drain longer list
	for ; i < xn; i++ {
		diffs = append(diffs, DiffLine{T: LEFTNEW, Left: x[i]})
	}
	for ; j < yn; j++ {
		diffs = append(diffs, DiffLine{T: RIGHTNEW, Right: y[j]})
	}
	dr = DiffResult{Diffs: diffs}
	return dr, err
}

func GetHtmlReport(diffs DiffResult) (string, error) {
	var outBuff bytes.Buffer
	var diffReport = template.Must(template.New("diffreport").
		Funcs(template.FuncMap{"showDiffType": showDiffType}).Parse(reportTemplate))
	err := diffReport.Execute(&outBuff, diffs)
	return outBuff.String(), err
}

var reportTemplate = `
<h1>diff report</h1>
<table>
	<tr><th>{{.Left}}</th><th>&nbsp;</th><th>{{.Right}}</th></tr>
	{{range .Diffs}}
	<tr>
		<td>{{.Left.Key}}</td>
		<td>{{.T | showDiffType}}</td>
		<td>{{.Right.Key}}</td>
	</tr>
	{{end}}
</table>
`

func showDiffType(t DiffType) string {
	switch t {
	case EQUAL:
		return "="
	case LEFTNEW:
		return "<"
	case RIGHTNEW:
		return ">"
	case DIFFERENT:
		return "x"
	}
	return "!"
}
