package storagefile

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	MAXREADERS = 5
)

type Header struct {
	Id         int64     // insert position
	Type       string    // type, path, key...
	Date       time.Time // insert datetime
	SkipSync   bool      // future feature skip sync
	DataLength int32     // actual data length
	Data       []byte    // data, value ...
	//	Guid       string
}

type reader struct {
	file *os.File
	idx  *os.File
}

func (r *reader) close() {
	r.file.Close()
	r.idx.Close()
}

type StorageFile struct {
	file          *os.File
	idx           *os.File
	filename      string
	lastptr       int64
	writer        *bufio.Writer
	idxwriter     *bufio.Writer
	idxrdr        *os.File
	datrdr        *os.File
	count         int64
	readers       chan *reader
	dirty         bool
	FlushOnWrites bool // Lower performance but better data integrity
	sync.Mutex
}

// Add terminator sequence to a failed integrity check data file at your own risk for recovery
// last data in the file will be invalid
func AddTerminator(filename string) {
	// FEATURE : write AddTerminator()

	// f, e := os.OpenFile(filename, os.O_WRONLY, 0644)
	// if e != nil {
	// 	return
	// }
	// defer f.Close()

}

// Open/create a stroage file (single writer/ multiple reader)
func Open(filename string) (*StorageFile, error) {
	sf := StorageFile{}
	sf.filename = filename

	if !fileExists(filename) {
		os.Remove(filename + ".idx")
		os.Remove(filename + ".dirty")
	}

	// integrity check
	if fileExists(filename + ".dirty") {
		e := sf.rebuildIndex()
		if e != nil {
			return nil, e
		}
	} else if fileExists(filename) {
		// check last 4 bytes of dat file == '||||' -> rebuild??
		if !sf.checkLast() {
			return nil, errors.New("data integrity error in file, last record does not have a terminator sequence (use AddTerminator() at your own risk)")
		}
	}

	var e error
	sf.file, e = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if e != nil {
		return nil, e
	}
	sf.idx, e = os.OpenFile(filename+".idx", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if e != nil {
		return nil, e
	}

	sf.writer = bufio.NewWriter(sf.file)
	sf.idxwriter = bufio.NewWriter(sf.idx)
	sf.idxrdr, e = os.OpenFile(filename+".idx", os.O_RDONLY, 0644)
	if e != nil {
		return nil, e
	}
	sf.datrdr, e = os.OpenFile(filename, os.O_RDONLY, 0644)
	if e != nil {
		return nil, e
	}
	// set sf.last
	fi, _ := sf.file.Stat()
	sf.lastptr = fi.Size()
	fi, _ = sf.idx.Stat()
	sf.count = fi.Size() / 8

	os.WriteFile(filename+".dirty", []byte("isdirty"), 0644)

	sf.readers = make(chan *reader, MAXREADERS)
	for i := 0; i < MAXREADERS; i++ {
		sf.readers <- makeReader(sf.filename)
	}

	return &sf, nil
}

func (sf *StorageFile) flush() {
	if sf.dirty {
		sf.dirty = false
		sf.writer.Flush()
		sf.idxwriter.Flush()
	}
}

func makeReader(filename string) *reader {
	r := reader{}
	var e error
	r.file, e = os.OpenFile(filename, os.O_RDONLY, 0644)
	if e != nil {
		return nil
	}
	r.idx, e = os.OpenFile(filename+".idx", os.O_RDONLY, 0644)
	if e != nil {
		r.file.Close()
		return nil
	}

	return &r
}

func (sf *StorageFile) checkLast() bool {
	f, _ := os.OpenFile(sf.filename, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	_, e := f.Seek(-4, os.SEEK_END)
	if e != nil {
		return true
	}
	b := make([]byte, 4)
	f.Read(b)
	if b[0] != '|' || b[1] != '|' || b[2] != '|' || b[3] != '|' {
		fmt.Println("terminator error")
		return false
	}
	return true
}

func fileExists(fn string) bool {
	_, e := os.Stat(fn)
	return e == nil
}

// Save data to storage file
func (sf *StorageFile) Save(dtype string, data []byte) int64 {
	sf.Lock()

	i := sf.internalSave(dtype, data, false)

	sf.Unlock()
	return i
}

func (sf *StorageFile) internalSave(dtype string, data []byte, skip bool) int64 {

	sf.dirty = true
	i := sf.count
	atomic.AddInt64(&sf.count, 1)
	len := sf.saveHeader(dtype, data, skip)
	sf.writer.Write(data)
	sf.writer.Write([]byte("||||")) // terminator for easy health checking
	binary.Write(sf.idxwriter, binary.LittleEndian, sf.lastptr)

	atomic.AddInt64(&sf.lastptr, len+4)

	// slows writes 2x, config for tradeoff
	if sf.FlushOnWrites {
		sf.flush()
	}

	return i
}

func now() time.Time {
	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)
	// the following compiles on linux/windows/pi
	return time.Unix(0, tv.Nano())
	// return time.Unix(0, syscall.TimevalToNsec(tv)) // syscall not available on windows/pi
}

func (sf *StorageFile) saveHeader(dtype string, data []byte, skipsync bool) int64 {

	hdr := new(bytes.Buffer)
	// 'ITEM' 4 bytes   identifier for rebuild if needed :0-3
	hdr.WriteString("{{{{")
	// exp    2 bytes   expansion bytes :4-5
	hdr.WriteByte(0)
	hdr.WriteByte(0)
	// datetime 4 bytes :6+15
	b, _ := now().MarshalBinary() //binary.Write(hdr, binary.LittleEndian, time.Now())
	hdr.Write(b)
	// skipsync 1 byte :21
	if skipsync {
		hdr.WriteByte(1)
	} else {
		hdr.WriteByte(0)
	}
	// dtype len 2 bytes :22-23
	binary.Write(hdr, binary.LittleEndian, int16(len(dtype)))
	// data len 4 bytes :24-27
	binary.Write(hdr, binary.LittleEndian, int32(len(data)))
	// id len 8 bytes :28-35
	binary.Write(hdr, binary.LittleEndian, sf.count)
	// dtype string 36+len
	hdr.WriteString(dtype)

	// write to file
	sf.writer.Write(hdr.Bytes())

	return int64(hdr.Len() + len(data))
}

// Get Header for the index in stroage file starts at 1
func (sf *StorageFile) GetHeader(id int64) (*Header, error) {
	sf.flush()

	if id > sf.count || id <= 0 {
		return nil, errors.New("id not available")
	}

	rdr := <-sf.readers

	h, e := rdr.getheader(id - 1)

	sf.readers <- rdr
	return h, e
}

func (r *reader) getheader(id int64) (*Header, error) {

	d := Header{}
	r.idx.Seek(id*8, io.SeekStart)
	var ptr int64
	binary.Read(r.idx, binary.LittleEndian, &ptr)

	r.file.Seek(ptr, io.SeekStart)
	hdrlen := 36
	// read the rest
	buf := make([]byte, 4096)
	n, e := io.ReadFull(r.file, buf)
	if e != nil && n == 0 {
		return nil, e
	}
	if n < hdrlen {
		return nil, errors.New("header byte count error")
	}
	if buf[0] != '{' || buf[1] != '{' || buf[2] != '{' || buf[3] != '{' || buf[4] != 0 || buf[5] != 0 {
		return nil, errors.New("header prefix invalid")
	}
	// read datetime 6+15
	d.Date.UnmarshalBinary(buf[6:21])

	d.SkipSync = false
	if buf[21] == 1 {
		d.SkipSync = true
	}
	dtlen := int16(binary.LittleEndian.Uint16(buf[22:]))
	d.DataLength = int32(binary.LittleEndian.Uint32(buf[24:]))
	d.Id = int64(binary.LittleEndian.Uint64(buf[28:]))

	if hdrlen+int(dtlen)+int(d.DataLength) > 4096 {
		// reread data
		r.file.Seek(ptr, io.SeekStart)
		buf = make([]byte, hdrlen+int(dtlen)+int(d.DataLength))
		n, e = io.ReadFull(r.file, buf)
		if e != nil {
			return nil, e
		}
		if n != hdrlen+int(dtlen)+int(d.DataLength) {
			return nil, errors.New("unable to read data")
		}
	}
	dtlen += int16(hdrlen)

	d.Type = string(buf[hdrlen:dtlen])

	n = int(dtlen) + int(d.DataLength)

	d.Data = buf[dtlen:n]

	return &d, nil
}

// Get the "type" and data bytes for id starts at 1
func (sf *StorageFile) Get(id int64) (string, []byte, error) {

	h, e := sf.GetHeader(id)
	if e != nil {
		return "", nil, e
	}
	return h.Type, h.Data, nil
}

// Get the "type" and "string" values for id starts at 1
func (sf *StorageFile) GetString(id int64) (string, string, error) {

	h, e := sf.GetHeader(id)
	if e != nil {
		return "", "", e
	}
	return h.Type, string(h.Data), nil
}

// Close storage file
func (sf *StorageFile) Close() {

	sf.flush()

	close(sf.readers)

	for r := range sf.readers {
		r.close()
	}

	sf.idxrdr.Close()
	sf.datrdr.Close()
	sf.file.Close()
	sf.idx.Close()
	os.Remove(sf.filename + ".dirty")
}

// Count of items in storage file
func (sf *StorageFile) Count() int64 {
	return sf.count
}

// Iterate over data in strorage file returns a chan of Header.
// * if you don't iterate to the end, close the channel
func (sf *StorageFile) Iterate() chan *Header {
	ch := make(chan *Header)
	index := int64(1)
	go func() {
		defer close(ch)
		for index <= sf.count {
			h, e := sf.GetHeader(index)
			if e != nil {
				fmt.Println("iterate failed", e)
				return
			}
			ch <- h
			index++
		}
	}()
	return ch
}

func (sf *StorageFile) rebuildIndex() error {
	// go through .dat file and rebuild .idx file
	fmt.Println("Rebuilding...")
	os.Remove(sf.filename + ".idx")

	sf.file, _ = os.OpenFile(sf.filename, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0644)
	sf.idx, _ = os.OpenFile(sf.filename+".idx", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	idxwriter := bufio.NewWriter(sf.idx)
	datrdr := bufio.NewReader(sf.file)
	buf := make([]byte, 36)
	var ptr int64 = 0
	var count int64 = 0

	for {

		// read header bytes
		n, e := io.ReadFull(datrdr, buf)

		if e != nil {
			fmt.Println(e)
			break
		}
		if n == 0 {
			break
		}
		if n < 36 {
			fmt.Println("not enough bytes for header @count=", count)
		}
		// check header -> err
		if buf[0] != '{' || buf[1] != '{' || buf[2] != '{' || buf[3] != '{' || buf[4] != 0 || buf[5] != 0 {
			fmt.Println("header error")
			break
		}
		dtlen := int16(binary.LittleEndian.Uint16(buf[22:]))
		datalen := int32(binary.LittleEndian.Uint32(buf[24:]))
		b := make([]byte, dtlen+int16(datalen))
		n, _ = io.ReadFull(datrdr, b)
		if n == 0 {
			fmt.Println("zero bytes @count=", count)
			break
		}
		if n < int(dtlen)+int(datalen) {
			fmt.Println("not enough bytes @count=", count)
			break
		}
		b = make([]byte, 4)
		io.ReadFull(datrdr, b)
		if b[0] != '|' || b[1] != '|' || b[2] != '|' || b[3] != '|' {
			fmt.Println("terminator error @count=", count)
			break
		}

		binary.Write(idxwriter, binary.LittleEndian, ptr)

		ptr = ptr + 36 + int64(dtlen) + int64(datalen) + 4

		count++
	}

	idxwriter.Flush()
	sf.file.Close()
	sf.idx.Close()
	os.Remove(sf.filename + ".dirty")
	return nil
}
