package main

// RoleService determines the roles for a given User
type RoleService struct {
	RoleBase     string     // BaseURI for roles
	DefaultRoles []string   // Default roles that all users have
	Cache        *RoleCache // Can be nil if no caching is desired
}

// Lookup looks up roles for a given user
func (r RoleService) Lookup(u *User) ([]Role, error) {

	generator := func() ([]Role, error) {
		var roles []Role
		for _, defaultRole := range r.DefaultRoles {
			roles = append(roles, Role{
				Base: r.RoleBase,
				Name: defaultRole,
			})
		}
		return roles, nil
	}

	if r.Cache != nil {
		return r.Cache.GetOrAdd(u.ID, generator)
	}

	return generator()
}
