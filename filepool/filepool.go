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
)

type openFile struct {
	File   *os.File
	fp     *filePool
	closed bool
}

type FilePool interface {
	OpenFile(string) *openFile
	Wait()
}

type filePool struct {
	wg                 sync.WaitGroup
	mu                 sync.Mutex
	chCanOpen          chan bool
	chOpenAqcuiredLock chan bool
	chClose            chan bool
	fileLimit          uint64
	openFiles          uint64
	openedFiles        int
	closedFiles        int
}

func NewFilePool(limit uint64) *filePool {
	ch1, ch2, ch3 := make(chan bool), make(chan bool), make(chan bool) //Create the channels

	// Create the filepool
	fp := filePool{sync.WaitGroup{},
		sync.Mutex{},
		ch1,
		ch2,
		ch3,
		limit,
		0,
		0,
		0,
	}

	go fp.opener()
	go fp.closer()

	return &fp
}

func (fp *filePool) GetLimitOfOpenFiles() uint64 {
	return fp.fileLimit
}

func (fp *filePool) changeOpenFiles(add bool) {
	if add {
		fp.openFiles += 1
		fp.openedFiles += 1
	} else {
		fp.openFiles -= 1
		fp.closedFiles += 1
	}

	if fp.openFiles > fp.fileLimit {
		fmt.Println("Open files: ", fp.openFiles)
		fmt.Println("Opened files: ", fp.openedFiles)
		fmt.Println("Closed files: ", fp.closedFiles)
	}
}

// This function handles the closing of files, and makes sure the resources become available again.
func (fp *filePool) closer() {
	for {
		<-fp.chClose

		fp.mu.Lock()

		fp.changeOpenFiles(false)
		fp.wg.Done()

		fp.mu.Unlock()
	}
}

// This function handles the opening of files, and blocks if no files are allowed to be opened again.
func (fp *filePool) opener() {
	for {
		// Quickly read the amount of currently opened files so we know whether we can open another one.
		fp.mu.Lock()

		// Calculate whether we can open another file.
		// If so, wait till a file wants to be opened, and open one.
		if fp.openFiles < fp.fileLimit {
			fp.mu.Unlock() // Unlock the readlock again
			fp.chCanOpen <- true
			fp.chOpenAqcuiredLock <- true
		} else {
			fp.mu.Unlock() // Unlock the readlock again
		}
	}
}

// This function opens the file from the filePool
func (fp *filePool) OpenFile(filename string) *openFile {
	<-fp.chCanOpen //Receive permission to open a file

	fp.mu.Lock()            // First aqcuire the lock
	<-fp.chOpenAqcuiredLock // Then allow the opener to once again cycle

	fp.wg.Add(1)
	fp.changeOpenFiles(true)

	fp.mu.Unlock()

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0775)

	if err != nil {
		panic(err)
	}
	return &openFile{f, fp, false}
}

// This function closes a file from the filePool
func (f *openFile) Close() {
	if !f.closed {

		f.File.Close()
		f.fp.chClose <- true

	} else {
		fmt.Println("FilePool: Warning: File is closed twice.")
	}
}

func (f *openFile) WriteBytes(bytes [][8]byte) {
	var s string

	for _, b := range bytes {
		s += string(b[:])
	}
	f.File.WriteString(s)
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
