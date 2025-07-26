// implment oauth2 flow for gmail
// https://developers.google.com/identity/protocols/oauth2/web-server
package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type oauthClient struct {
	cilentID     string
	clientSecret string
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type OAuthCredentials struct {
	AccessToken  string
	ExpiresAt    time.Time
	RefreshToken string // RefreshToken is not set when the token is obtained from the refresh token endpoint
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

func (client *oauthClient) getCredentials() *OAuthCredentials {
	u, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		panic(err)
	}

	addr, codeChan, shutdown := serveOnce()

	q := u.Query()
	redirectURI := "http://" + addr
	q.Set("client_id", client.cilentID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("access_type", "offline")
	q.Set("scope", "https://www.googleapis.com/auth/gmail.readonly")

	u.RawQuery = q.Encode()
	fmt.Printf("open this URL in your browser: %s\n", u.String())

	code := <-codeChan
	fmt.Println("received grant code")

	shutdown()

	params := url.Values{}
	params.Set("code", code)
	params.Set("client_id", client.cilentID)
	params.Set("client_secret", client.clientSecret)
	params.Set("redirect_uri", redirectURI)
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

	expiresAt := time.Now().Add(time.Duration(tokenRes.ExpiresIn) * time.Second)

	return &OAuthCredentials{
		AccessToken:  tokenRes.AccessToken,
		ExpiresAt:    expiresAt,
		RefreshToken: tokenRes.RefreshToken,
	}
}

func (client *oauthClient) refreshGmailCreds(refreshToken string) *OAuthCredentials {
	params := url.Values{
		"client_id":     []string{client.cilentID},
		"client_secret": []string{client.clientSecret},
		"refresh_token": []string{refreshToken},
		"grant_type":    []string{"refresh_token"},
	}

	res, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		log.Fatalf("unable to create request: %s", err.Error())
	}

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		log.Fatalf("unable to refresh token: %d - %s", res.StatusCode, string(body))
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("unable to read response body: %s", err.Error())
	}

	var tr tokenResponse
	err = json.Unmarshal(b, &tr)
	if err != nil {
		log.Fatalf("unable to unmarshal response body: %s", err.Error())
	}

	log.Printf("gmail: access token refresh success")
	return &OAuthCredentials{
		AccessToken: tr.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}
}
