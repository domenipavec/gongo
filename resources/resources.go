package resources

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/matematik7/gongo"
	"github.com/pkg/errors"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

type Resources struct {
	db            *gorm.DB
	authorization gongo.Authorization

	admin *admin.Admin

	prefix string
}

func New(prefix string) *Resources {
	return &Resources{
		prefix: prefix,
	}
}

func (r *Resources) Configure(app gongo.App) error {
	r.db = app.DB
	r.authorization = app.Authorization

	r.admin = admin.New(&qor.Config{DB: r.db})
	r.admin.SetAuth(&QorAuth{})

	return nil
}

func (r *Resources) ServeMux() http.Handler {
	return r.admin.NewServeMux(r.prefix)
}

func (r Resources) Resources() []interface{} {
	return nil
}

func (r Resources) Name() string {
	return "Resources"
}

func (r *Resources) Register(group string, models ...interface{}) error {
	if err := r.db.AutoMigrate(models...).Error; err != nil {
		return errors.Wrap(err, "could not auto migrate models")
	}

	createPermissions := make([]string, len(models))
	readPermissions := make([]string, len(models))
	updatePermissions := make([]string, len(models))
	deletePermissions := make([]string, len(models))
	for i, model := range models {
		name := r.db.NewScope(model).TableName()
		createPermissions[i] = "create_" + name
		readPermissions[i] = "read_" + name
		updatePermissions[i] = "update_" + name
		deletePermissions[i] = "delete_" + name

		if err := r.authorization.AddPermission(createPermissions[i], "Can create "+name); err != nil {
			return errors.Wrap(err, "could not add create permission")
		}
		if err := r.authorization.AddPermission(readPermissions[i], "Can read "+name); err != nil {
			return errors.Wrap(err, "could not add read permission")
		}
		if err := r.authorization.AddPermission(updatePermissions[i], "Can update "+name); err != nil {
			return errors.Wrap(err, "could not add update permission")
		}
		if err := r.authorization.AddPermission(deletePermissions[i], "Can delete "+name); err != nil {
			return errors.Wrap(err, "could not add delete permission")
		}
	}

	r.admin.AddMenu(&admin.Menu{Name: group, Permission: roles.Allow(roles.Read, readPermissions...)})
	for i, model := range models {
		r.admin.AddResource(model, &admin.Config{
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

	return nil
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
	intf := c.Request.Context().Value("user")
	if intf == nil {
		return nil
	}
	return intf.(qor.CurrentUser)
}
