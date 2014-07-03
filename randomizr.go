package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var (
	fname string
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
	options := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	return os.OpenFile(fname, options, 0644)
}

func main() {
	flag.Parse()

	f, err := outFile()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	for t := range time.Tick(100 * time.Millisecond) {
		fmt.Fprintf(f, "%v %v %v\n", t.UnixNano(), randomString(32), randomString(32))
	}
}

func init() {
	flag.StringVar(&fname, "file", "", "destination file to which random data will be appended")
	rand.Seed(time.Now().UnixNano())
}
