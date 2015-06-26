package mitml

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"appengine/user"
)

const (
	inviteUri         = "/api/users.admin.invite"
	kindRequest       = "request"
	configFile        = "config.json"
	avoidCallingSlack = true
)

var config struct {
	Team         string `json:"team"`
	CommandToken string `json:"command-token"`
	ApiToken     string `json:"api-token"`
}

type Request struct {
	Email       string
	First       string
	Last        string
	ChannelId   string
	ChannelName string
	UserId      string
	UserName    string
	Time        time.Time
	Invited     bool
	Succeeded   bool
}

func (r *Request) Store(ctx appengine.Context, key *datastore.Key) (*datastore.Key, error) {
	if key == nil {
		key = datastore.NewIncompleteKey(ctx, kindRequest, nil)
	}

	return datastore.Put(ctx, key, r)
}

func inviteUrl() string {
	return fmt.Sprintf("https://%s.slack.com%s?t=%d",
		config.Team,
		inviteUri,
		time.Now().Unix())
}

func (r *Request) Invite(ctx appengine.Context) error {
	res, err := urlfetch.Client(ctx).PostForm(inviteUrl(), url.Values{
		"email":      {r.Email},
		"first_name": {r.First},
		"last_name":  {r.Last},
		"set_active": {"true"},
		"_attempts":  {"1"},
		"token":      {config.ApiToken},
	})
	if err != nil {
		ctx.Errorf("email=%s, first=%s, last=%s, error=%s",
			r.Email,
			r.First,
			r.Last,
			err)
		return errors.New("slack is not responding")
	}
	defer res.Body.Close()

	var rsp struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&rsp); err != nil {
		ctx.Errorf("email=%s, first=%s, last=%s, error=%s",
			r.Email,
			r.First,
			r.Last,
			err)
		return errors.New("slack is not responding")
	}

	if !rsp.Ok {
		ctx.Errorf("email=%s, first=%s, last=%s, error=%s",
			r.Email,
			r.First,
			r.Last,
			rsp.Error)
		return errors.New(rsp.Error)
	}

	return nil
}

func (r *Request) Parse(txt, channelId, channelName, userId, userName string) error {
	email, names, err := parseEmail(txt)
	if err != nil {
		return err
	}

	r.Email = email
	r.First, r.Last = parseNames(names)
	r.ChannelId = channelId
	r.ChannelName = channelName
	r.UserId = userId
	r.UserName = userName
	r.Time = time.Now()
	return nil
}

func (r *Request) Set(email, first, last string) {
	r.Email = email
	r.First = first
	r.Last = last
	r.Time = time.Now()
}

func LoadConfig() error {
	r, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer r.Close()

	return json.NewDecoder(r).Decode(&config)
}

func isEmailValid(email string) bool {
	ix := strings.Index(email, "@")
	if ix < 0 {
		return false
	}

	name, domain := email[:ix], email[ix+1:]
	if len(name) == 0 || len(domain) == 0 {
		return false
	}

	return !strings.Contains(domain, "@") && strings.Contains(domain, ".")
}

func parseEmail(txt string) (string, string, error) {
	email := txt
	rest := ""

	ix := strings.Index(txt, " ")
	if ix >= 0 {
		email = txt[:ix]
		rest = txt[ix+1:]
	}

	if !isEmailValid(email) {
		return email, rest, fmt.Errorf("%s is not a valid email address", email)
	}

	return email, strings.TrimSpace(rest), nil
}

func parseNames(txt string) (string, string) {
	chr := []rune(txt)

	// will hold all the parsed substrings
	var all []string

	for len(chr) > 0 {
		// each pass will accumulate a single term
		var cur []rune
		i, n := 0, len(chr)
		for ; i < n; i++ {
			if chr[i] == '"' {
				// when you find a ", just collect characters until
				// we find the closing "
				i++
				for ; i < n; i++ {
					if chr[i] == '"' {
						break
					}
					cur = append(cur, chr[i])
				}
			} else if chr[i] == ' ' {
				// when you find a space, we'll push the current term
				// into the all slice. If the current term is empty,
				// though, we don't want it. This allows coalescing spaces
				// on term boundaries.
				if len(cur) == 0 {
					chr = chr[1:]
					break
				}

				all = append(all, string(cur))
				cur = nil
				chr = chr[i+1:]
				break
			} else {
				cur = append(cur, chr[i])
			}
		}

		// did we terminate the last loop because we reached the end?
		if i == n {
			all = append(all, string(cur))
			chr = nil
		}
	}

	if len(all) == 0 {
		return "", ""
	}

	return all[0], strings.Join(all[1:], " ")
}

