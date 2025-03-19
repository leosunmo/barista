// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mockio

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStdout(t *testing.T) {
	stdout := Stdout()

	require.Empty(t, stdout.ReadNow(), "starts empty")

	_, err := stdout.ReadUntil('x', 1*time.Millisecond)
	require.Equal(t, io.EOF, err, "EOF when timeout expires")

	io.WriteString(stdout, "te")
	io.WriteString(stdout, "st")

	val, err := stdout.ReadUntil('s', 1*time.Millisecond)
	require.Nil(t, err, "no error when multiple writes")
	require.Equal(t, "tes", val, "read joins output from multiple writes")
	require.Equal(t, "t", stdout.ReadNow(), "remaining string after ReadUntil returned by ReadNow")

	wait := make(chan struct{})

	go (func(w io.Writer) {
		io.WriteString(w, "ab")
		io.WriteString(w, "cdef")
		wait <- struct{}{}
	})(stdout)

	<-wait
	val, err = stdout.ReadUntil('c', 1*time.Millisecond)
	require.Nil(t, err, "no error when multiple writes in goroutine")
	require.Equal(t, "abc", val, "read joins output from multiple writes in goroutine")
	require.Equal(t, "def", stdout.ReadNow(), "remaining string after ReadUntil returned by ReadNow")

	go (func(w io.Writer) {
		io.WriteString(w, "ab")
		wait <- struct{}{}
		<-wait
		io.WriteString(w, "cd")
		wait <- struct{}{}
	})(stdout)

	<-wait
	val, err = stdout.ReadUntil('d', 1*time.Millisecond)
	require.Equal(t, io.EOF, err, "EOF when delimiter write does not happen within timeout")
	require.Equal(t, "ab", val, "returns content written before timeout")

	wait <- struct{}{}
	<-wait
	require.Equal(t, "cd", stdout.ReadNow(), "continues normally after timeout")

	go func() {
		val, err = stdout.ReadUntil('b', 50*time.Millisecond)
		// This is a test of data races, the assertions are not very important.
		require.Equal(t, err, io.EOF)
		require.Equal(t, 50, len(val))
		wait <- struct{}{}
	}()
	for i := 0; i < 50; i++ {
		go (func(w io.Writer) {
			io.WriteString(w, "a")
		})(stdout)
	}
	<-wait

	bytes := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for i := 0; i < 20; i++ {
		go (func(w io.Writer) {
			wait <- struct{}{}
			for _, b := range bytes {
				w.Write([]byte{b})
			}
		})(stdout)
		<-wait
		for _, b := range bytes {
			_, err = stdout.ReadUntil(b, 10*time.Millisecond)
			require.NoError(t, err)
		}
	}

	stdout.ShouldError(io.ErrClosedPipe)
	_, err = stdout.Write([]byte{','})
	require.Error(t, err, "next error on write")
	_, err = stdout.Write([]byte{':'})
	require.NoError(t, err, "next error is consumed")
}

func TestTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode.")
	}

	stdout := Stdout()
	wait := make(chan struct{})

	go (func(w io.Writer) {
		<-wait
		io.WriteString(w, "ab")
		time.Sleep(4 * time.Second)
		io.WriteString(w, "cd")
		time.Sleep(4 * time.Second)
		io.WriteString(w, "ef")
		time.Sleep(4 * time.Second)
		io.WriteString(w, "gh")
		time.Sleep(4 * time.Second)
		io.WriteString(w, "ijklmnop")
		wait <- struct{}{}
	})(stdout)

	wait <- struct{}{}
	val, err := stdout.ReadUntil('i', 10*time.Second)
	require.Equal(t, io.EOF, err, "EOF when delimiter write does not happen within timeout")
	require.Equal(t, "abcdef", val, "returns content written before timeout")

	val, err = stdout.ReadUntil('j', 8*time.Second)
	require.Equal(t, "ghij", val, "returns only content up to delimiter")
	require.NoError(t, err, "if delimiter is written within timeout")
	<-wait
	require.Equal(t, "klmnop", stdout.ReadNow(), "subsequent readnow returns content after timeout")
}

func TestWaiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping waiting test in short mode.")
	}

	stdout := Stdout()

	io.WriteString(stdout, "123")
	_, err := stdout.ReadUntil('3', time.Second)
	require.NoError(t, err)

	require.False(t, stdout.WaitForWrite(time.Second),
		"wait when buffer is empty")

	wait := make(chan struct{})

	go (func(w io.Writer) {
		<-wait
		io.WriteString(w, "ab")
		<-wait
		time.Sleep(4 * time.Second)
		io.WriteString(w, "cd")
	})(stdout)

	wait <- struct{}{}
	require.True(t, stdout.WaitForWrite(time.Second),
		"wait when buffer is not empty")
	require.True(t, stdout.WaitForWrite(time.Second),
		"wait when buffer is still not empty")

	stdout.ReadNow()
	wait <- struct{}{}
	require.False(t, stdout.WaitForWrite(time.Second),
		"wait after buffer is emptied")

	require.True(t, stdout.WaitForWrite(6*time.Second),
		"write before timeout expires")
}

func TestStdin(t *testing.T) {
	stdin := Stdin()

	type readResult struct {
		contents string
		err      error
	}

	request := make(chan int)
	result := make(chan readResult)

	go (func(reader io.Reader, result chan<- readResult) {
		for r := range request {
			out := make([]byte, r)
			count, err := reader.Read(out)
			result <- readResult{string(out)[:count], err}
		}
	})(stdin, result)

	resultOrTimeout := func(timeout time.Duration) (readResult, bool) {
		select {
		case r := <-result:
			return r, true
		case <-time.After(timeout):
			return readResult{}, false
		}
	}

	request <- 1
	_, ok := resultOrTimeout(time.Millisecond)
	require.False(t, ok, "read should not return when nothing has been written")

	stdin.WriteString("")
	r, ok := resultOrTimeout(time.Second)
	require.True(t, ok, "read returns empty string if written")
	require.Equal(t, readResult{"", nil}, r, "read returns empty string if written")

	stdin.WriteString("test")
	request <- 2
	r, ok = resultOrTimeout(time.Second)
	require.True(t, ok, "read should not time out when content is available")
	require.Equal(t, readResult{"te", nil}, r, "read returns only requested content when more is available")

	request <- 2
	r, ok = resultOrTimeout(time.Second)
	require.True(t, ok, "read should not time out when content is available")
	require.Equal(t, readResult{"st", nil}, r, "read returns leftover on subsequent call")

	stdin.WriteString("abcd")
	request <- 10
	r, ok = resultOrTimeout(time.Second)
	require.True(t, ok, "read should not time out when returning partial content")
	require.Equal(t, readResult{"abcd", nil}, r, "read returns partial content if available")

	stdin.WriteString("12")
	stdin.WriteString("34")
	stdin.WriteString("56")
	stdin.WriteString("78")
	request <- 8
	r, ok = resultOrTimeout(time.Second)
	require.True(t, ok, "read should not time out when concatenating")
	require.Equal(t, readResult{"12345678", nil}, r, "read returns concatenation of multiple writes")

	request <- 4
	_, ok = resultOrTimeout(time.Millisecond)
	require.False(t, ok, "read should wait for a write when buffer has been emptied")

	stdin.Write([]byte("xyz"))
	stdin.Write([]byte("abc"))
	r, ok = resultOrTimeout(time.Second)
	require.True(t, ok, "read does not time out when returning partial content")
	require.Equal(t, readResult{"xyz", nil}, r, "read returns contents of first write (does not wait)")

	request <- 1
	r = <-result
	require.Equal(t, readResult{"a", nil}, r, "remaining writes are read by later requests")

	request <- 10
	r = <-result
	require.Equal(t, readResult{"bc", nil}, r, "remaining writes are read by later requests")

	stdin.WriteString("foo")
	stdin.ShouldError(io.ErrClosedPipe)
	request <- 3
	r = <-result
	require.Error(t, r.err, "next error on read")
	request <- 3
	r = <-result
	require.NoError(t, r.err, "next error is consumed")
	require.Equal(t, "foo", r.contents)

	request <- 4
	// Give request some time to block.
	time.Sleep(10 * time.Millisecond)
	stdin.ShouldError(io.ErrShortBuffer)
	r = <-result
	require.Error(t, r.err, "ShouldError when Read is blocked")
}
