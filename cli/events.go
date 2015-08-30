package cli

import (
	"encoding/json"
	"errors"
	"golang.org/x/net/websocket"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
)

type EventsCommand struct {
	App    *CliApp
	Id     string
	Format string
}

// __jsonp1__([{"id":"1","channel":"/meta/handshake","successful":true,"version":"1.0","supportedConnectionTypes":["long-polling","cross-origin-long-polling","callback-polling","websocket","eventsource","in-process"],"clientId":"jn6854gieb65ld2i0nkwz3yf2rtczwl","advice":{"reconnect":"retry","interval":0,"timeout":25000}}]);

type FayeAuth struct {
	AuthToken string `json:"auth_token"`
}
type FayeMsg struct {
	Channel                  string    `json:"channel"`
	ClientId                 string    `json:"clientId,omitempty"`
	ConnectionType           string    `json:"connectionType,omitempty"`
	Id                       string    `json:"id,omitempty"`
	Subscription             string    `json:"subscription,omitempty"`
	Ext                      *FayeAuth `json:"ext,omitempty"`
	Version                  string    `json:"version,omitempty"`
	SupportedConnectionTypes []string  `json:"supportedConnectionTypes,omitempty"`
	Successful               bool      `json:"successful,omitempty"`
	Error                    string    `json:"error,omitempty"`
	Data                     string    `json:"data,omitempty"`
}

func sendmsg(ws *websocket.Conn, msgs ...FayeMsg) error {
	jmsg, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	if _, err := ws.Write(jmsg); err != nil {
		return err
	}
	return nil
}
func recvmsg(ws *websocket.Conn) ([]FayeMsg, error) {
	var jmsg = make([]byte, 4096)
	var n int
	n, err := ws.Read(jmsg)
	if err != nil {
		return nil, err
	}
	msgl := make([]FayeMsg, 1)
	err = json.Unmarshal(jmsg[0:n], &msgl)
	if err != nil {
		return nil, err
	}
	if msgl[0].Successful == false {
		return nil, errors.New(msgl[0].Error)
	}
	return msgl, nil
}

func (l *TokenCommand) watch(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()

	token, err := l.App.Client.Token()
	if token == nil {
		l.App.Fatalf("No cached OAuth token found for %s", l.App.ClientName)
	}

	handshake := FayeMsg{
		Channel:                  "/meta/handshake",
		Version:                  "1.0",
		SupportedConnectionTypes: []string{"long-polling", "websocket"},
	}

	origin := "https://events.gb1s.brightbox.com"
	url := "wss://events.gb1s.brightbox.com/stream"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return err
	}
	log.Println("Handshake...")
	err = sendmsg(ws, handshake)
	if err != nil {
		return err
	}

	msgl, err := recvmsg(ws)
	if err != nil {
		return err
	}
	log.Println(msgl[0].SupportedConnectionTypes)
	cid := msgl[0].ClientId
	log.Println("Hands shook, got Client Id " + cid)

	connect := FayeMsg{
		Channel:        "/meta/connect",
		ClientId:       cid,
		ConnectionType: "websocket",
		Ext: &FayeAuth{
			AuthToken: token.AccessToken,
		},

	}

	log.Println("Connecting...")
	err = sendmsg(ws, connect)
	if err != nil {
		return err
	}
	msgl, err = recvmsg(ws)
	if err != nil {
		return err
	}
	log.Println("Subscribing...")
	subscribe := FayeMsg{
		Channel:      "/meta/subscribe",
		ClientId:     cid,
		Subscription: "/account/" + l.App.accountId(),
		Ext: &FayeAuth{
			AuthToken: token.AccessToken,
		},
	}
	err = sendmsg(ws, subscribe)
	if err != nil {
		return err
	}
	msgl, err = recvmsg(ws)
	if err != nil {
		return err
	}
	log.Println("Subscribed")

	var jmsg = make([]byte, 4096)
	var n int
	for {
		n, err = ws.Read(jmsg)
		if err == io.EOF {
			log.Println("EOF Disconnected")
			return nil
		}
		if err != nil {
			return err
		}
		log.Println(string(jmsg[0:n]))
	}
}

func ConfigureEventsCommand(app *CliApp) {
	cmd := TokenCommand{App: app}
	events := app.Command("events", "view event stream")
	events.Command("watch", "listen for events and output them").Action(cmd.watch)
}
