package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type installed struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type credentials struct {
	Installed installed `json:"installed"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func serveOnce() (string, <-chan string, func()) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	addr := l.Addr().String()
	s := &http.Server{}
	s.SetKeepAlivesEnabled(false)

	codeChan := make(chan string, 1)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		codeChan <- r.URL.Query().Get("code")
		w.Header().Set("Connection", "close")
		io.WriteString(w, "You can close this page now")
	})

	go func() {
		_ = s.Serve(l)
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Shutdown(ctx)
	}

	return addr, codeChan, shutdown
}

func getClientId() credentials {
	fd, err := os.Open("credentials.json")
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	b, err := io.ReadAll(fd)
	if err != nil {
		panic(err)
	}

	creds := credentials{}

	err = json.Unmarshal(b, &creds)
	if err != nil {
		panic(err)
	}

	return creds
}

func getGmailCreds() tokenResponse {
	u, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		panic(err)
	}

	clientCredentials := getClientId()
	addr, codeChan, shutdown := serveOnce()

	q := u.Query()
	redirect_uri := "http://" + addr
	q.Set("client_id", clientCredentials.Installed.ClientId)
	q.Set("redirect_uri", redirect_uri)
	q.Set("response_type", "code")
	q.Set("scope", "https://www.googleapis.com/auth/gmail.readonly")

	u.RawQuery = q.Encode()
	fmt.Printf("open this URL in your browser: %s\n", u.String())

	code := <-codeChan
	fmt.Println("received grant code")

	shutdown()

	params := url.Values{}
	params.Set("code", code)
	params.Set("client_id", clientCredentials.Installed.ClientId)
	params.Set("client_secret", clientCredentials.Installed.ClientSecret)
	params.Set("redirect_uri", redirect_uri)
	params.Set("grant_type", "authorization_code")
	fmt.Println("requesting access token ...")

	res, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		panic(err)
	}

	tokenRes := tokenResponse{}
	b, err := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		panic(fmt.Sprintf("error getting oauth2 tokens: %s", b))
	}

	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &tokenRes)
	if err != nil {
		panic(err)
	}
	fmt.Println("success!")
	return tokenRes
}
