package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

func execute(host, port string) (err error) {
	listner, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if closeErr := listner.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
				return
			}
			log.Println(closeErr)
		}
	}()

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		err = handle(conn)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return
}

func handle(conn net.Conn) (err error) {
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
				return
			}
			log.Println(closeErr)
		}
	}()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		log.Printf("%s", buf[:n])
		return nil
	}

	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	if requestLineEnd == -1 {
		log.Print("requestLineEndErr: ", requestLineEnd)
		return nil
	}

	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		log.Print("partsErr: ", parts)
		return nil
	}
	method, path, version := parts[0], parts[1], parts[2]

	if method != "GET" {
		log.Print("wrong method")
		return nil
	}

	if version != "HTTP/1.1" {
		log.Print("wrong version")
		return nil
	}

	if path == "/" {
		body, err := ioutil.ReadFile("static/index.html")
		if err != nil {
			return fmt.Errorf("can't read index.html: %w", err)
		}
		marker := "{{now}}"
		now := time.Now().String()
		body = bytes.ReplaceAll(body, []byte(marker), []byte(now))
		_, err = conn.Write([]byte(
			version + "200 OK\r\n" +
				"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
				"Content-Type: text/html\r\n" +
				"Connection: close\r\n" +
				"\r\n" +
				string(body),
		))
		if err != nil {
			return err
		}
	}
	return nil
}