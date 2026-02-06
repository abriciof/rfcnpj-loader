package email

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestCheckConnection_Success(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	errCh := make(chan error, 1)
	go runFakeSMTPServer(ln, errCh)

	host, port, err := splitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("splitHostPort failed: %v", err)
	}

	cfg := SMTPConfig{
		Host: host,
		Port: port,
	}

	if err := CheckConnection(cfg); err != nil {
		t.Fatalf("CheckConnection returned error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("fake smtp server failed: %v", err)
	}
}

func TestEnabled(t *testing.T) {
	t.Parallel()

	if Enabled(SMTPConfig{}) {
		t.Fatal("expected Enabled=false for empty config")
	}
	if !Enabled(SMTPConfig{User: "u", Pass: "p", To: "a@b.com"}) {
		t.Fatal("expected Enabled=true when user/pass/to are set")
	}
}

func TestParseRecipients(t *testing.T) {
	t.Parallel()

	got, err := parseRecipients("a@x.com, b@y.com; c@z.com")
	if err != nil {
		t.Fatalf("parseRecipients returned error: %v", err)
	}
	want := []string{"a@x.com", "b@y.com", "c@z.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected recipients: got=%v want=%v", got, want)
	}

	if _, err := parseRecipients(" , ; "); err == nil {
		t.Fatal("expected error when recipient list is empty")
	}
}

func runFakeSMTPServer(ln net.Listener, errCh chan<- error) {
	conn, err := ln.Accept()
	if err != nil {
		errCh <- err
		return
	}
	defer conn.Close()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	write := func(s string) error {
		if _, err := w.WriteString(s); err != nil {
			return err
		}
		return w.Flush()
	}

	if err := write("220 localhost ESMTP ready\r\n"); err != nil {
		errCh <- err
		return
	}

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}

		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
			if err := write("250-localhost\r\n250 AUTH PLAIN\r\n"); err != nil {
				errCh <- err
				return
			}
		case strings.HasPrefix(cmd, "NOOP"):
			if err := write("250 OK\r\n"); err != nil {
				errCh <- err
				return
			}
		case strings.HasPrefix(cmd, "QUIT"):
			if err := write("221 Bye\r\n"); err != nil {
				errCh <- err
				return
			}
			errCh <- nil
			return
		default:
			if err := write("250 OK\r\n"); err != nil {
				errCh <- err
				return
			}
		}
	}
}

func splitHostPort(addr string) (string, int, error) {
	host, p, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}

	var port int
	if _, err := fmt.Sscanf(p, "%d", &port); err != nil {
		return "", 0, err
	}
	return host, port, nil
}
