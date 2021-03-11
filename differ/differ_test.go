package differ

import (
	"strconv"
	"testing"

	"github.com/pjovanovic05/drift/checker"
)

func TestShowDiffType(t *testing.T) {
	if showDiffType(EQUAL) != "=" {
		t.Error("EQUAL conversion failed.")
	}
	if showDiffType(LEFTNEW) != "<" {
		t.Error("LEFTNEW conversion failed.")
	}
	if showDiffType(RIGHTNEW) != ">" {
		t.Error("RIGHTNEW conversion failed.")
	}
	if showDiffType(DIFFERENT) != "x" {
		t.Error("DIFFERENT conversion failed.")
	}
	if showDiffType(33) != "!" {
		t.Error("Catch all conversion failed.")
	}
}

func TestDiff(t *testing.T) {
	// create checker pairs, calculate diff and examine the diff.
	p1 := checker.Pair{Key: "test equal", Value: "equal"}
	p2 := checker.Pair{Key: "test different", Value: "abc"}
	p3 := checker.Pair{Key: "test different", Value: "abcd"}
	p4 := checker.Pair{Key: "test leftnew", Value: "zxcv"}
	x := []checker.Pair{p1, p2, p4}
	y := []checker.Pair{p1, p3}
	dres, err := Diff(x, y)
	if err != nil {
		t.Error(err)
	}
	if len(dres.Diffs) != 3 {
		t.Error("Wrong length of diffs: " + strconv.Itoa(len(dres.Diffs)))
	}

	dl1 := DiffLine{T: EQUAL, Left: p1, Right: p1}
	dl2 := DiffLine{T: DIFFERENT, Left: p2, Right: p3}
	dl3 := DiffLine{T: LEFTNEW, Left: p4}

	expectedDif := DiffResult{
		Left:  "",
		Right: "",
		Diffs: []DiffLine{dl1, dl2, dl3},
	}
	for i := 0; i < len(dres.Diffs); i++ {
		if expectedDif.Diffs[i] != dres.Diffs[i] {
			t.Error("Unexpected diff result.")
		}
	}
}
