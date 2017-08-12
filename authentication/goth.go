package authentication

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/amazon"
	"github.com/markbates/goth/providers/auth0"
	"github.com/markbates/goth/providers/battlenet"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/box"
	"github.com/markbates/goth/providers/dailymotion"
	"github.com/markbates/goth/providers/deezer"
	"github.com/markbates/goth/providers/digitalocean"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/dropbox"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/fitbit"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/gplus"
	"github.com/markbates/goth/providers/heroku"
	"github.com/markbates/goth/providers/instagram"
	"github.com/markbates/goth/providers/intercom"
	"github.com/markbates/goth/providers/lastfm"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/meetup"
	"github.com/markbates/goth/providers/onedrive"
	"github.com/markbates/goth/providers/paypal"
	"github.com/markbates/goth/providers/salesforce"
	"github.com/markbates/goth/providers/slack"
	"github.com/markbates/goth/providers/soundcloud"
	"github.com/markbates/goth/providers/spotify"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/goth/providers/stripe"
	"github.com/markbates/goth/providers/twitch"
	"github.com/markbates/goth/providers/twitter"
	"github.com/markbates/goth/providers/uber"
	"github.com/markbates/goth/providers/wepay"
	"github.com/markbates/goth/providers/xero"
	"github.com/markbates/goth/providers/yahoo"
	"github.com/markbates/goth/providers/yammer"
	"github.com/spf13/viper"
)

type key int

const gothStateKey key = iota

const stateTokenLength = 256 / 8

type providerConstructor func(key, secret, callbackUrl string) goth.Provider

var availableProviders = map[string]providerConstructor{
	"twitter": func(key, secret, callbackUrl string) goth.Provider {
		return twitter.New(key, secret, callbackUrl)
	},
	"facebook": func(key, secret, callbackUrl string) goth.Provider {
		return facebook.New(key, secret, callbackUrl)
	},
	"fitbit": func(key, secret, callbackUrl string) goth.Provider {
		return fitbit.New(key, secret, callbackUrl)
	},
	"gplus": func(key, secret, callbackUrl string) goth.Provider {
		return gplus.New(key, secret, callbackUrl)
	},
	"github": func(key, secret, callbackUrl string) goth.Provider {
		return github.New(key, secret, callbackUrl)
	},
	"spotify": func(key, secret, callbackUrl string) goth.Provider {
		return spotify.New(key, secret, callbackUrl)
	},
	"linkedin": func(key, secret, callbackUrl string) goth.Provider {
		return linkedin.New(key, secret, callbackUrl)
	},
	"lastfm": func(key, secret, callbackUrl string) goth.Provider {
		return lastfm.New(key, secret, callbackUrl)
	},
	"twitch": func(key, secret, callbackUrl string) goth.Provider {
		return twitch.New(key, secret, callbackUrl)
	},
	"dropbox": func(key, secret, callbackUrl string) goth.Provider {
		return dropbox.New(key, secret, callbackUrl)
	},
	"digitalocean": func(key, secret, callbackUrl string) goth.Provider {
		return digitalocean.New(key, secret, callbackUrl)
	},
	"bitbucket": func(key, secret, callbackUrl string) goth.Provider {
		return bitbucket.New(key, secret, callbackUrl)
	},
	"instagram": func(key, secret, callbackUrl string) goth.Provider {
		return instagram.New(key, secret, callbackUrl)
	},
	"intercom": func(key, secret, callbackUrl string) goth.Provider {
		return intercom.New(key, secret, callbackUrl)
	},
	"box": func(key, secret, callbackUrl string) goth.Provider {
		return box.New(key, secret, callbackUrl)
	},
	"salesforce": func(key, secret, callbackUrl string) goth.Provider {
		return salesforce.New(key, secret, callbackUrl)
	},
	"amazon": func(key, secret, callbackUrl string) goth.Provider {
		return amazon.New(key, secret, callbackUrl)
	},
	"yammer": func(key, secret, callbackUrl string) goth.Provider {
		return yammer.New(key, secret, callbackUrl)
	},
	"onedrive": func(key, secret, callbackUrl string) goth.Provider {
		return onedrive.New(key, secret, callbackUrl)
	},
	"battlenet": func(key, secret, callbackUrl string) goth.Provider {
		return battlenet.New(key, secret, callbackUrl)
	},
	"yahoo": func(key, secret, callbackUrl string) goth.Provider {
		return yahoo.New(key, secret, callbackUrl)
	},
	"slack": func(key, secret, callbackUrl string) goth.Provider {
		return slack.New(key, secret, callbackUrl)
	},
	"stripe": func(key, secret, callbackUrl string) goth.Provider {
		return stripe.New(key, secret, callbackUrl)
	},
	"wepay": func(key, secret, callbackUrl string) goth.Provider {
		return wepay.New(key, secret, callbackUrl, "view_user")
	},
	"paypal": func(key, secret, callbackUrl string) goth.Provider {
		return paypal.New(key, secret, callbackUrl)
	},
	"steam": func(key, secret, callbackUrl string) goth.Provider {
		return steam.New(key, callbackUrl)
	},
	"heroku": func(key, secret, callbackUrl string) goth.Provider {
		return heroku.New(key, secret, callbackUrl)
	},
	"uber": func(key, secret, callbackUrl string) goth.Provider {
		return uber.New(key, secret, callbackUrl)
	},
	"soundcloud": func(key, secret, callbackUrl string) goth.Provider {
		return soundcloud.New(key, secret, callbackUrl)
	},
	"gitlab": func(key, secret, callbackUrl string) goth.Provider {
		return gitlab.New(key, secret, callbackUrl)
	},
	"dailymotion": func(key, secret, callbackUrl string) goth.Provider {
		return dailymotion.New(key, secret, callbackUrl, "email")
	},
	"deezer": func(key, secret, callbackUrl string) goth.Provider {
		return deezer.New(key, secret, callbackUrl, "email")
	},
	"discord": func(key, secret, callbackUrl string) goth.Provider {
		return discord.New(key, secret, callbackUrl, discord.ScopeIdentify, discord.ScopeEmail)
	},
	"meetup": func(key, secret, callbackUrl string) goth.Provider {
		return meetup.New(key, secret, callbackUrl)
	},
	"auth0": func(key, secret, callbackUrl string) goth.Provider {
		return auth0.New(key, secret, callbackUrl, viper.GetString("goth.auth0.domain"))
	},
	"xero": func(key, secret, callbackUrl string) goth.Provider {
		return xero.New(key, secret, callbackUrl)
	},
}

