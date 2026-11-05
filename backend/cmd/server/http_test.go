package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestServeHTTPShutdown(t *testing.T) {
	for _, forceClose := range []bool{false, true} {
		name := "drains active request"
		if forceClose {
			name = "closes connection on timeout"
		}
		t.Run(name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			release := make(chan struct{})
			finished := make(chan struct{})
			server := &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer close(finished)
					close(started)
					if forceClose {
						<-r.Context().Done()
						return
					}
					<-release
					_, _ = io.WriteString(w, "completed")
				}),
			}
			t.Cleanup(func() {
				close(release)
				_ = server.Close()
			})
			stopped := make(chan struct{})
			server.RegisterOnShutdown(func() { close(stopped) })
			timeout := 5 * time.Second
			if forceClose {
				timeout = 100 * time.Millisecond
			}
			result := make(chan error, 1)
			go func() {
				result <- serveHTTP(ctx, server, listener, timeout, zap.NewNop())
			}()
			client := &http.Client{
				Timeout:   5 * time.Second,
				Transport: &http.Transport{Proxy: nil, DisableKeepAlives: true},
			}
			response := make(chan error, 1)
			go func() {
				res, err := client.Get("http://" + listener.Addr().String())
				if err == nil {
					defer res.Body.Close()
					var body []byte
					body, err = io.ReadAll(res.Body)
					if err == nil && string(body) != "completed" {
						err = errors.New("incomplete response")
					}
				}
				response <- err
			}()
			select {
			case <-started:
			case err := <-response:
				t.Fatalf("request failed before reaching handler: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("request did not reach handler")
			}
			cancel()
			waitHTTPEvent(t, stopped)
			if !forceClose {
				select {
				case err := <-result:
					t.Fatalf("returned before active request completed: %v", err)
				default:
				}
				release <- struct{}{}
			}
			select {
			case err := <-result:
				if forceClose && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("want shutdown deadline, got %v", err)
				}
				if !forceClose && err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("server did not stop")
			}
			waitHTTPEvent(t, finished)
			if err := <-response; (err != nil) != forceClose {
				t.Fatalf("response error = %v, force close = %v", err, forceClose)
			}
			conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
			if err == nil {
				conn.Close()
				t.Fatal("listener still accepts connections")
			}
		})
	}
}

func TestServeHTTPListenerFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	err = serveHTTP(context.Background(), &http.Server{}, listener, time.Second, zap.NewNop())
	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("want listener error, got %v", err)
	}
}

func waitHTTPEvent(t *testing.T, event <-chan struct{}) {
	t.Helper()
	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for HTTP lifecycle event")
	}
}
