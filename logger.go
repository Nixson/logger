package logger

import (
	"encoding/json"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	l "log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

var Info *l.Logger
var out io.Writer
var logChannel chan []byte

func output(caller int, logInterface ...interface{}) {
	var buf []byte
	_, file, line, ok := runtime.Caller(caller)
	if !ok {
		file = "???"
		line = 0
	}
	formatHeader(&buf, time.Now(), file, line)
	buf = append(buf, "\t"...)
	toString(&buf, logInterface)
	logChannel <- buf
}
func fast(buf *[]byte, logInterface []interface{}) {
	*buf = append(*buf, fmt.Sprintln(logInterface...)...)
}
func toString(buf *[]byte, logInterface []interface{}) {
	str := make([]interface{}, len(logInterface))
	for num, element := range logInterface {
		types := reflect.TypeOf(element)
		if types == nil {
			continue
		}
		var bodyStr string
		switch types.Kind() {
		case reflect.String:
			bodyStr = element.(string)
		default:
			body, _ := json.Marshal(element)
			bodyStr = string(body)
		}
		str[num] = bodyStr + " "
		//		str = append(str, bodyStr)
	}
	*buf = append(*buf, strings.TrimSpace(fmt.Sprint(str...))...)
	*buf = append(*buf, "\n"...)
}
func PrintLn(logInterface ...interface{}) {
	output(2, logInterface...)
}
func Printf(format string, v ...interface{}) {
	output(2, fmt.Sprintf(format, v...))
}
func Println(logInterface ...interface{}) {
	output(2, logInterface...)
}
func PrintUp(logInterface ...interface{}) {
	output(3, logInterface...)
}
func Fatal(logInterface ...interface{}) {
	output(2, logInterface...)
	os.Exit(1)
}
func send(logChannel chan []byte) {
	for value := range logChannel {
		out.Write(value)
	}
}
func Close() {
	close(logChannel)
}

func init() {
	logChannel = make(chan []byte, 500)
	filename, _ := filepath.Abs("./bin/logs/access.log")
	e, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	//	defer e.Close()
	Info = l.New(e, "", l.Ldate|l.Lmicroseconds|l.Lshortfile)
	out = io.MultiWriter(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    25, // megabytes after which new file is created
		MaxBackups: 1,  // number of backups
		MaxAge:     1,  //days
	} /*, os.Stdout*/)
	Info.SetOutput(io.MultiWriter( /*&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    25, // megabytes after which new file is created
			MaxBackups: 1,  // number of backups
			MaxAge:     1,  //days
		} , */os.Stdout))
	go send(logChannel)

}

func formatHeader(buf *[]byte, t time.Time, file string, line int) {
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, '.')
	itoa(buf, t.Nanosecond()/1e3, 6)
	*buf = append(*buf, ' ')
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, line, -1)
	*buf = append(*buf, ": "...)
}
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
