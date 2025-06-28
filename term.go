/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/aymanbagabas/go-pty"
	"github.com/gorilla/websocket"
)

var errPortInUse = errors.New("port is in use")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { // github.com/gorilla/websocket.checkSameOrigin
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return false
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}

		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			host = u.Host

			if err.(*net.AddrError).Err != "missing port in address" {
				log.Println("Split host and port error:", err)
				return false
			}
		}

		return equalASCIIFold(host, "wails.localhost") || equalASCIIFold(host, "wails")
	},
}

type termSize struct {
	Rows int
	Cols int
}

type term struct {
	pty  pty.Pty
	size *termSize

	wsUrl string
}

func (t *term) Close() error {
	log.Println("closing terminal")

	err := t.pty.Close()
	t.pty = nil
	t.size = nil
	return err
}

func (t *term) GetWsUrl() string {
	return t.wsUrl
}

func (t *term) Resize(rows, cols int) error {
	if t.size.Rows == rows && t.size.Cols == cols {
		return nil
	}

	t.size.Rows = rows
	t.size.Cols = cols

	return t.pty.Resize(t.size.Cols, t.size.Rows)
}

func initTerm(exitCallback func()) (*term, error) {
	ptmx, err := pty.New()
	if err != nil {
		return nil, err
	}

	c := ptmx.Command(config.Shell.Path, config.Shell.Args...)
	home := os.Getenv("HOME")
	c.Env = append(os.Environ(), config.Shell.Envs.AsStrings()...)
	c.Dir = home
	if err := c.Start(); err != nil {
		log.Println("Start pty command error:", err)
		return nil, err
	}

	ts := termSize{
		Rows: 60,
		Cols: 10,
	}

	if err := ptmx.Resize(ts.Cols, ts.Rows); err != nil {
		log.Println("error resizing pty:", err)
		return nil, err
	}

	port, err := getPort()
	if err != nil {
		return nil, err
	}

	wsListenAddress := net.JoinHostPort("127.0.0.1", port)
	wsPath := "/ws/pty/"
	accessToken := generateRandomString(32)

	t := &term{
		pty:   ptmx,
		size:  &ts,
		wsUrl: "ws://" + wsListenAddress + wsPath + accessToken,
	}

	go func(t *term) {
		http.HandleFunc(wsPath+accessToken, getSshHandler(t, exitCallback))
		log.Fatal(http.ListenAndServe(wsListenAddress, nil))
	}(t)

	return t, nil
}

func getSshHandler(t *term, exitCallback func()) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if t == nil || t.pty == nil {
			return
		}

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer func() {
			err := wsConn.Close()
			if err != nil {
				log.Println("Close websocket connection error:", err)
			}
		}()

		go func() {
			buf := make([]byte, 1024)
			for {
				if t == nil || t.pty == nil {
					return
				}

				n, err := io.LimitReader(t.pty, 1024).Read(buf)

				if err != nil {
					if err != io.EOF {
						log.Println("Read from PTY stdout error:", err)
					}

					exitCallback()

					return
				}
				if n > 0 {
					err = wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
					if err != nil {
						log.Println("Write to WebSocket error:", err)
						return
					}
				}
			}
		}()

		for {
			messageType, p, err := wsConn.ReadMessage()

			if err != nil {
				if err != io.EOF {
					log.Println("Read from WebSocket error:", err)
				}
				return
			}
			if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
				if t == nil || t.pty == nil {
					return
				}

				_, err = io.Copy(t.pty, bytes.NewReader(p))
				if err != nil {
					log.Println("Write to PTY stdin error:", err)
					return
				}
			}
		}
	}
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(s int) string {
	b, err := generateRandomBytes(s)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}

func getPort() (string, error) {
	for i := 62400; i < 62500; i++ {
		p := strconv.Itoa(i)

		if err := checkPort(p); err != nil {
			if errors.Is(err, errPortInUse) {
				continue
			}

			return "", err
		}

		return p, nil
	}

	return "", errPortInUse
}

func checkPort(port string) error {
	timeout := 50 * time.Millisecond
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", port), timeout)
	if err != nil {
		return nil
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			return err
		}

		return errPortInUse
	}

	return nil
}

// github.com/gorilla/websocket.equalASCIIFold
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}
