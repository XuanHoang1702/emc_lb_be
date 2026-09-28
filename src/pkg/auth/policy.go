package auth

type UserContext struct {
	UserID string
	Role   string
}

// CanManageOrder evaluates if the user is an admin or the owner of the order.
func CanManageOrder(u UserContext, orderOwnerID string) bool {
	if u.Role == "admin" || u.Role == "superadmin" {
		return true
	}

	if GetRBACManager().HasPermission(u.Role, "manage_orders") {
		return true
	}

	return u.UserID == orderOwnerID
}

// GetRoleLevel returns an integer representing role privilege level for hierarchy comparisons
func GetRoleLevel(role string) int {
	switch role {
	case "superadmin":
		return 100
	case "admin":
		return 50
	case "shop_owner":
		return 20
	case "shop_staff":
		return 10
	case "customer":
		return 0
	default:
		return 0
	}
}

// CanManageUser evaluates if the caller can manage the target user's account (like changing password).
// A caller can only manage users with a strictly LOWER privilege level than themselves.
func CanManageUser(callerRole, targetRole string) bool {
	return GetRoleLevel(callerRole) > GetRoleLevel(targetRole)
}
