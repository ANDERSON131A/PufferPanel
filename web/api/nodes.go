/*
 Copyright 2018 Padduck, LLC
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 	http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/v4/response"
	"github.com/pufferpanel/apufferi/v4/scope"
	"github.com/pufferpanel/pufferpanel/v2"
	"github.com/pufferpanel/pufferpanel/v2/models"
	"github.com/pufferpanel/pufferpanel/v2/services"
	"github.com/pufferpanel/pufferpanel/v2/web/handlers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func registerNodes(g *gin.RouterGroup) {
	g.Handle("GET", "", handlers.OAuth2Handler(scope.NodesView, false), getAllNodes)
	g.Handle("OPTIONS", "", response.CreateOptions("GET"))
	g.Handle("POST", "", handlers.OAuth2Handler(scope.NodesEdit, false), createNode)

	g.Handle("GET", "/:id", handlers.OAuth2Handler(scope.NodesView, false), getNode)
	g.Handle("PUT", "/:id", handlers.OAuth2Handler(scope.NodesEdit, false), updateNode)
	g.Handle("DELETE", "/:id", handlers.OAuth2Handler(scope.NodesEdit, false), deleteNode)
	g.Handle("OPTIONS", "/:id", response.CreateOptions("PUT", "GET", "POST", "DELETE"))

	g.Handle("GET", "/:id/deployment", handlers.OAuth2Handler(scope.NodesDeploy, false), response.NotImplemented)
	g.Handle("OPTIONS", "/:id/deployment", response.CreateOptions("GET"))

	//g.Handle("POST", "/:id/reset", handlers.OAuth2(scope.NodesDeploy, false), response.NotImplemented)
	//g.Handle("OPTIONS", "/:id/reset", response.CreateOptions("POST"))
}

func getAllNodes(c *gin.Context) {
	var err error
	db := handlers.GetDatabase(c)
	ns := &services.Node{DB: db}

	var nodes *models.Nodes
	if nodes, err = ns.GetAll(); response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	data := models.FromNodes(nodes)

	c.JSON(http.StatusOK, data)
}

func getNode(c *gin.Context) {
	var err error
	db := handlers.GetDatabase(c)
	ns := &services.Node{DB: db}

	id, ok := validateId(c)
	if !ok {
		return
	}

	node, err := ns.Get(id)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	data := models.FromNode(node)

	c.JSON(http.StatusOK, data)
}

func createNode(c *gin.Context) {
	var err error
	db := handlers.GetDatabase(c)
	ns := &services.Node{DB: db}

	model := &models.NodeView{}
	if err = c.BindJSON(model); response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	if err = model.Valid(false); response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	create := &models.Node{}
	model.CopyToModel(create)
	create.Secret = strings.Replace(uuid.NewV4().String(), "-", "", -1)
	if err = ns.Create(create); response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.JSON(http.StatusOK, create)
}

func updateNode(c *gin.Context) {
	var err error
	db := handlers.GetDatabase(c)
	ns := &services.Node{DB: db}

	viewModel := &models.NodeView{}
	if err = c.BindJSON(viewModel); response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	id, ok := validateId(c)
	if !ok {
		return
	}

	if err = viewModel.Valid(true); response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	node, err := ns.Get(id)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	viewModel.CopyToModel(node)
	if err = ns.Update(node); response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.JSON(http.StatusOK, node)
}

func deleteNode(c *gin.Context) {
	var err error
	db := handlers.GetDatabase(c)
	ns := &services.Node{DB: db}

	id, ok := validateId(c)
	if !ok {
		return
	}

	node, err := ns.Get(id)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	err = ns.Delete(node.ID)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.Status(http.StatusNoContent)
}

func deployNode(c *gin.Context) {
	var err error
	db := handlers.GetDatabase(c)
	ns := &services.Node{DB: db}

	id, ok := validateId(c)
	if !ok {
		return
	}

	node, err := ns.Get(id)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	services.ValidateTokenLoaded()
	file, err := ioutil.ReadFile(viper.GetString("token.public"))
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	data := &Deployment{
		ClientId:     fmt.Sprintf(".node_%d", node.ID),
		ClientSecret: node.Secret,
		BaseUrl:      fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.URL.Host),
		PublicKey:    string(file),
	}

	c.JSON(http.StatusOK, data)
}

func validateId(c *gin.Context) (uint, bool) {
	param := c.Param("id")

	id, err := strconv.Atoi(param)

	if response.HandleError(c, err, http.StatusBadRequest) || id <= 0 {
		response.HandleError(c, pufferpanel.ErrFieldTooSmall("id", 0), http.StatusBadRequest)
		return 0, false
	}

	return uint(id), true
}

type Deployment struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	BaseUrl      string `json:"baseUrl"`
	PublicKey    string `json:"publicKey"`
}
