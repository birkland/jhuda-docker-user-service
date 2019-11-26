package main

type RoleService struct {
	RoleBase     string
	DefaultRoles []string
}

func (r RoleService) Lookup(u *User) ([]Role, error) {
	var roles []Role

	for _, defaultRole := range r.DefaultRoles {
		roles = append(roles, Role{
			Base: r.RoleBase,
			Name: defaultRole,
		})
	}

	return roles, nil
}
