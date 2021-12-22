package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type (
	user struct {
		ID   	 int    `json:"id"`
		UserInfo userInfo `json:"userInfo"`
		Posts []post `json:"posts"`
	}
)

type (
	userInfo struct {
		Name   	 string  `json:"name"`
		Username string  `json:"username"`
		Email	 string	 `json:"email"`
	}
)

type (
	post struct {
		ID float64 `json:"id"`
		Title string `json:"title"`
		Body string `json:"body"`
	}
)

//declare the client
var (
	tr = &http.Transport{}
	client = &http.Client{Transport: tr}
)

//----------
// Handlers
//----------

func main() {
	e := echo.New()

	// Routes
	e.GET("/v1/user-posts/:id", getUserPosts)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func getUserPosts(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// TODO run the two calls concurrently
	var userErr, info = getUser(c)
	//TODO check actual response code and don't return a 500 every time
	if userErr != nil {
		return c.JSON(http.StatusInternalServerError, userErr)
	}

	var postError, userPosts = getPostsByUserId(c)
	if postError != nil {
		return c.JSON(http.StatusInternalServerError, userErr)
	}

	var userResp = user {
		ID:       id,
		UserInfo: info,
		Posts:    userPosts,
	}

	return c.JSON(http.StatusOK, userResp)
}

func getUser(c echo.Context) (error, userInfo) {
	id := c.Param("id")
	url := "https://jsonplaceholder.typicode.com/users/" + id
	resp, err := client.Get(url)

	if err != nil {
		return err, userInfo{}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// avoid creating a big model - just get the json we care about
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if len(result) == 0 {
		return errors.New("Not Found"), userInfo{}
	}

	// TODO fail gracefully - if these fields aren't found we will fail
	return nil, userInfo {
		Name: result["name"].(string),
		Username: result["username"].(string),
		Email: result["email"].(string),
	}
}

func getPostsByUserId(c echo.Context) (error, []post) {
	id := c.Param("id")
	req, err := http.NewRequest("GET","https://jsonplaceholder.typicode.com/posts", nil)
	params := req.URL.Query()
	params.Add("userId", id)
	req.URL.RawQuery = params.Encode()

	resp, err := client.Do(req)

	if err != nil {
		return err, []post{}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// model is smaller, so this doesn't save as much time
	var result []map[string]interface{}
	json.Unmarshal(body, &result)

	if len(result) == 0 {
		return errors.New("Not Found"), []post{}
	}

	var userPosts []post
	for _, elem := range result {
		//TODO Like the user, if these fields are wrongly typed, we will fail
		userPosts = append(userPosts, post{
			ID: elem["id"].(float64),
			Title: elem["title"].(string),
			Body: elem["body"].(string),
		})
	}

	return nil, userPosts
}