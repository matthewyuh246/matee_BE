package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/matthewyuh246/Matee/internal/domain"
	"github.com/matthewyuh246/Matee/internal/usecase"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type IUserController interface {
	GitHubLogin(c echo.Context) error
	GitHubCallback(c echo.Context) error
	Logout(c echo.Context) error
}

type userController struct {
	uu usecase.IUserUsecase
}

func NewUserController(uu usecase.IUserUsecase) IUserController {
	return &userController{uu}
}

var (
	clientID     string
	clientSecret string
	redirectURL  string
	config       *oauth2.Config
)

func init() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	clientID = os.Getenv("GITHUB_CLIENT_ID")
	clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURL = os.Getenv("GITHUB_CALLBACK_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		log.Fatalf("必要な環境変数(GITHUB_CLIENT_ID, CITHUB_CLIENT_SECRET, GITHUB_CALLBACK_URL)が設定されていません")
	}

	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  redirectURL,
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Error generating random string: %v", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (uc *userController) GitHubLogin(c echo.Context) error {
	state := generateRandomString(16)
	c.SetCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: state,
		Path:  "/",
	})
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (uc *userController) GitHubCallback(c echo.Context) error {

	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		log.Printf("Failed to get state cookie: %v", err)
		return c.String(http.StatusBadRequest, "Missing state cookie")
	}
	savedState := stateCookie.Value

	state := c.QueryParam("state")
	code := c.QueryParam("code")

	if state != savedState {
		log.Printf("Invalid state parameter: expected %s, got %s", savedState, state)
		return c.String(http.StatusBadRequest, "Invalid state parameter")
	}

	oauthToken, err := config.Exchange(context.Background(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to exchange token")
	}

	client := config.Client(context.Background(), oauthToken)

	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to get user info")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(userResp.Body)

	var guser domain.GithubUser
	if err := json.NewDecoder(userResp.Body).Decode(&guser); err != nil {
		log.Printf("Failed to parse user info: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to parse user info")
	}

	if guser.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			log.Printf("Failed to get user emails: %v", err)
		} else {
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(emailResp.Body)
			var emails []struct {
				Email      string `json:"email"`
				Primary    bool   `json:"primary"`
				Verified   bool   `json:"verified"`
				Visibility string `json:"visibility"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary {
						guser.Email = e.Email
						break
					}
				}
			}
		}
	}

	newUser := &domain.User{
		GithubID:  fmt.Sprintf("%d", guser.ID),
		Name:      guser.Name,
		Email:     guser.Email,
		AvatarURL: guser.AvatarURL,
	}

	user, err := uc.uu.FindOrCreateUserByGitHub(newUser)
	if err != nil {
		log.Printf("Failed to find or create user: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to find or create user")
	}

	sessionID := fmt.Sprintf("session_for_user_%d", user.ID)
	cookie := new(http.Cookie)
	cookie.Name = "session_id"
	cookie.Value = sessionID
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	cookie.Domain = os.Getenv("API_DOMAIN")
	cookie.Secure = true
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteNoneMode
	c.SetCookie(cookie)
	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173")
}

func (uc *userController) Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "session_id"
	cookie.Value = ""
	cookie.Expires = time.Now()
	cookie.MaxAge = -1
	cookie.Path = "/"
	cookie.Domain = os.Getenv("API_DOMAIN")
	cookie.Secure = true
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteNoneMode
	c.SetCookie(cookie)
	return c.String(http.StatusOK, "Logged out")
}
