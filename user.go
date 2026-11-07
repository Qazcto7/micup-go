package micup

import (
	"context"
	"fmt"
)

// UserMe fetches the current authenticated user/bot profile.
func (c *RESTClient) UserMe(ctx context.Context) (*User, error) {
	var user User
	err := c.RequestJSON(ctx, "GET", "/users/@me", nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUser fetches a specific user by their ID.
func (c *RESTClient) GetUser(ctx context.Context, userID string) (*User, error) {
	var user User
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/users/%s", userID), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ModifyCurrentUser updates the current bot username or avatar.
type UserModify struct {
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"` // base64 encoded image data
}

// ModifyCurrentUser performs an update on the current bot user.
func (c *RESTClient) ModifyCurrentUser(ctx context.Context, data *UserModify) (*User, error) {
	var user User
	err := c.RequestJSON(ctx, "PATCH", "/users/@me", data, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
