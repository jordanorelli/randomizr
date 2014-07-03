package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var (
	fname   string
	fappend bool
	reopen  bool
	freq    float64
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
    if fappend {
        options |= os.O_APPEND
    } else {
        options |= os.O_TRUNC
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
	f, err := outFile()
	if err != nil {
		fmt.Printf("ERROR: unable to open outfile: %v", err)
		return
	}
	defer f.Close()

	for line := range c {
		if _, err := f.WriteString(line); err != nil {
			fmt.Printf("ERROR: unable to write line: %v", err)
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

func main() {
	flag.Parse()
	c := make(chan string)

	writeLines(c)

	for t := range time.Tick(time.Duration(1e9 / freq)) {
		c <- fmt.Sprintf("%v %v %v\n", t.UnixNano(), randomString(32), randomString(32))
	}
}

func init() {
	flag.StringVar(&fname, "file", "", "destination file to which random data will be written")
	flag.BoolVar(&fappend, "append", false, "append to file instead of truncating file on open")
	flag.BoolVar(&reopen, "reopen", false, "reopen file handle on every write instead of using a persistent handle")
	flag.Float64Var(&freq, "freq", 1, "frequency in hz at which lines will be written")
	rand.Seed(time.Now().UnixNano())
}
