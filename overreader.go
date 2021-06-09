package overreader

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
)

type Range struct {
	Content io.ReadSeeker
	Offset  int64
	length  int64
}

func (r *Range) setLength() error {
	size, err := r.Content.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	_, err = r.Content.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	r.length = size
	return nil
}

type rangeSlice []*Range

func (s rangeSlice) Len() int           { return len(s) }
func (s rangeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s rangeSlice) Less(i, j int) bool { return s[i].Offset < s[j].Offset }

// Valid returns no error iff none of the content in the slice overlaps and the slice is sorted
func (s rangeSlice) valid() error {
	if len(s) == 0 {
		return nil
	}

	if !sort.IsSorted(s) {
		return fmt.Errorf("Cannot check the validity of an unsorted rangeSlice")
	}

	// check all but the last range for overlap with the following range
	for i := 0; i < len(s)-1; i++ {
		if s[i].Offset+s[i].length > s[i+1].Offset {
			return fmt.Errorf("Range list is invalid: range with offset %d and length %d overlaps with range at offset %d", s[i].Offset, s[i].length, s[i+1].Offset)
		}
	}

	return nil
}

type skipReader struct {
	reader io.Reader
}

func newSkipReader(reader io.Reader) *skipReader {
	return &skipReader{reader: reader}
}

var _ io.Reader = &skipReader{}

func (sr *skipReader) Read(p []byte) (int, error) {
	_, err := io.Copy(ioutil.Discard, sr.reader)
	if err == nil {
		err = io.EOF
	}

	return 0, err
}

func NewReader(baseReader io.Reader, overrideRanges ...*Range) (io.Reader, error) {
	ranges := rangeSlice(overrideRanges)
	for _, r := range ranges {
		if err := r.setLength(); err != nil {
			return nil, err
		}
	}

	sort.Sort(ranges)
	if err := ranges.valid(); err != nil {
		return nil, err
	}

	var loc int64
	readers := make([]io.Reader, 0)
	for _, r := range ranges {
		limit := r.Offset - loc
		readers = append(readers, io.LimitReader(baseReader, limit), r.Content, newSkipReader(io.LimitReader(baseReader, r.length)))
		loc += limit + r.length
	}
	readers = append(readers, baseReader)

	return io.MultiReader(readers...), nil
}
