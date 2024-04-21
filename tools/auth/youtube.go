package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/coddmeistr/pocketbase/tools/types"
	"golang.org/x/oauth2"
)

var _ Provider = (*Youtube)(nil)

// NameYoutube is the unique name of the Youtube provider.
const NameYoutube string = "youtube"

// Youtube allows authentication via Youtube OAuth2.
//
// Youtube oauth uses Google oauth integration so better dont touch RawUser field in AuthUser instance.
type Youtube struct {
	*baseProvider
	youtubeInfoUrl string
}

// NewYoutubeProvider creates new Youtube provider instance with some defaults.
func NewYoutubeProvider() *Youtube {
	return &Youtube{&baseProvider{
		ctx:         context.Background(),
		displayName: "Youtube",
		pkce:        true,
		scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/youtube.readonly",
		},
		authUrl:    "https://accounts.google.com/o/oauth2/auth",
		tokenUrl:   "https://accounts.google.com/o/oauth2/token",
		userApiUrl: "https://www.googleapis.com/oauth2/v1/userinfo",
	},
		"https://www.googleapis.com/youtube/v3/channels?mine=true&part=snippet",
	}
}

// FetchAuthUser returns an AuthUser instance based the Youtube's user api.
func (p *Youtube) FetchAuthUser(token *oauth2.Token) (*AuthUser, error) {
	data, err := p.FetchRawUserData(token)
	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(data, &rawUser); err != nil {
		return nil, err
	}

	extracted := struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}{}
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, err
	}

	// Fetch youtube channel info
	youtubeChannelResponse := struct {
		items []struct {
			id      string
			snippet struct {
				title     string
				customUrl string
			}
		}
	}{}
	youtubeUrl, err := url.ParseRequestURI(p.youtubeInfoUrl)
	if err != nil {
		return nil, err
	}
	req := &http.Request{
		Method: http.MethodGet,
		URL:    youtubeUrl,
	}
	token.SetAuthHeader(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &youtubeChannelResponse)
	if err != nil {
		return nil, err
	}
	if len(youtubeChannelResponse.items) == 0 || youtubeChannelResponse.items[0].snippet.title == "" {
		return nil, errors.New("no youtube account data found")
	}
	channel := youtubeChannelResponse.items[0]

	user := &AuthUser{
		Id:           channel.id,
		Name:         channel.snippet.title,
		Username:     channel.snippet.customUrl,
		AvatarUrl:    extracted.Picture,
		RawUser:      rawUser,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	user.Expiry, _ = types.ParseDateTime(token.Expiry)

	if extracted.VerifiedEmail {
		user.Email = extracted.Email
	}

	return user, nil
}
