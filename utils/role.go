package utils

import (
	"github/OfrenDialsa/go-gin-starter/lib"
	"slices"
)

func ContainsRole(userRole lib.Role, allowedRoles []lib.Role) bool {
	return slices.Contains(allowedRoles, userRole)
}
