package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  redirectURL,
	}
}

func (uc *userController) GitHubLogin(c echo.Context) error {
	state := "state"
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (uc *userController) GitHubCallback(c echo.Context) error {
	state := c.QueryParam("state")
	code := c.QueryParam("code")

	if state != "state" {
		return c.String(http.StatusBadRequest, "Invalid state parameter")
	}

	oauthToken, err := config.Exchange(context.Background(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to exchange token")
	}

	client := config.Client(context.Background(), oauthToken)

	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get user info")
	}
	defer userResp.Body.Close()

	var guser domain.GithubUser
	if err := json.NewDecoder(userResp.Body).Decode(&guser); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse user info")
	}

	if guser.Email == "" {
		emailResp, err := client.Get("https://api.gihub.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
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
		return c.String(http.StatusInternalServerError, "Failed to find or create user")
	}

	sessionID := fmt.Sprintf("session_for_user_%d", user.ID)
	c.SetCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/",
	})

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173")
}

func (uc *userController) Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	return c.String(http.StatusOK, "Logged out")
}
