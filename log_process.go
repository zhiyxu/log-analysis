package main

import (
	"fmt"
	"strings"
	"time"
)

type Reader interface {
	Read(chan string)
}

type Writer interface {
	Write(chan string)
}

type LogProcess struct {
	rc chan string
	wc chan string
	// change to interface type
	// then whatever struct implement interface could passed into
	// e.g ReadFromStdin, ReadFromDB... not only ReadFromFile
	// e.g WriteToFile, WriteToStdout... not only WriteToInfluxDB
	// improve scalability a lot
	read  Reader
	write Writer
}

// a type of Reader
type ReadFromFile struct {
	path string // File path
}

func (r ReadFromFile) Read(rc chan string) {
	message := "string"
	rc <- message
}

// a type of Writer
type WriteToInfluxDB struct {
	InfluxDBDsn string // InfluxDB data source
}

func (w WriteToInfluxDB) Write(wc chan string) {
	fmt.Println(<-wc)
}

func (l *LogProcess) Process() {
	data := <-l.rc
	l.wc <- strings.ToUpper(data)
}

func main() {
	lp := &LogProcess{
		rc: make(chan string),
		wc: make(chan string),
		read: ReadFromFile{
			path: "/tmp/access.log",
		},
		write: WriteToInfluxDB{
			InfluxDBDsn: "username&password",
		},
	}

	go lp.read.Read(lp.rc)
	go lp.Process()
	go lp.write.Write(lp.wc)

	time.Sleep(time.Second)
}
