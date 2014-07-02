package main

import (
    "fmt"
    "math/rand"
    "time"
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

func main() {
    for t := range time.Tick(100 * time.Millisecond) {
        fmt.Printf("%v %v %v\n", t.UnixNano(), randomString(32), randomString(32))
    }
}

func init() {
    rand.Seed(time.Now().UnixNano())
}
