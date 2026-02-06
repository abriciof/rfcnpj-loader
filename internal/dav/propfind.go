package dav

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type MultiStatus struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Responses []Response `xml:"response"`
}

type Response struct {
	Href     string     `xml:"href"`
	Propstat []Propstat `xml:"propstat"`
}

type Propstat struct {
	Prop   Prop   `xml:"prop"`
	Status string `xml:"status"`
}

type Prop struct {
	GetLastModified  string `xml:"getlastmodified"`
	GetContentLength int64  `xml:"getcontentlength"`
	GetContentType   string `xml:"getcontenttype"`
}

type Item struct {
	Href          string
	ContentLength int64
	ContentType   string
	LastModified  string
}

func (c *Client) ListZips(ctx context.Context, listURL string) ([]Item, error) {
	body := `<?xml version="1.0"?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:getlastmodified/>
    <d:getcontentlength/>
    <d:getcontenttype/>
  </d:prop>
</d:propfind>`

	req, err := http.NewRequestWithContext(ctx, "PROPFIND", listURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("PROPFIND falhou (%d): %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ms MultiStatus
	if err := xml.Unmarshal(raw, &ms); err != nil {
		return nil, fmt.Errorf("erro parse XML PROPFIND: %w", err)
	}

	items := make([]Item, 0, len(ms.Responses))
	for _, r := range ms.Responses {
		href := strings.TrimSpace(r.Href)
		if !strings.HasSuffix(strings.ToLower(href), ".zip") {
			continue
		}

		var chosen Prop
		for _, ps := range r.Propstat {
			if strings.Contains(ps.Status, "200") {
				chosen = ps.Prop
				break
			}
		}

		items = append(items, Item{
			Href:          href,
			ContentLength: chosen.GetContentLength,
			ContentType:   chosen.GetContentType,
			LastModified:  chosen.GetLastModified,
		})
	}

	return items, nil
}