func writeErrorAsText(w http.ResponseWriter, err error) {
	if _, err := fmt.Fprintf(w,
		"Oh boy, that didn't work. As far as I can tell it was because %s",
		err); err != nil {
		log.Panic(err)
	}
}

func writeOkAsText(w http.ResponseWriter, req *Request) {
	if _, err := fmt.Fprintf(w,
		"Done. %s should get an email right ... about ... now!",
		req.Email); err != nil {
		log.Panic(err)
	}
}

func writeJson(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Panic(err)
	}
}

func writeErrorAsJson(w http.ResponseWriter, err error) {
	writeJson(w, map[string]interface{}{
		"ok":    false,
		"error": err.Error(),
	}, http.StatusOK)
}

func writeOkAsJson(w http.ResponseWriter) {
	writeJson(w, map[string]bool{
		"ok": true,
	}, http.StatusOK)
}

func InviteAlum(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method != "POST" {
		http.Error(w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	if r.FormValue("token") != config.CommandToken {
		http.Error(w,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden)
		return
	}

	var req Request

	if err := req.Parse(
		strings.TrimSpace(r.FormValue("text")),
		r.FormValue("channel_id"),
		r.FormValue("channel_name"),
		r.FormValue("user_id"),
		r.FormValue("user_name")); err != nil {
		writeErrorAsText(w, err)
		return
	}

	req.Invited = true
	if err := req.Invite(ctx); err != nil {
		req.Succeeded = false
		writeErrorAsText(w, err)
	} else {
		req.Succeeded = true
		writeOkAsText(w, &req)
	}

	if _, err := req.Store(ctx, nil); err != nil {
		log.Panic(err)
	}
}

func InviteMe(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method != "POST" {
		http.Error(w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	if !isEmailValid(email) {
		writeErrorAsJson(w, errors.New("invalid email"))
		return
	}

	var req Request

	req.Set(email,
		strings.TrimSpace(r.FormValue("first")),
		strings.TrimSpace(r.FormValue("last")))

	if _, err := req.Store(ctx, nil); err != nil {
		writeErrorAsJson(w, err)
	} else {
		writeOkAsJson(w)
	}

	// TODO(knorton): send an email
}

func ApproveRequest(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method != "POST" {
		http.Error(w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	if !user.IsAdmin(ctx) {
		http.Error(w,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden)
		return
	}

	key, err := datastore.DecodeKey(r.FormValue("key"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var req Request
	if err := datastore.Get(ctx, key, &req); err != nil {
		http.NotFound(w, r)
		return
	}

	req.Invited = true
	if err := req.Invite(ctx); err != nil {
		req.Succeeded = false
		writeErrorAsJson(w, err)
	} else {
		req.Succeeded = true
		writeOkAsJson(w)
	}

	if _, err := req.Store(ctx, key); err != nil {
		log.Panic(err)
	}
}

type fetchRequestsRsp struct {
	Key     *datastore.Key
	Request *Request
}

func FetchRequests(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if !user.IsAdmin(ctx) {
		http.Error(w,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden)
		return
	}

	var reqs []*fetchRequestsRsp
	for t := datastore.NewQuery(kindRequest).Order("-Time").Run(ctx); ; {
		req := &Request{}
		key, err := t.Next(req)
		if err == datastore.Done {
			break
		} else if err != nil {
			log.Panic(err)
		}

		reqs = append(reqs, &fetchRequestsRsp{
			Key:     key,
			Request: req,
		})
	}

	writeJson(w, reqs, http.StatusOK)
}

func init() {
	if err := LoadConfig(); err != nil {
		log.Panic(err)
	}

	http.HandleFunc("/api/v1/invite-alum", InviteAlum)
	http.HandleFunc("/api/v1/invite-me", InviteMe)

	// admin
	http.HandleFunc("/api/v1/approve-request", ApproveRequest)
	http.HandleFunc("/api/v1/requests", FetchRequests)
}
