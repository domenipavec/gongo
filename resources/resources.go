package resources

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

type PermissionManager interface {
	AddPermission(code, name string) error
}

type Resources struct {
	db                *gorm.DB
	permissionManager PermissionManager

	admin *admin.Admin
}

func New() *Resources {
	return &Resources{}
}

func (r *Resources) Configure(DB *gorm.DB, permissionManager PermissionManager) error {
	r.db = DB
	r.permissionManager = permissionManager

	r.admin = admin.New(&qor.Config{DB: DB})
	r.admin.SetAuth(&QorAuth{})

	return nil
}

func (r *Resources) ServeMux(prefix string) http.Handler {
	return r.admin.NewServeMux("/admin")
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

		if err := r.permissionManager.AddPermission(createPermissions[i], "Can create "+name); err != nil {
			return errors.Wrap(err, "could not add create permission")
		}
		if err := r.permissionManager.AddPermission(readPermissions[i], "Can read "+name); err != nil {
			return errors.Wrap(err, "could not add read permission")
		}
		if err := r.permissionManager.AddPermission(updatePermissions[i], "Can update "+name); err != nil {
			return errors.Wrap(err, "could not add update permission")
		}
		if err := r.permissionManager.AddPermission(deletePermissions[i], "Can delete "+name); err != nil {
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
