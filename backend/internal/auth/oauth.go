package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

func GoogleConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func GitHubConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func FetchGoogleProfile(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (email, providerUserID string, err error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", "", fmt.Errorf("fetching google profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("reading google profile response: %w", err)
	}

	var profile struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return "", "", fmt.Errorf("parsing google profile: %w", err)
	}

	if profile.Sub == "" || profile.Email == "" {
		return "", "", fmt.Errorf("google profile missing required fields")
	}

	return profile.Email, profile.Sub, nil
}

func FetchGitHubProfile(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (email, providerUserID string, err error) {
	client := cfg.Client(ctx, token)

	// Fetch user ID
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return "", "", fmt.Errorf("fetching github user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("reading github user response: %w", err)
	}

	var user struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return "", "", fmt.Errorf("parsing github user: %w", err)
	}

	if user.ID == 0 {
		return "", "", fmt.Errorf("github user missing id")
	}

	providerUserID = fmt.Sprintf("%d", user.ID)

	// If email not public, fetch from emails endpoint
	if user.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			return "", "", fmt.Errorf("fetching github emails: %w", err)
		}
		defer emailResp.Body.Close()

		emailBody, err := io.ReadAll(emailResp.Body)
		if err != nil {
			return "", "", fmt.Errorf("reading github emails response: %w", err)
		}

		var emails []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}
		if err := json.Unmarshal(emailBody, &emails); err != nil {
			return "", "", fmt.Errorf("parsing github emails: %w", err)
		}

		for _, e := range emails {
			if e.Primary {
				user.Email = e.Email
				break
			}
		}
	}

	if user.Email == "" {
		return "", "", fmt.Errorf("github profile missing email")
	}

	return user.Email, providerUserID, nil
}

