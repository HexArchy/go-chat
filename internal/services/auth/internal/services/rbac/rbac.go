package rbac

import (
	"strings"
	"sync"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type RBAC struct {
	policies map[entities.Role]map[string][]string
	mu       sync.RWMutex
}

func NewRBAC() *RBAC {
	rbac := &RBAC{
		policies: make(map[entities.Role]map[string][]string),
	}

	// Define default policies
	rbac.AddPolicy(entities.RoleUser, "read", "profile")
	rbac.AddPolicy(entities.RoleUser, "update", "own_profile")
	rbac.AddPolicy(entities.RoleModerator, "read", "all_profiles")
	rbac.AddPolicy(entities.RoleModerator, "update", "all_profiles")
	rbac.AddPolicy(entities.RoleAdmin, "read", "*")
	rbac.AddPolicy(entities.RoleAdmin, "update", "*")
	rbac.AddPolicy(entities.RoleAdmin, "delete", "*")

	return rbac
}

func (r *RBAC) AddPolicy(role entities.Role, action, resource string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.policies[role]; !ok {
		r.policies[role] = make(map[string][]string)
	}
	r.policies[role][action] = append(r.policies[role][action], resource)
}

func (r *RBAC) RemovePolicy(role entities.Role, action, resource string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.policies[role]; ok {
		if resources, ok := r.policies[role][action]; ok {
			for i, res := range resources {
				if res == resource {
					r.policies[role][action] = append(resources[:i], resources[i+1:]...)
					break
				}
			}
		}
	}
}

func (r *RBAC) CheckPermission(roles []entities.Role, action, resource string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, role := range roles {
		if r.checkRolePermission(role, action, resource) {
			return true
		}
	}
	return false
}

func (r *RBAC) checkRolePermission(role entities.Role, action, resource string) bool {
	if resources, ok := r.policies[role][action]; ok {
		for _, res := range resources {
			if res == "*" || res == resource || (strings.HasSuffix(res, "*") && strings.HasPrefix(resource, res[:len(res)-1])) {
				return true
			}
		}
	}
	return false
}

func (r *RBAC) IsValidRole(role entities.Role) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.policies[role]
	return ok
}

func (r *RBAC) GetAllRoles() []entities.Role {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := make([]entities.Role, 0, len(r.policies))
	for role := range r.policies {
		roles = append(roles, role)
	}
	return roles
}
