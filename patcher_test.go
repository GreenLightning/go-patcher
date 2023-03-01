package patcher

import (
	"fmt"
	"strings"
	"testing"
)

func TestPassthrough(t *testing.T) {
	input := "abc"
	var p Patcher

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abc")
}

func TestEmpty(t *testing.T) {
	var p Patcher

	if !p.Empty() {
		t.Fatal("expected empty patcher")
	}

	p.Delete(1, 1)

	if p.Empty() {
		t.Fatal("expected non-empty patcher")
	}

	p.Reset()

	if !p.Empty() {
		t.Fatal("expected empty patcher after reset")
	}
}

func TestEmptyZeroLength(t *testing.T) {
	var p Patcher

	p.InsertString(1, "")
	p.RewriteString(2, 0, "")

	if !p.Empty() {
		t.Fatal("expected empty patcher")
	}
}

func TestDelete(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(1, 2)

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "ade")
}

func TestDeleteNegativeOffset(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(-1, 2)

	output, err := p.PatchString(input)

	checkError(t, output, err, "negative offset", checkPatch(-1, 2, ""))
}

func TestDeleteNegativeLength(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(2, -1)

	output, err := p.PatchString(input)

	checkError(t, output, err, "negative length", checkPatch(2, -1, ""))
}

func TestDeleteNothingAtEnd(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(5, 0)

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abcde")
}

func TestDeleteOutOfRange(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(5, 1)

	output, err := p.PatchString(input)

	checkError(t, output, err, "out of range", checkPatch(5, 1, ""))
}

func TestDeleteTouching(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(1, 2)
	p.Delete(3, 1)

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "ae")
}

func TestDeleteOutOfOrder(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(3, 1)
	p.Delete(1, 1)

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "ace")
}

func TestDeleteConflict(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.Delete(1, 3)
	p.Delete(2, 1)

	output, err := p.PatchString(input)

	checkError(t, output, err, "conflict", checkPatch(2, 1, ""), checkPatch(1, 3, ""))
}

func TestInsert(t *testing.T) {
	input := "ae"
	var p Patcher
	p.InsertString(1, "bcd")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abcde")
}

func TestInsertNegativeOffset(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.InsertString(-1, "z")

	output, err := p.PatchString(input)

	checkError(t, output, err, "negative offset", checkPatch(-1, 0, "z"))
}

func TestInsertAtEnd(t *testing.T) {
	input := "abcd"
	var p Patcher
	p.InsertString(4, "e")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abcde")
}

func TestInsertOutOfRange(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.InsertString(10, "z")

	output, err := p.PatchString(input)

	checkError(t, output, err, "out of range", checkPatch(10, 0, "z"))
}

func TestInsertMultiple(t *testing.T) {
	input := "ae"
	var p Patcher
	p.InsertString(1, "b")
	p.InsertString(1, "c")
	p.InsertString(1, "d")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abcde")
}

func TestInsertOutOfOrder(t *testing.T) {
	input := "ace"
	var p Patcher
	p.InsertString(2, "d")
	p.InsertString(1, "b")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abcde")
}

func TestRewrite(t *testing.T) {
	input := "abc"
	var p Patcher
	p.RewriteString(1, 1, "xyz")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "axyzc")
}

func TestRewriteInsert(t *testing.T) {
	input := "ac"
	var p Patcher
	p.RewriteString(1, 0, "b")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abc")
}

func TestRewriteDelete(t *testing.T) {
	input := "abc"
	var p Patcher
	p.RewriteString(1, 1, "")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "ac")
}

func TestRewriteTouching(t *testing.T) {
	input := "abc"
	var p Patcher
	p.RewriteString(0, 2, "x")
	p.RewriteString(2, 1, "yz")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "xyz")
}

func TestRewriteConflict(t *testing.T) {
	input := "abc"
	var p Patcher
	p.RewriteString(0, 2, "x")
	p.RewriteString(1, 1, "yz")

	output, err := p.PatchString(input)

	checkError(t, output, err, "conflict", checkPatch(0, 2, "x"), checkPatch(1, 1, "yz"))
}

func TestReset(t *testing.T) {
	input := "abc"
	var p Patcher
	p.RewriteString(1, 1, "xyz")
	p.Reset()
	p.InsertString(3, "z")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "abcz")
}

func TestCombined(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.InsertString(3, "k")
	p.Delete(1, 1)
	p.InsertString(3, "l")
	p.RewriteString(3, 1, "xyz")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "acklxyze")
}

func TestCombinedConflict(t *testing.T) {
	input := "abcde"
	var p Patcher
	p.InsertString(0, "g")
	p.InsertString(4, "f")
	p.Delete(1, 1)
	p.RewriteString(2, 1, "uvw")
	p.RewriteString(3, 2, "xyz")

	output, err := p.PatchString(input)

	checkError(t, output, err, "conflict", checkPatch(4, 0, "f"), checkPatch(3, 2, "xyz"))
}

func TestExample(t *testing.T) {
	input := "The brown fox jumps twice over the lazy horse"

	var p Patcher
	p.InsertString(3, " quick")
	p.Delete(strings.Index(input, "twice "), len("twice "))
	p.RewriteString(40, 5, "dog")

	actual, err := p.PatchString(input)

	checkOutput(t, actual, err, "The quick brown fox jumps over the lazy dog")
}

func checkOutput(t *testing.T, actual string, err error, expected string) {
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if actual != expected {
		t.Errorf("wrong result:\nexpected: %s\nactual:   %s", expected, actual)
	}
}

func checkError(t *testing.T, output string, err error, phrase string, patches ...string) {
	if err == nil {
		t.Fatalf("missing error (output: %s)", output)
	}
	if !strings.Contains(err.Error(), phrase) {
		t.Fatal("wrong error:", err)
	}
	for _, patch := range patches {
		if !strings.Contains(err.Error(), patch) {
			t.Fatal("wrong error:", err)
		}
	}
}

func checkPatch(offset, length int, data string) string {
	return fmt.Sprintf("(%d,%d,%s)", offset, length, data)
}
