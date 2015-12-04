package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"time"
)

// User is the minimal interface a type reprenting a user must satisfy.
type User interface {
	Object
}

// ParseUser is a type that should be embedded in any custom User object.
type ParseUser struct {
	ParseObject
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	SessionToken  string `json:"sessionToken,omitempty"`
	Email         string `json:"email,omitempty"`
	EmailVerified string `json:"emailVerified,omitempty"`

	AuthData *authData `json:"authData,omitempty"`
}

type authData struct {
	Anonymous *struct {
		ID string `json:"id,omitempty"`
	} `json:"anonymous,omitempty"`
	Facebook *struct {
		AccessToken    string `json:"access_token,omitempty"`
		ExpirationDate string `json:"expiration_date,omitempty"`
		ID             string `json:"id,omitempty"`
	} `json:"facebook,omitempty"`
	Twitter *struct {
		AuthToken       string `json:"auth_token,omitempty"`
		AuthTokenSecret string `json:"auth_token_secret,omitempty"`
		ConsumerKey     string `json:"consumer_key,omitempty"`
		ConsumerSecret  string `json:"consumer_secret,omitempty"`
		ID              string `json:"id,omitempty"`
		ScreenName      string `json:"screen_name,omitempty"`
	} `json:"twitter,omitempty"`
}

// CreateUser creates a user from the specified object. On success the new user is
// returned. The provided object is not modified.
func (c *Client) CreateUser(user User) (*ParseUser, error) {
	payload, err := json.Marshal(user)
	c.trace("CreateUser >", "/1/users", string(payload))
	resp, err := c.doWithBody("POST", "/1/users", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var u *ParseUser
	err = json.Unmarshal(body, &u)
	c.trace("CreateUser <", "/1/users", string(body))
	return u, err
}

// LoginUser attempts to log in a user given the provided name an password.
// The provided object is populated with the user fields.
func (c *Client) LoginUser(username, password string, user User) error {
	uri, _ := url.Parse("/1/login")
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	uri.RawQuery = params.Encode()

	resp, err := c.doSimple("GET", uri.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.trace("LoginUser", uri, string(body))
	// TODO(tmc): warn if not == .Zero() before populating?
	return json.Unmarshal(body, user)
}

// GetUser looks up a user by ID. The provided user is populated on success.
func (c *Client) GetUser(userID string) (*User, error) {
	uri := fmt.Sprintf("/1/users/%s", userID)
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.trace("GetUser", uri, string(body))
	var user *User
	// TODO(tmc): warn if not == .Zero() before populating?
	return user, json.Unmarshal(body, &user)
}

// CurrentUser looks up the user associated with the provided credentials. The provided user is populated on success.
func (c *Client) CurrentUser(user User) error {
	uri := fmt.Sprintf("/1/users/me")
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.trace("CurrentUser", uri, string(body))
	// TODO(tmc): warn if not == .Zero() before populating?
	return json.Unmarshal(body, &user)
}

// UpdateUser updates the provided user with any provided fields and on success
// returns the updated at time.
func (c *Client) UpdateUser(user User) (updateTime time.Time, err error) {
	payload, err := json.Marshal(user)
	uri := fmt.Sprintf("/1/users/%s", user.ObjectID())
	resp, err := c.doWithBody("PUT", uri, bytes.NewReader(payload))
	log.Println("OI", string(payload))
	c.trace("UpdateUser >", uri, string(payload))
	if err != nil {
		return updateTime, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return updateTime, err
	}
	c.trace("UpdateUser < ", uri, string(body))
	updatedAt := &struct {
		Time time.Time `json:"updatedAt"`
	}{}
	err = json.Unmarshal(body, updatedAt)
	return updatedAt.Time, err
}

// DeleteUser deletes the provided user.
func (c *Client) DeleteUser(user User) error {
	uri := fmt.Sprintf("/1/users/%s", user.ObjectID())
	resp, err := c.doSimple("DELETE", uri)
	defer resp.Body.Close()
	c.trace("DeleteUser", uri)
	return err
}

// PasswordResetRequest sends a password reset email to the provided email address.
func (c *Client) PasswordResetRequest(email string) error {
	payload, err := json.Marshal(struct {
		Email string `json:"email"`
	}{Email: email})
	if err != nil {
		return err
	}
	uri := "/1/requestPasswordReset"
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.trace("PasswordResetRequest", uri, string(payload))
	return err
}
