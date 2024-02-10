package lastfm

import "fmt"

const (
	statusOK     = "ok"
	statusFailed = "failed"
)

type Response[T any] struct {
	Status string `xml:"status,attr"`
	Error  *Error `xml:"error,omitempty"`

	Value T `xml:",any,omitempty"`
}

type Error struct {
	Code    int    `xml:"code,attr"`
	Message string `xml:",chardata"`
}

func (e Error) Error() string {
	return fmt.Sprintf("Last.FM API Error Code %d: %s", e.Code, e.Message)
}

type ScrobbleResult struct {
	Accepted int `xml:"accepted,attr"`
	Ignored  int `xml:"ignored,attr"`

	Tracks []Track `xml:"scrobble"`
}

type String struct {
	Corrected bool   `xml:"corrected,attr"`
	Value     string `xml:",chardata"`
}

type Track struct {
	Track       String `xml:"track"`
	Artist      String `xml:"artist"`
	Album       String `xml:"album"`
	AlbumArtist String `xml:"albumArtist,omitempty"`
	Timestamp   int    `xml:"timestamp"`

	IgnoreReason String `xml:"ignoreMessage,omitempty"`
}

type Session struct {
	Name       string `xml:"name"`
	Key        string `xml:"key"`
	Subscriber int    `xml:"subscriber"`
}
