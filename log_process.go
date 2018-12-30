package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	TimeLocal                    time.Time
	BytesSent                    int
	Path, Method, Scheme, Status string
	UpstreamTime, RequestTime    float64
}

type Reader interface {
	Read(chan []byte)
}

type Writer interface {
	Write(chan *Message)
}

type LogProcess struct {
	rc chan []byte
	wc chan *Message
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

func (w WriteToInfluxDB) Write(wc chan *Message) {
	for data := range wc {
		fmt.Println(data)
	}
}

func (l *LogProcess) Process() {
	// 1. Read every line of log from Read Channel
	// 2. Parse useful data from every line of log
	// 3. Write useful data into Write Channel
	/*
		'$remote_addr\t$http_x_forwarded_for\t$remote_user\t[$time_local]\t$scheme\t"$request"\t$status\t$body_bytes_sent\t"$http_referer"\t"$http_user_agent"\t"$gzip_ratio"\t$upstream_response_time\t$request_time'
	*/

	rep := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)

	loc, _ := time.LoadLocation("Asia/Shanghai")
	for v := range l.rc {
		ret := rep.FindStringSubmatch(string(v))
		if len(ret) < 13 {
			log.Println("wrong input data:", v)
			continue
		}

		timeLocal, err := time.ParseInLocation("02/Jan/2006:15:04:05 +0000", ret[4], loc)
		if err != nil {
			log.Println("time parse error:", err)
			continue
		}

		request := ret[6]
		requestSli := strings.Split(request, " ")
		if len(requestSli) < 3 {
			log.Println("input request wrong:", request)
			continue
		}
		method := strings.TrimLeft(requestSli[0], "\"")
		u, err := url.Parse(requestSli[1])
		if err != nil {
			log.Println("input url parse error:", err)
			continue
		}
		path := u.Path
		scheme := ret[5]
		status := ret[7]
		bytesSent, _ := strconv.Atoi(ret[8])
		upstreamTime, _ := strconv.ParseFloat(ret[12], 64)
		requestTime, _ := strconv.ParseFloat(ret[13], 64)

		l.wc <- &Message{
			TimeLocal:    timeLocal,
			Path:         path,
			Method:       method,
			Scheme:       scheme,
			Status:       status,
			BytesSent:    bytesSent,
			UpstreamTime: upstreamTime,
			RequestTime:  requestTime,
		}
	}
}

func main() {
	lp := &LogProcess{
		rc: make(chan []byte),
		wc: make(chan *Message),
		read: ReadFromFile{
			path: "./access.log",
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
