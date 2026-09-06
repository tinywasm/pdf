package fpdf

import (
	"bytes"
	"io"
	"os"
	"sort"

	. "webtyp.com/fmt"
)

type sortType struct {
	length int
	less   func(int, int) bool
	swap   func(int, int)
}

func (s *sortType) Len() int {
	return s.length
}

func (s *sortType) Less(i, j int) bool {
	return s.less(i, j)
}

func (s *sortType) Swap(i, j int) {
	s.swap(i, j)
}

func gensort(Len int, Less func(int, int) bool, Swap func(int, int)) {
	sort.Sort(&sortType{length: Len, less: Less, swap: Swap})
}

func writeBytes(leadStr string, startPos int, sl []byte) {
	var pos, max int
	var b byte
	out := Sprintf("%s %07x", leadStr, startPos)
	max = len(sl)
	for pos < max {
		out += " "
		for k := 0; k < 8; k++ {
			if pos < max {
				out += Sprintf(" %02x", sl[pos])
			} else {
				out += "   "
			}
			pos++
		}
	}
	out += "  |"
	pos = 0
	for pos < max {
		b = sl[pos]
		if b < 32 || b >= 128 {
			b = '.'
		}
		out += Sprintf("%c", b)
		pos++
	}
	out += "|"
	println(out)
}

func checkBytes(pos int, sl1, sl2 []byte, printDiff bool) (eq bool) {
	eq = bytes.Equal(sl1, sl2)
	if !eq && printDiff {
		writeBytes("<", pos, sl1)
		writeBytes(">", pos, sl2)
	}
	return
}

// CompareBytes compares the bytes referred to by sl1 with those referred to by
// sl2. Nil is returned if the buffers are equal, otherwise an error.
func CompareBytes(sl1, sl2 []byte, printDiff bool) (err error) {
	var posStart, posEnd, len1, len2, length int
	var diffs bool

	len1 = len(sl1)
	len2 = len(sl2)

	// Check if files have different sizes
	if len1 != len2 {
		diffs = true
	}

	length = len1
	if length > len2 {
		length = len2
	}
	for posStart < length {
		posEnd = posStart + 16
		if posEnd > length {
			posEnd = length
		}
		if !checkBytes(posStart, sl1[posStart:posEnd], sl2[posStart:posEnd], printDiff) {
			diffs = true
		}
		posStart = posEnd
	}
	if diffs {
		err = Err("documents are different")
	}
	return
}

// ComparePDFs reads and compares the full contents of the two specified
// readers byte-for-byte. Nil is returned if the buffers are equal, otherwise
// an error.
func ComparePDFs(rdr1, rdr2 io.Reader, printDiff bool) (err error) {
	b1 := bytes.NewBuffer(nil)
	b2 := bytes.NewBuffer(nil)
	_, err = b1.ReadFrom(rdr1)
	if err == nil {
		_, err = b2.ReadFrom(rdr2)
		if err == nil {
			err = CompareBytes(b1.Bytes(), b2.Bytes(), printDiff)
		}
	}
	return
}

// ComparePDFFiles reads and compares the full contents of the two specified
// files byte-for-byte. Nil is returned if the file contents are equal, or if
// the second file is missing, otherwise an error.
func ComparePDFFiles(file1Str, file2Str string, printDiff bool) (err error) {
	var sl1, sl2 []byte
	sl1, err = os.ReadFile(file1Str)
	if err == nil {
		sl2, err = os.ReadFile(file2Str)
		if err == nil {
			err = CompareBytes(sl1, sl2, printDiff)
		} else {
			// Second file is missing; treat this as success
			err = nil
		}
	}
	return
}
