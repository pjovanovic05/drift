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
		Funcs(template.FuncMap{"showDiffType": showDiffType, "checkType": checkType}).Parse(reportTemplate))
	err := diffReport.Execute(&outBuff, diffs)
	return outBuff.String(), err
}

var reportTemplate = `
<!DOCTYPE html>
<html lang="en" dir="ltr">
  <head>
    <meta charset="utf-8">
    <title>Diff view</title>
    <link rel="stylesheet"
      href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css"
      integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO"
      crossorigin="anonymous">
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js"
      integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo"
      crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js"
      integrity="sha384-ZMP7rVo3mIykV+2+9J3UJ46jBk0WLaUAdn689aCwoqbBJiSnjAK/l8WvCWPIPm49"
      crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js"
      integrity="sha384-ChfqqxuZUCnJSK3+MXmPNIyE6ZbWh2IMqE241rYiqJxyMiZ6OW/JmZQ5stwEULTy"
      crossorigin="anonymous"></script>
    <style>
      .equal {
        background-color: green;
        color: white;
      }
      .different {
        background-color: yellow;
      }
      .leftnew {
        background-color: blue;
        color: white;
      }
      .rightnew {
        background-color: red;
        color: white;
      }
      .tbar {
        background-color: white;
      }
    </style>
  </head>
  <body>
    <div class="fixed-top tbar">
      Toggle:
      <button class="btn btn-success" type="button" data-toggle="collapse" data-target=".equal">Equal</button>
      <button class="btn btn-warning" type="button" data-toggle="collapse" data-target=".different">Different</button>
      <button class="btn btn-primary" type="button" data-toggle="collapse" data-target=".leftnew">Left New</button>
      <button class="btn btn-danger" type="button" data-toggle="collapse" data-target=".rightnew">Right New</button>
    </div>
    <br/><br/>
    <table class="table table-sm">
      <thead>
        <tr><th>left</th><th>&nbsp;</th><th>right</th></tr>
      </thead>
      <tbody>
        {{range .Diffs}}
        {{if checkType .T "="}}
        <tr class="equal collapse">
        {{else if checkType .T "x"}}
        <tr class="different collapse">
        {{else if checkType .T "<"}}
        <tr class="leftnew collapse">
        {{else if checkType .T ">"}}
        <tr class="rightnew collapse">
        {{end}}
          <td>{{.Left.Key}}</td>
          <td>{{.T | showDiffType}}</td>
          <td>{{.Right.Key}}</td>
        {{end}}
      </tbody>
    </table>
  </body>
</html>
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

func checkType(t DiffType, s string) bool {
	return showDiffType(t) == s
}
