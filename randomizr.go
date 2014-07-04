package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	fname     string        // output filename
	freq      float64       // frequency at which lines are written
	ftruncate bool          // whether or not to truncate file on open
	pidfile   string        // path of pidfile to write out
	reopen    bool          // whether or not to reopen the file handle on every line write
	tsformat  string        // timestamp format
	ts        func() string // function to get a timestamp string
)

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

func flags() {
	flag.Parse()
	switch tsformat {
	case "":
		ts = func() string {
			t := time.Now()
			return fmt.Sprintf("%s %4.4d", t.Format("15:04:05"), t.Nanosecond()/1e5)
		}
	case "ns":
		ts = func() string {
			t := time.Now()
			return fmt.Sprintf("%d", t.UnixNano())
		}
	case "ms":
		ts = func() string {
			t := time.Now()
			return fmt.Sprintf("%d", t.UnixNano()/1e3)
		}
	case "epoch", "unix":
		ts = func() string {
			t := time.Now()
			return fmt.Sprintf("%d", t.Unix())
		}
	default:
		ts = func() string {
			return time.Now().Format(tsformat)
		}
	}
}

func main() {
	flags()
	writePid()
	c := make(chan string)

	writeLines(c)

	for _ = range time.Tick(time.Duration(1e9 / freq)) {
		c <- fmt.Sprintf("%s %s %s\n", ts(), randomString(32), randomString(32))
	}
}

func init() {
	flag.StringVar(&fname, "file", "", "destination file to which random data will be written")
	flag.StringVar(&tsformat, "ts-format", "", "timestamp format")
	flag.StringVar(&pidfile, "pidfile", "", "file to which a pid is written")
	flag.BoolVar(&ftruncate, "truncate", false, "truncate file on opening instead of appending")
	flag.BoolVar(&reopen, "reopen", false, "reopen file handle on every write instead of using a persistent handle")
	flag.Float64Var(&freq, "freq", 10, "frequency in hz at which lines will be written")
	rand.Seed(time.Now().UnixNano())
}
