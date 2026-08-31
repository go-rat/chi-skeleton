package service

import (
	"net/http"

	"github.com/libtnb/chi-skeleton/internal/shared/transport"
	"github.com/libtnb/chi-skeleton/internal/user/biz"
)

func UserRoutes(user *UserService) transport.Endpoints {
	return transport.Endpoints{
		{Method: http.MethodGet, Path: "/users", Handler: user.List,
			Summary: "List users", Tags: []string{"user"},
			Document: transport.Describe[transport.Paginate, transport.Envelope[transport.Page[*biz.User]]](http.StatusOK)},
		{Method: http.MethodPost, Path: "/users", Handler: user.Create,
			Summary: "Create a user", Tags: []string{"user"},
			Document: transport.Describe[UserAdd, transport.Envelope[biz.User]](http.StatusOK)},
		{Method: http.MethodGet, Path: "/users/{id}", Handler: user.Get,
			Summary: "Get a user", Tags: []string{"user"},
			Document: transport.Describe[UserID, transport.Envelope[biz.User]](http.StatusOK)},
		{Method: http.MethodPut, Path: "/users/{id}", Handler: user.Update,
			Summary: "Update a user", Tags: []string{"user"},
			Document: transport.Describe[UserUpdate, transport.Envelope[biz.User]](http.StatusOK)},
		{Method: http.MethodDelete, Path: "/users/{id}", Handler: user.Delete,
			Summary: "Delete a user", Tags: []string{"user"},
			Document: transport.DescribeNoBody[UserID](http.StatusOK)},
	}
}