func (auth *Authentication) ConfigureGoth(store sessions.Store, appURL string) {
	gothic.Store = store

	providers := []goth.Provider{}
	for name, constructor := range availableProviders {
		viperKey := fmt.Sprintf("goth.%s.key", name)
		if !viper.IsSet(viperKey) {
			continue
		}

		viperSecret := fmt.Sprintf("goth.%s.secret", name)
		callbackURL := fmt.Sprintf("%s/%s/callback/", appURL, name)

		log.Printf("Auto configured goth for: %s", name)

		providers = append(providers, constructor(
			viper.GetString(viperKey),
			viper.GetString(viperSecret),
			callbackURL,
		))
	}

	goth.UseProviders(providers...)
}

func (auth *Authentication) ConfigureGothRoutes(router chi.Router) {
	gothic.GetProviderName = func(r *http.Request) (string, error) {
		return chi.URLParam(r, "provider"), nil
	}
	gothic.SetState = func(r *http.Request) string {
		return r.Context().Value(gothStateKey).(string)
	}

	// TODO: use render for error pages
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state := r.URL.Query().Get("state")
			if len(state) == 0 {
				randomState, err := getRandomString(stateTokenLength)
				if err != nil {
					http.Error(w, fmt.Sprintf("could not generate random state: %v", err), http.StatusInternalServerError)
					return
				}

				state = randomState
			}

			ctx := context.WithValue(r.Context(), gothStateKey, state)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// TODO: redirect back on login or logout
	router.Route("/{provider}", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
				auth.loginGoth(w, r, gothUser)
			} else {
				gothic.BeginAuthHandler(w, r)
			}
		})

		router.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
			gothUser, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			auth.loginGoth(w, r, gothUser)
		})

		router.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			// TODO: general /auth/logout that auto determines provider
			gothic.Logout(w, r)
			err := auth.authorization.Logout(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			http.Redirect(w, r, "/", http.StatusFound)
		})
	})
}

func (auth *Authentication) loginGoth(w http.ResponseWriter, r *http.Request, gothUser goth.User) {
	id := fmt.Sprintf("goth:%s:%s", chi.URLParam(r, "provider"), gothUser.UserID)

	err := auth.authorization.Login(w, r, id, gothUser.Name, gothUser.Email, gothUser.AvatarURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func getRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
