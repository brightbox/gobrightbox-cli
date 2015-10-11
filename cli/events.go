package cli

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"strings"
)

type EventsCommand struct {
	*CliApp
	Id     string
	Format string
}

type FayeAdvice struct {
	Timeout int
}
type FayeAuth struct {
	AuthToken string `json:"auth_token"`
}
type FayeMsg struct {
	Channel                  string           `json:"channel,omitempty"`
	ClientId                 string           `json:"clientId,omitempty"`
	ConnectionType           string           `json:"connectionType,omitempty"`
	Id                       string           `json:"id,omitempty"`
	Subscription             string           `json:"subscription,omitempty"`
	Ext                      *FayeAuth        `json:"ext,omitempty"`
	Version                  string           `json:"version,omitempty"`
	SupportedConnectionTypes []string         `json:"supportedConnectionTypes,omitempty"`
	Successful               bool             `json:"successful,omitempty"`
	Error                    string           `json:"error,omitempty"`
	Data                     *json.RawMessage `json:"data,omitempty"`
	Advice                   *FayeAdvice      `json:"advice,omitempty"`
}

type EventResource struct {
	Id    string
	Name  string
	Email *string
}
type Event struct {
	Id       string
	Action   string
	State    string
	Resource EventResource
	Account  EventResource
	Affects  []EventResource
	Touches  []EventResource
	User     EventResource
	Client   EventResource
}

func sendmsg(ws *websocket.Conn, msgs ...*FayeMsg) error {
	jmsg, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	if err = ws.WriteMessage(websocket.TextMessage, jmsg); err != nil {
		return err
	}
	return nil
}
func recvmsg(ws *websocket.Conn) ([]FayeMsg, error) {
	_, jmsg, err := ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	msgl := make([]FayeMsg, 1)
	err = json.Unmarshal(jmsg, &msgl)
	if err != nil {
		return nil, err
	}
	return msgl, nil
}

func (l *EventsCommand) watch(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()

	token, err := l.Client.TokenSource().Token()
	if err != nil {
		return err
	}

	handshake := FayeMsg{
		Channel:                  "/meta/handshake",
		Version:                  "1.0",
		SupportedConnectionTypes: []string{"long-polling", "websocket"},
	}

	url := "wss://events." + l.Client.findRegionDomain() + "/stream"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer ws.Close()
	err = sendmsg(ws, &handshake)
	if err != nil {
		return err
	}

	msgl, err := recvmsg(ws)
	if err != nil {
		return err
	}
	msg := msgl[0]
	cid := msg.ClientId

	err = sendmsg(ws, nil)

	connect := FayeMsg{
		Channel:        "/meta/connect",
		ClientId:       cid,
		ConnectionType: "websocket",
	}
	subscribe := FayeMsg{
		Channel:      "/meta/subscribe",
		ClientId:     cid,
		Subscription: "/account/" + l.accountId(),
		Ext:          &FayeAuth{AuthToken: token.AccessToken},
	}
	err = sendmsg(ws, &connect, &subscribe)
	if err != nil {
		return err
	}
	for {
		msgl, err := recvmsg(ws)
		if err == io.EOF {
			log.Println("EOF Disconnected")
			return nil
		}
		if err != nil {
			return err
		}
		for _, msg := range msgl {
			if msg.Data != nil && strings.HasPrefix(msg.Channel, "/account/acc") {
				e := Event{}
				err = json.Unmarshal(*msg.Data, &e)
				if err != nil {
					log.Println(err)
					continue
				}
				//log.Println(string(*msg.Data))
				var s string
				if e.User.Email != nil {
					s = fmt.Sprintf("<%s>", *e.User.Email)
				}
				if e.Client.Id != "" {
					s += fmt.Sprintf(" client:%s", e.Client.Id)
				}
				if e.Action != "" {
					s += fmt.Sprintf(" action:%s", e.Action)
				}
				if e.Resource.Id != "" {
					s += fmt.Sprintf(" resource:%s", e.Resource.Id)
				} else {
					s += fmt.Sprintf(" event:%s", string(*msg.Data))
				}
				if len(e.Affects) > 0 && (len(e.Affects) > 1 || e.Affects[0].Id != e.Resource.Id) {
					s += fmt.Sprintf(" affects:%s", collectById(e.Affects))
				}
				if len(e.Touches) > 0 && (len(e.Touches) > 1 || e.Touches[0].Id != e.Resource.Id) {
					s += fmt.Sprintf(" touches:%s", collectById(e.Touches))
				}
				log.Println(s)
			}
			if msg.Channel == "/meta/connect" {
				if msg.Successful {
					err = sendmsg(ws, &connect)
					if err != nil {
						log.Println("Connect error: ", err.Error())
					}
				} else {
					return fmt.Errorf("Event connection failure: " + msg.Error)
				}
			} else if msg.Channel == "/meta/subscribe" && !msg.Successful {
				return fmt.Errorf("Event subscription failure: " + msg.Error)
			}
		}
	}
}

func ConfigureEventsCommand(app *CliApp) {
	cmd := EventsCommand{CliApp: app}
	events := app.Command("events", "view event stream")
	events.Command("watch", "listen for events and output them").Action(cmd.watch)
}
