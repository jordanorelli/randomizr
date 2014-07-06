package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	fname      string        // output filename
	freq       float64       // frequency at which lines are written
	ftruncate  bool          // whether or not to truncate file on open
	pidfile    string        // path of pidfile to write out
	reopen     bool          // whether or not to reopen the file handle on every line write
	tsformat   string        // timestamp format
	lineLength lengthArg     // length of the lines to be generated
	ts         func() string // function to get a timestamp string
	line       func() string // function to generate a line
)

type lengthArg struct {
	n      int
	random bool
}

func (l *lengthArg) String() string {
	return "length."
}

func (l *lengthArg) Set(v string) error {
	if i, err := strconv.Atoi(v); err == nil {
		*l = lengthArg{n: i}
		return nil
	}

	switch v {
	case "rand", "random":
		*l = lengthArg{random: true}
		return nil
	default:
		return fmt.Errorf("bad length arg: %s", v)
	}
}

func (l *lengthArg) mkLineFn() (func() string, error) {
	if ts == nil {
		ts = mkTsFn()
	}
	tsLen := len(ts())
	if l.random {
		return func() string {
			return randomString(rand.Intn(80 - tsLen))
		}, nil
	}
	if l.n == 0 {
		l.n = 80
	}
	if !l.random && tsLen > l.n {
		return nil, fmt.Errorf("line length %d is too small for timestamps like %s", l.n, ts())
	}
	return func() string {
		return randomString(l.n - tsLen)
	}, nil
}

// generates a pseudorandom string of length n that is composed of alphanumeric
// characters.
func randomString(n int) string {
	var alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, n)
	for i := 0; i < len(buf); i++ {
		buf[i] = alpha[rand.Intn(len(alpha)-1)]
	}
	return string(buf)
}

func outFile() (*os.File, error) {
	if fname == "" {
		return os.Stdout, nil
	}
	options := os.O_WRONLY | os.O_CREATE
	if ftruncate {
		options |= os.O_TRUNC
	} else {
		options |= os.O_APPEND
	}

	return os.OpenFile(fname, options, 0644)
}

func reopenWrite(c chan string) {
	for line := range c {
		f, err := outFile()
		if err != nil {
			fmt.Printf("ERROR: unable to open outfile: %v", err)
			continue
		}
		if _, err := f.WriteString(line); err != nil {
			fmt.Printf("ERROR: unable to write line: %v", err)
		}
		f.Close()
	}
}

func regularWrite(c chan string) {
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP, syscall.SIGUSR1)

START:
	f, err := outFile()
	if err != nil {
		fmt.Printf("ERROR: unable to open outfile: %v", err)
		return
	}

	for {
		select {
		case line := <-c:
			if _, err := f.WriteString(line); err != nil {
				fmt.Printf("ERROR: unable to write line: %v", err)
			}
		case <-hup:
			fmt.Fprintf(f, "%s HUP\n", ts())
			f.Close()
			goto START
		}
	}
}

func writeLines(c chan string) {
	if reopen {
		go reopenWrite(c)
	} else {
		go regularWrite(c)
	}
}

func writePid() {
	if pidfile == "" {
		return
	}
	f, err := os.OpenFile(pidfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR unable to open pidfile: %v", err)
		return
	}
	fmt.Fprintln(f, os.Getpid())
}

func mkTsFn() func() string {
	switch tsformat {
	case "":
		return func() string {
			t := time.Now()
			return fmt.Sprintf("%s %4.4d", t.Format("15:04:05"), t.Nanosecond()/1e5)
		}
	case "ns":
		return func() string {
			t := time.Now()
			return fmt.Sprintf("%d", t.UnixNano())
		}
	case "ms":
		return func() string {
			t := time.Now()
			return fmt.Sprintf("%d", t.UnixNano()/1e3)
		}
	case "epoch", "unix":
		return func() string {
			t := time.Now()
			return fmt.Sprintf("%d", t.Unix())
		}
	default:
		return func() string {
			return time.Now().Format(tsformat)
		}
	}
}

func flags() (err error) {
	flag.Parse()
	ts = mkTsFn()
	line, err = lineLength.mkLineFn()
	return
}

func main() {
	if err := flags(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	writePid()
	c := make(chan string)

	writeLines(c)

	for _ = range time.Tick(time.Duration(1e9 / freq)) {
		c <- fmt.Sprintf("%s %s\n", ts(), line())
	}
}

func init() {
	flag.StringVar(&fname, "file", "", "destination file to which random data will be written")
	flag.StringVar(&tsformat, "ts-format", "", "timestamp format")
	flag.StringVar(&pidfile, "pidfile", "", "file to which a pid is written")
	flag.BoolVar(&ftruncate, "truncate", false, "truncate file on opening instead of appending")
	flag.BoolVar(&reopen, "reopen", false, "reopen file handle on every write instead of using a persistent handle")
	flag.Float64Var(&freq, "freq", 10, "frequency in hz at which lines will be written")
	flag.Var(&lineLength, "line-length", "length of the lines to be generated (in bytes)")
	rand.Seed(time.Now().UnixNano())
}
