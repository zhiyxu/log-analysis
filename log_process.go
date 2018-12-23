package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Reader interface {
	Read(chan []byte)
}

type Writer interface {
	Write(chan string)
}

type LogProcess struct {
	rc chan []byte
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

func (r ReadFromFile) Read(rc chan []byte) {
	// 1. Open File
	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error:%s", err.Error()))
	}

	// 2. Read File from Bottom per line
	//f.Seek(0, 2)
	rd := bufio.NewReader(f)

	for {
		line, err := rd.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("ReadBytes error:%s", err.Error()))
		}
		// 3. Pass data to rc channel, without \n
		rc <- line[:len(line)-1]
	}
}

// a type of Writer
type WriteToInfluxDB struct {
	InfluxDBDsn string // InfluxDB data source
}

func (w WriteToInfluxDB) Write(wc chan string) {
	for data := range wc {
		fmt.Println(data)
	}
}

func (l *LogProcess) Process() {
	for data := range l.rc {
		l.wc <- strings.ToUpper(string(data))
	}
}

func main() {
	lp := &LogProcess{
		rc: make(chan []byte),
		wc: make(chan string),
		read: ReadFromFile{
			path: "C:\\Users\\zhiyuxu\\go\\src\\github.com\\zhiyxu\\log-analysis\\access.log",
		},
		write: WriteToInfluxDB{
			InfluxDBDsn: "username&password",
		},
	}

	go lp.read.Read(lp.rc)
	go lp.Process()
	go lp.write.Write(lp.wc)

	time.Sleep(30 * time.Second)
}
