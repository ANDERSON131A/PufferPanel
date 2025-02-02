package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/v4/response"
	"github.com/pufferpanel/pufferpanel/v2"
	"github.com/pufferpanel/pufferpanel/v2/database"
	"github.com/pufferpanel/pufferpanel/v2/services"
	"net/http"
	"strings"
)

var noLogin = []string{"/auth/", "/error/", "/daemon/", "/api/"}
var assetFiles = []string{".js", ".css", ".img", ".ico", ".png", ".gif"}
var overrideRequireLogin = []string{"/auth/reauth"}

const WWWAuthenticateHeader = "WWW-Authenticate"
const WWWAuthenticateHeaderContents = "Bearer realm=\"\""

func AuthMiddleware(c *gin.Context) {
	for _, v := range noLogin {
		if strings.HasPrefix(c.Request.URL.Path, v) {
			//and now we see if it's actually one we override
			skip := false
			for _, o := range overrideRequireLogin {
				if o == c.Request.URL.Path {
					skip = true
					break
				}
			}
			if !skip {
				c.Next()
				return
			}
		}
	}

	cookie, err := c.Cookie("puffer_auth")

	if err != nil || cookie == "" {
		//determine if it's an asset, otherwise, we can redirect if it's a GET
		//dev only requirement?
		if c.Request.Method == "GET" && strings.Count(c.Request.URL.Path, "/") == 1 {
			for _, v := range assetFiles {
				if strings.HasSuffix(c.Request.URL.Path, v) {
					return
				}
			}
		}

		c.Header(WWWAuthenticateHeader, WWWAuthenticateHeaderContents)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	db, err := database.GetConnection()
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	token, err := services.ParseToken(cookie)

	if response.HandleError(c, err, http.StatusUnauthorized) {
		c.Header(WWWAuthenticateHeader, WWWAuthenticateHeaderContents)
		return
	}
	if !token.Valid {
		c.Header(WWWAuthenticateHeader, WWWAuthenticateHeaderContents)
		response.HandleError(c, pufferpanel.ErrTokenInvalid, http.StatusUnauthorized)
		return
	}

	us := services.User{DB: db}
	user, err := us.Get(token.Claims.Subject)
	if response.HandleError(c, err, http.StatusUnauthorized) {
		c.Header(WWWAuthenticateHeader, WWWAuthenticateHeaderContents)
		return
	}

	c.Set("user", user)
	c.Next()
}
