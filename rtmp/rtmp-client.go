package rtmp

import (
	"fmt"
	"github.com/onedss/onedss/rtsp"
	"net/url"
	"time"
)

type RTMPClient struct {
	URL         string
	Path        string
	CustomPath  string //custom path for pusher
	UdpHostPort string
	SDPRaw      string
	InitFlag    bool

	Agent string

	RTPHandles  []func(*rtsp.RTPPack)
	StopHandles []func()
}

func (client *RTMPClient) String() string {
	return fmt.Sprintf("client[%s]", client.URL)
}

func (client *RTMPClient) GetPath() string {
	return client.Path
}

func (client *RTMPClient) GetInitFlag() bool {
	return client.InitFlag
}

func (client *RTMPClient) GetCustomPath() string {
	return client.CustomPath
}

func (client *RTMPClient) GetURL() string {
	return client.URL
}

func (client *RTMPClient) GetSDPRaw() string {
	return client.SDPRaw
}

func (client *RTMPClient) AddRTPHandles(f func(*rtsp.RTPPack)) {
	client.RTPHandles = append(client.RTPHandles, f)
}

func (client *RTMPClient) AddStopHandles(f func()) {
	client.StopHandles = append(client.StopHandles, f)
}

func NewRTMPClient(rawUrl string, sendOptionMillis int64, agent string) (client *RTMPClient, err error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return
	}
	client = &RTMPClient{
		URL:   rawUrl,
		Path:  url.Path,
		Agent: agent,
	}
	return
}

func (client *RTMPClient) Start() bool {
	return false
}

func (client *RTMPClient) Stop() {

}

func (client *RTMPClient) Init(timeout time.Duration) (err error) {
	return nil
}
