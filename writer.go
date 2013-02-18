package ar

import (
	"io"
	"time"
	"strconv"
)

const HEADER_BYTE_SIZE = 68 

type slicer []byte

func (sp *slicer) next(n int) (b []byte) {
	s := *sp
	b, *sp = s[0:n], s[n:]
	return
}

// Writer provides sequential writing of an ar archive.
// An ar archive is sequence of header file pairs
// Call WriteHeader to begin writing a new file, then call Write to supply the file's data
//
// Example:
// archive := ar.NewWriter(writer)
// header := new(ar.Header)
// header.Size = 15 // bytes
// if err := archive.WriteHeader(header); err != nil {
// 	return err
// }
// io.Copy(archive, data)
type Writer struct {
	w io.Writer
	globalHeader bool
}

type Header struct {
	Name string
	ModTime time.Time
	Uid int
	Gid int
	Mode int64
	Size int64
}

// Create a new ar writer that writes to w
func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }

func (aw *Writer) numeric(b []byte, x int64) {
	s := strconv.FormatInt(x, 10)
	for len(s) < len(b) {
		s = s + " "
	}
	copy(b, []byte(s))
}

func (aw *Writer) octal(b []byte, x int64) {
	s := "100" + strconv.FormatInt(x, 8)
	for len(s) < len(b) {
		s = s + " "
	}
	copy(b, []byte(s))
}

func (aw *Writer) string(b []byte, str string) {
	s := str
	for len(s) < len(b) {
		s = s + " "
	}
	copy(b, []byte(s))
}

// Writes to the current entry in the ar archive
// Returns ErrWriteTooLong if more than header.Size
// bytes are written after a call to WriteHeader
func (aw *Writer) Write(b []byte) (n int, err error) {
	n, err = aw.w.Write(b)
	if len(b)%2 == 1 {
		n2, _ := aw.w.Write([]byte{'\n'})
		return n+n2, err
	}

	return
}

// Writes the header to the underlying writer and prepares
// to receive the file payload
func (aw *Writer) WriteHeader(hdr *Header) error {
	header := make([]byte, HEADER_BYTE_SIZE)
	s := slicer(header)

	if !aw.globalHeader {
		aw.string(s.next(8), "!<arch>\n")
		aw.globalHeader = true
	}

	aw.string(s.next(16), hdr.Name)
	aw.numeric(s.next(12), hdr.ModTime.Unix())
	aw.numeric(s.next(6), int64(hdr.Uid))
	aw.numeric(s.next(6), int64(hdr.Gid))
	aw.octal(s.next(8), hdr.Mode)
	aw.numeric(s.next(10), hdr.Size)
	aw.string(s.next(2), "`\n")

	_, err := aw.w.Write(header)

	return err
}