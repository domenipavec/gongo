package admin

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/matematik7/gongo"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

type Admin struct {
	qor *admin.Admin

	prefix string
}

func New(prefix string) *Admin {
	return &Admin{
		prefix: prefix,
	}
}

func (a *Admin) Configure(app gongo.App) error {
	DB := app["DB"].(*gorm.DB)

	a.qor = admin.New(&qor.Config{DB: DB})
	a.qor.SetAuth(&QorAuth{})

	for group, itf := range app {
		if resourcer, ok := itf.(gongo.Resourcer); ok {
			models := resourcer.Resources()

			// TODO: generating permission names should be part of authorization
			createPermissions := make([]string, len(models))
			readPermissions := make([]string, len(models))
			updatePermissions := make([]string, len(models))
			deletePermissions := make([]string, len(models))
			for i, model := range models {
				name := DB.NewScope(model).TableName()
				createPermissions[i] = "create_" + name
				readPermissions[i] = "read_" + name
				updatePermissions[i] = "update_" + name
				deletePermissions[i] = "delete_" + name
			}

			a.qor.AddMenu(&admin.Menu{Name: group, Permission: roles.Allow(roles.Read, readPermissions...)})
			for i, model := range models {
				a.qor.AddResource(model, &admin.Config{
					Menu: []string{group},
					Permission: roles.Allow(
						roles.Read, readPermissions[i],
					).Allow(
						roles.Create, createPermissions[i],
					).Allow(
						roles.Update, updatePermissions[i],
					).Allow(
						roles.Delete, deletePermissions[i],
					),
				})
			}
		}
	}

	return nil
}

func (a *Admin) ServeMux() http.Handler {
	return a.qor.NewServeMux(a.prefix)
}

type QorAuth struct{}

func (QorAuth) LoginURL(c *admin.Context) string {
	// TODO: fix urls
	return "/login"
}

func (QorAuth) LogoutURL(c *admin.Context) string {
	return "/logout"
}

func (QorAuth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	// TODO: this should be part of authorzation
	intf := c.Request.Context().Value("user")
	if intf == nil {
		return nil
	}
	return intf.(qor.CurrentUser)
}
