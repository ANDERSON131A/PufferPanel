package models

import (
	"github.com/pufferpanel/apufferi/v4/scope"
	"reflect"
)

type Permissions struct {
	ID uint `gorm:"PRIMARY_KEY,AUTO_INCREMEMT" json:"-"`

	//owners of this permission set
	UserId *uint `json:"-"`
	User   User  `gorm:"ASSOCIATION_SAVE_REFERENCE:false" json:"-" validate:"-"`

	ClientId *uint  `json:"-"`
	Client   Client `gorm:"ASSOCIATION_SAVE_REFERENCE:false" json:"-" validate:"-"`

	//if this set is for a server, what server
	ServerIdentifier *string `json:"-"`
	Server           Server  `gorm:"ASSOCIATION_SAVE_REFERENCE:false" json:"-" validate:"-"`

	//and here are all the perms we support
	Admin           bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	ViewServer      bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	CreateServer    bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	ViewNodes       bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	EditNodes       bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	DeployNodes     bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	ViewTemplates   bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	EditUsers       bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	ViewUsers       bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	EditServerAdmin bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	DeleteServer    bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`

	//these only will exist if tied to a server, and for a user
	EditServerData    bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	EditServerUsers   bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	InstallServer     bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	UpdateServer      bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""` //this is unused currently
	ViewServerConsole bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	SendServerConsole bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	StopServer        bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	StartServer       bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	ViewServerStats   bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	ViewServerFiles   bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	SFTPServer        bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
	PutServerFiles    bool `gorm:"NOT NULL;DEFAULT:0" json:"-" oneOf:""`
}

type MultiplePermissions []*Permissions

func (p *Permissions) BeforeSave() {
	if p.ServerIdentifier != nil && *p.ServerIdentifier == "" {
		p.ServerIdentifier = nil
	}
}

func (p *Permissions) ToScopes() []scope.Scope {
	scopes := make([]scope.Scope, 0)

	if p.Admin {
		scopes = append(scopes, scope.ServersAdmin)

		if p.ServerIdentifier == nil {
			scopes = append(scopes, scope.ServersCreate, scope.NodesView, scope.NodesDeploy, scope.NodesEdit, scope.TemplatesView, scope.UsersView, scope.UsersEdit)
		} else {
			scopes = append(scopes, scope.ServersDelete, scope.ServersEditAdmin)
		}
	} else {
		if p.ServerIdentifier == nil {
			if p.CreateServer {
				scopes = append(scopes, scope.ServersCreate)
			}

			if p.ViewNodes {
				scopes = append(scopes, scope.NodesView)
			}

			if p.EditNodes {
				scopes = append(scopes, scope.NodesEdit)
			}

			if p.ViewTemplates {
				scopes = append(scopes, scope.TemplatesView)
			}

			if p.EditUsers {
				scopes = append(scopes, scope.UsersEdit)
			}

			if p.ViewUsers {
				scopes = append(scopes, scope.UsersView)
			}

			if p.DeployNodes {
				scopes = append(scopes, scope.NodesDeploy)
			}
		} else {
			if p.DeleteServer {
				scopes = append(scopes, scope.ServersDelete)
			}

			if p.EditServerAdmin {
				scopes = append(scopes, scope.ServersEditAdmin)
			}
		}
	}

	if p.ViewServer {
		scopes = append(scopes, scope.ServersView)
	}

	//these only apply if there is a server involved
	if p.ServerIdentifier != nil {
		if p.EditServerData {
			scopes = append(scopes, scope.ServersEdit)
		}

		if p.EditServerUsers {
			scopes = append(scopes, scope.ServersEditUsers)
		}

		if p.InstallServer {
			scopes = append(scopes, scope.ServersInstall)
		}

		if p.UpdateServer {
			scopes = append(scopes, scope.ServersUpdate)
		}

		if p.ViewServerConsole {
			scopes = append(scopes, scope.ServersConsole)
		}

		if p.SendServerConsole {
			scopes = append(scopes, scope.ServersConsoleSend)
		}

		if p.StopServer {
			scopes = append(scopes, scope.ServersStop)
		}

		if p.StartServer {
			scopes = append(scopes, scope.ServersStart)
		}

		if p.ViewServerStats {
			scopes = append(scopes, scope.ServersStat)
		}

		if p.ViewServerFiles {
			scopes = append(scopes, scope.ServersFilesGet)
		}

		if p.PutServerFiles {
			scopes = append(scopes, scope.ServersFilesPut)
		}

		if p.SFTPServer {
			scopes = append(scopes, scope.ServersSFTP)
		}
	}

	return scopes
}

func (p *Permissions) SetDefaults() {
	p.ViewServer = true

	if p.ServerIdentifier != nil {
		p.EditServerData = true
		p.EditServerUsers = true
		p.InstallServer = true
		p.UpdateServer = true
		p.ViewServerConsole = true
		p.SendServerConsole = true
		p.StopServer = true
		p.StartServer = true
		p.ViewServerStats = true
		p.ViewServerFiles = true
		p.SFTPServer = true
		p.PutServerFiles = true
	}
}

func (p *Permissions) ShouldDelete() bool {
	val := reflect.ValueOf(p)

	// If it's an interface or a pointer, unwrap it.
	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldKind := field.Kind()

		if fieldKind != reflect.Bool {
			continue
		}

		typeField := val.Type().Field(i)

		// Get the field tag value.
		_, exist := typeField.Tag.Lookup("oneOf")

		if !exist {
			continue
		}

		if field.Bool() {
			return false
		}
	}

	return true
}
