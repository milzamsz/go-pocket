package server

import (
	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/kanban"
	"github.com/milzamsz/go-pocket/internal/services/products"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
)

type Deps struct {
	Config   config.Config
	Auth     auth.Service
	Billing  billing.Service
	Email    email.Service
	Tenancy  tenancy.Service
	Products products.Service
	Kanban   kanban.Service
}
