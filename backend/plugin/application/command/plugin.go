package command

import pluginiface "goadmin/plugin/interface"

type CreatePlugin struct {
	Name        string
	Description string
	Enabled     bool
	Menus       []pluginiface.Menu
	Permissions []pluginiface.Permission
}

type UpdatePlugin struct {
	Description *string
	Enabled     *bool
	Menus       []pluginiface.Menu
	Permissions []pluginiface.Permission
}
