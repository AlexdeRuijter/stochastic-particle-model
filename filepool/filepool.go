/*
* This is to stay within the limits of the maximum amount of open files per thread.
* Sort of an worker pool, but then also handling the fileopening.
 */
package filepool

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

type openFile struct {
	File *os.File
	fp   *filePool
}

type FilePool interface {
	OpenFile(string) *openFile
	Wait()
}

type filePool struct {
	wg        sync.WaitGroup
	mu        sync.Mutex
	rwmu      sync.RWMutex
	chOpen    chan bool
	chClose   chan bool
	fileLimit uint64
	openFiles uint64
}

func NewFilePool(limit uint64) *filePool {
	chOpen, chClose := make(chan bool), make(chan bool) //Create the channels

	// Create the filepool
	fp := filePool{sync.WaitGroup{},
		sync.Mutex{},
		sync.RWMutex{},
		chOpen,
		chClose,
		limit,
		0,
	}

	go fp.opener()
	go fp.closer()

	return &fp
}

func (fp *filePool) changeOpenFiles(delta uint64, negative bool) {
	fp.rwmu.Lock() // We're changing things (maybe concurrently, so lock the writing aspect)
	// Make sure to unlock it though
	defer fp.rwmu.Unlock()

	if negative {
		fp.openFiles -= delta
	} else {
		fp.openFiles += delta
	}
}

// This function handles the closing of files, and makes sure the resources become available again.
func (fp *filePool) closer() {
	for {
		<-fp.chClose
		fp.changeOpenFiles(1, true)
		fp.wg.Done()
	}
}

// This function handles the opening of files, and blocks if no files are allowed to be opened again.
func (fp *filePool) opener() {
	for {
		// Quickly read the amount of currently opened files so we know whether we can open another one.
		fp.rwmu.RLock()
		nOpenFiles := fp.openFiles
		fp.rwmu.RUnlock()

		// Calculate whether we can open another file.
		// If so, wait till a file wants to be opened, and open one.
		if nOpenFiles < fp.fileLimit {
			fp.chOpen <- true
			fp.changeOpenFiles(1, false)
			fp.wg.Add(1)
		}

		// Make sure to let another goroutine run if available
		time.Sleep(time.Nanosecond)
	}
}

// This function opens the file from the filePool
func (fp *filePool) OpenFile(filename string) *openFile {
	<-fp.chOpen
	f, err := os.Open(filename)

	if err != nil {
		fp.chClose <- true
		panic(err)
	}

	return &openFile{f, fp}
}

// This function closes a file from the filePool
func (f *openFile) Close() {
	f.File.Close()
	f.fp.chClose <- true
}

func (fp *filePool) Wait() {
	fp.wg.Wait()
}

func TestfilePool() {
	syscall.Umask(0)
	os.Mkdir("data", 0775)
	f, err := os.Create("data/test.txt")
	if err != nil {
		panic(err)
	}

	f.WriteString("Hello!")

	f.Close()

	limit := uint64(8000)
	fp := NewFilePool(limit)

	fh := fp.OpenFile("data/test.txt")
	defer fh.Close()

	b := make([]byte, 6)

	fh.File.Read(b)

	fmt.Println(string(b))
}
