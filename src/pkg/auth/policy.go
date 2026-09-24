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
