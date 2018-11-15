package differ

import (
	"drift/checker"
	"html/template"
	"log"
	"os"
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
func Diff(x, y []checker.Pair) (diffs []DiffLine, err error) {
	var i, j int
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
				diffs = append(diffs, DiffLine{T: RIGHTNEW, Right: y[i]})
				j++
			}
		}
	}
	// drain longer list
	for ; i < xn; i++ {
		diffs = append(diffs, DiffLine{T: LEFTNEW, Left: x[i]})
	}
	for ; j < yn; j++ {
		diffs = append(diffs, DiffLine{T: RIGHTNEW, Right: y[i]})
	}
	return diffs, err
}

func SaveDiff(diffs []DiffLine) error {
	return nil
}

func LoadDiffs(location string) ([]DiffLine, error) {

}

func SaveDiffsForVim(loc string, x, y []checker.Pair) error {

}

func SaveHTMLReport(location string, diffs DiffResult) error {
	// TODO make template
	var diffReport = template.Must(template.New("diffreport").
		Funcs(template.FuncMap{"showDiffType": showDiffType}).Parse(reportTemplate))
	// TODO populate template
	out, err := os.Create(location)
	if err != nil {
		log.Fatal(err)
	}
	err = diffReport.Execute(out, diffs)
	// TODO write to file
	return err
}

var reportTemplate = `
<h1>diff report</h1>
<table>
	<tr><th>{{.Left}}</th><th>&nbsp;</th><th>{{.Right}}</th></tr>
	{{range .diffs}}
	<tr>
		<td>{{.Left.Value}}</td>
		<td>{{.T | showDiffType}}</td>
		<td>{{.Right.Value}}</td>
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
