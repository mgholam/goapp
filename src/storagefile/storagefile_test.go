package storagefile_test

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"syscall"
	"testing"
	"time"

	// jsoniter "github.com/json-iterator/go"
	"goapp/src/storagefile"
)

type Book struct {
	ID     int       `json:"id"`
	Title  string    `json:"name"`
	Author string    `json:"author"`
	Rating int       `json:"rating"`
	Date   time.Time `json:"date" gorm:"column:date"`
}

func Test_speedstringonly(t *testing.T) {
	sf, e := storagefile.Open("speed.dat")
	if e != nil {
		panic(e)
	}
	defer sf.Close()
	defer func() {
		os.Remove("speed.dat")
		os.Remove("speed.dat.idx")
	}()

	b := Book{
		ID:     1,
		Title:  "dune",
		Author: "frank herbert",
		Rating: 5,
		Date:   time.Now(),
	}
	by, _ := json.Marshal(&b)

	dt := time.Now()

	count := 640_000

	fmt.Printf("saving %d ...\n", count)

	for i := 1; i < count; i++ {
		sf.Save("test", by)
	}
	fmt.Println("time : ", time.Since(dt))

	dt = time.Now()

	fmt.Printf("reading %d ...\n", count)

	for i := 1; i < count; i++ {
		_, _, e := sf.Get(int64(i))
		if e != nil {
			fmt.Println(e)
			fmt.Println("error reading ", i)
		}
	}
	fmt.Println("time : ", time.Since(dt))
}

func Test_iterate(t *testing.T) {

	sf, e := storagefile.Open("iterate.dat")
	if e != nil {
		panic(e)
	}
	defer sf.Close()
	defer func() {
		os.Remove("iterate.dat")
		os.Remove("iterate.dat.idx")
	}()

	b := Book{
		// ID:     1,
		Title:  "dune",
		Author: "frank herbert",
		Rating: 5,
		Date:   time.Now(),
	}
	by, _ := json.Marshal(&b)

	count := 100

	fmt.Printf("saving %d ...\n", count)

	for i := 0; i < count; i++ {
		sf.Save("test", by)
	}

	var j int64 = 1

	for hdr := range sf.Iterate() {
		if hdr.Id != j {
			t.Error("failed")
			return
		}
		j++
		fmt.Printf("%d : %v\r\n", hdr.Id, string(hdr.Data))
	}
}

func Test_concurrent_write(t *testing.T) {
	sf, e := storagefile.Open("ccwrite.dat")
	if e != nil {
		panic(e)
	}
	defer sf.Close()
	defer func() {
		os.Remove("ccwrite.dat")
		os.Remove("ccwrite.dat.idx")
	}()

	count := 10_000
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func(w *sync.WaitGroup) {
		for i := 1; i <= count; i++ {
			sf.Save("22", []byte("111111"))
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 1; i <= count; i++ {
			sf.Save("11", []byte("222222"))

		}
		w.Done()
	}(&wg)
	wg.Wait()

	i := sf.Count()
	t.Log(i)
	if i < int64(2*count) {
		t.Error("count does not match")
	}

}

func Test_run_100k_read_write(t *testing.T) {

	sf, e := storagefile.Open("docs.dat")
	if e != nil {
		panic(e)
	}
	defer sf.Close()
	defer func() {
		os.Remove("docs.dat")
		os.Remove("docs.dat.idx")
	}()

	count := 100_000

	fmt.Println("saving count", count)
	dosave(sf, count)
	count = int(sf.Count())
	fmt.Println("reading count", count)
	doread(sf, count)

	h, _ := sf.GetHeader(10)
	fmt.Println(h)

	tt, ss, _ := sf.GetString(10)
	fmt.Println(tt, ss)

	tt, b, _ := sf.Get(10)
	fmt.Println(tt, b)
}

func Test_rebuild(t *testing.T) {
	defer func() {
		os.Remove("rebuild.dat")
		os.Remove("rebuild.dat.idx")
	}()
	os.Remove("rebuild.dat")
	sf, e := storagefile.Open("rebuild.dat")
	if e != nil {
		panic(e)
	}

	for i := 1; i <= 1000; i++ {
		sf.Save("22", []byte("111111"))
	}
	for i := 1; i <= 1000; i++ {
		tt, s, _ := sf.GetString(int64(i))
		if s != "111111" || tt != "22" {
			t.Fail()
		}
	}

	sf.Close()
	os.WriteFile("rebuild.dat.dirty", []byte("hello"), 0644)
	sf, e = storagefile.Open("rebuild.dat")
	if e != nil {
		panic(e)
	}
	for i := 1; i <= 1000; i++ {
		tt, s, e := sf.GetString(int64(i))
		if e != nil {
			t.Log(e)
			t.Fail()
		}
		if s != "111111" || tt != "22" {
			t.Fail()
		}
	}

	sf.Close()

}

func Test_open_close_twice(t *testing.T) {
	defer func() {
		os.Remove("oc2.dat")
		os.Remove("oc2.dat.idx")
	}()
	os.Remove("oc2.dat")
	sf, e := storagefile.Open("oc2.dat")
	if e != nil {
		panic(e)
	}

	sf.Save("11", []byte("1111111"))
	sf.Save("11", []byte("1111111"))
	sf.Close()

	sf, e = storagefile.Open("oc2.dat")
	if e != nil {
		panic(e)
	}

	sf.Save("22", []byte("2222222"))
	sf.Save("22", []byte("2222222"))

	if sf.Count() != int64(4) {
		t.Fail()
	}

	for i := 1; i <= 4; i++ {
		h, e := sf.GetHeader(int64(i))
		if e != nil {
			t.Log(e)
			t.Fail()
		}
		if h.Id != int64(i) {
			t.Log("id mismatch", i, h.Id)
			t.Fail()
		}
	}
	tt, d, e := sf.GetString(1)
	if e != nil {
		t.Log(e)
		t.Fail()
	}
	if tt != "11" || d != "1111111" {
		t.Log("type and data mismatch")
		t.Fail()
	}
	sf.Close()
}

func Test_invalid(t *testing.T) {
	defer func() {
		os.Remove("inv.dat")
		os.Remove("inv.dat.idx")
	}()
	os.Remove("inv.dat")
	sf, e := storagefile.Open("inv.dat")
	if e != nil {
		panic(e)
	}
	defer sf.Close()

	_, _, e = sf.Get(1)
	if e != nil {
		t.Log(e)
	}

	_, _, e = sf.GetString(1)
	if e != nil {
		t.Log(e)
	}
}

func doread(sf *storagefile.StorageFile, count int) {
	t := time.Now()
	for i := 1; i <= count; i++ {
		h, e := sf.GetHeader(int64(i))
		if e != nil {
			fmt.Println("err", e, i)
			continue
		}
		if h.Type != "/api/book" {
			fmt.Println("data not matching", i)
		}
	}
	fmt.Println("read time =", time.Since(t))
}

func dosave(sf *storagefile.StorageFile, count int) {

	t := time.Now()

	// var json = jsoniter.ConfigCompatibleWithStandardLibrary

	for i := 1; i <= count; i++ {
		book := Book{
			ID:     i,
			Author: "tolkien",
			Title:  "lord of the rings",
			Rating: 5,
			Date:   now(),
		}
		b, _ := json.Marshal(book)

		sf.Save("/api/book", b)
	}
	fmt.Println("save time =", time.Since(t))
}

func now() time.Time {
	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)
	return time.Unix(0, syscall.TimevalToNsec(tv))
}
