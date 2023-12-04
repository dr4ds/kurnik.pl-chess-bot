package utils

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/fatih/color"
)

var DEBUG = true

func LogSuccess(v ...interface{}) {
	if DEBUG {
		color.Set(color.FgGreen)
		defer color.Unset()
		log.Println(v)
	}
}
func LogError(v ...interface{}) {
	if DEBUG {
		color.Set(color.FgRed)
		defer color.Unset()
		log.Println(v)
	}
}
func LogInfo(v ...interface{}) {
	if DEBUG {
		color.Set(color.FgCyan)
		defer color.Unset()
		log.Println(v)
	}
}
func LogWarning(v ...interface{}) {
	if DEBUG {
		color.Set(color.FgYellow)
		defer color.Unset()
		log.Println(v)
	}
}
func StringToInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
func IntToString(n int) string {
	return strconv.Itoa(n)
}

type GoCounter struct {
	i   int64
	mux sync.Mutex
}

func (c *GoCounter) Add() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.i++
}
func (c *GoCounter) Subtract() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.i--
}
func ReadLinesFromFile(filepath string) (lines []string) {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func LoadJsonFile(filename string, v interface{}) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, v)
	return err
}
func IndexByte(c byte, b []byte) int {
	for i := 0; i < len(b); i++ {
		if c == b[i] {
			return i
		}
	}
	return -1
}
