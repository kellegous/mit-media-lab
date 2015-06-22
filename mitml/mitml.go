package mitml

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

const (
	inviteUri         = "/api/users.admin.invite"
	kindRequest       = "request"
	configFile        = "config.json"
	avoidCallingSlack = true
)

var config struct {
	Slack struct {
		Token string `json:"token"`
		Team  string `json:"team"`
	} `json:"slack"`
}

type EmailType int

const (
	EmailApproved EmailType = iota
	EmailValid
	EmailInvalid
)

type Request struct {
	Email   string
	First   string
	Last    string
	Time    time.Time
	Invited bool
}

type InviteErr struct {
	Err error
	Msg string
}

func (e *InviteErr) Error() string {
	return e.Err.Error()
}

func (r *Request) Store(ctx appengine.Context, key *datastore.Key) (*datastore.Key, error) {
	if key == nil {
		key = datastore.NewIncompleteKey(ctx, kindRequest, nil)
	}
	return datastore.Put(ctx, key, r)
}

func (r *Request) Invite(ctx appengine.Context) *InviteErr {
	res, err := urlfetch.Client(ctx).PostForm(inviteUrl(), url.Values{
		"email":      {r.Email},
		"first_name": {r.First},
		"last_name":  {r.Last},
		"set_active": {"true"},
		"_attempts":  {"1"},
		"token":      {config.Slack.Token},
	})
	if err != nil {
		return &InviteErr{
			Err: err,
			Msg: "Slack Error",
		}
	}
	defer res.Body.Close()

	var rsp struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}

	var buf bytes.Buffer

	if err := json.NewDecoder(io.TeeReader(res.Body, &buf)).Decode(&rsp); err != nil {
		return &InviteErr{
			Err: err,
			Msg: "Slack Error",
		}
	}

	if !rsp.Ok {
		return &InviteErr{
			Err: errors.New(rsp.Error),
			Msg: rsp.Error,
		}
	}

	return nil
}

func inviteUrl() string {
	return fmt.Sprintf("https://%s.slack.com%s?t=%d",
		config.Slack.Team,
		inviteUri,
		time.Now().Unix())
}

func EmailTypeFor(email string) EmailType {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 {
		return EmailInvalid
	}

	if strings.HasSuffix(parts[1], "media.mit.edu") {
		return EmailApproved
	}

	return EmailValid
}

func WriteJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Panic(err)
	}
}

func WriteJsonSuccess(w http.ResponseWriter) {
	WriteJson(w, map[string]interface{}{
		"ok": true,
	})
}

func WriteJsonError(w http.ResponseWriter, msg string) {
	WriteJson(w, map[string]interface{}{
		"error": msg,
		"ok":    false,
	})
}

func LoadConfig() error {
	r, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer r.Close()

	return json.NewDecoder(r).Decode(&config)
}

func init() {
	if err := LoadConfig(); err != nil {
		log.Panic(err)
	}

	http.HandleFunc("/api/v1/invite-me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w,
				http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		first := r.FormValue("first")
		last := r.FormValue("last")

		ctx := appengine.NewContext(r)

		t := EmailTypeFor(email)

		if t == EmailInvalid {
			WriteJsonError(w, "invalid email")
			return
		}

		req := Request{
			Email: email,
			First: first,
			Last:  last,
			Time:  time.Now(),
		}

		if t == EmailApproved && !avoidCallingSlack {
			if err := req.Invite(ctx); err != nil {
				ctx.Errorf("email=%s, first=%s, last=%s, error=%s",
					email,
					first,
					last,
					err.Error())
				WriteJsonError(w, fmt.Sprintf("%s", err.Msg))
				return
			}

			req.Invited = true
		} else {
			// email admin for approval
		}

		if _, err := req.Store(ctx, nil); err != nil {
			WriteJsonError(w, "server error")
			return
		}

		time.Sleep(5 * time.Second)
		WriteJsonSuccess(w)
	})
}
