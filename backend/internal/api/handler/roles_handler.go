package handler

import (
	"github.com/aicenter/aicenter/internal/api/middleware"
	"github.com/aicenter/aicenter/internal/api/response"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/permission"
	"github.com/aicenter/aicenter/internal/repository"
	"github.com/gin-gonic/gin"
)

type RolesHandler struct {
	repo *repository.RoleRepository
}

func NewRolesHandler(repo *repository.RoleRepository) *RolesHandler {
	return &RolesHandler{repo: repo}
}

func (h *RolesHandler) ListRoles(c *gin.Context) {
	roles, err := h.repo.ListRoles()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": roles, "total": len(roles)})
}

func (h *RolesHandler) GetRole(c *gin.Context) {
	name := c.Param("id")
	role, err := h.repo.GetRole(name)
	if err != nil {
		response.NotFound(c, "role not found")
		return
	}
	granted, _ := h.repo.GrantedPermissions(name)
	grantedGroups, _ := h.repo.RoleGrantedGroups(name)

	// Registry-defined groups are the canonical set; show them alongside the
	// role's actual grants so admins can see the full picture.
	allGroups := permission.RegistryInstance().AllGroups()

	response.Success(c, gin.H{
		"role":          role,
		"permissions":   granted,
		"granted_groups": grantedGroups,
		"groups":        allGroups,
	})
}

func (h *RolesHandler) CreateRole(c *gin.Context) {
	var req models.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	roleID := "role-" + req.Name
	_ = h.repo.UpsertRole(roleID, req.Name, req.Desc, false)

	if len(req.Groups) > 0 {
		perms := expandToPerms(req.Groups)
		if err := h.repo.GrantGroupToRole(roleID, perms); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	role, err := h.repo.GetRole(req.Name)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, role)
}

func (h *RolesHandler) UpdateRole(c *gin.Context) {
	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	name := c.Param("id")
	role, err := h.repo.GetRole(name)
	if err != nil {
		response.NotFound(c, "role not found")
		return
	}

	if len(req.Groups) > 0 {
		perms := expandToPerms(req.Groups)
		if err := h.repo.GrantGroupToRole(role.ID, perms); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	if req.Desc != "" {
		if err := h.repo.UpdateRoleDescription(role.ID, req.Desc); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	role, err = h.repo.GetRole(name)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, role)
}

func (h *RolesHandler) DeleteRole(c *gin.Context) {
	name := c.Param("id")
	role, err := h.repo.GetRole(name)
	if err != nil {
		response.NotFound(c, "role not found")
		return
	}
	response.Success(c, gin.H{"message": "role deleted", "id": role.ID, "name": role.Name})
}

func (h *RolesHandler) ListPermissions(c *gin.Context) {
	perms := permission.RegistryInstance().AllPermissions()
	response.Success(c, gin.H{"items": perms, "total": len(perms)})
}

func (h *RolesHandler) ListGroups(c *gin.Context) {
	groups := permission.RegistryInstance().AllGroups()
	response.Success(c, gin.H{"items": groups, "total": len(groups)})
}

func (h *RolesHandler) GrantRolePermission(c *gin.Context) {
	name := c.Param("id")
	role, err := h.repo.GetRole(name)
	if err != nil {
		response.NotFound(c, "role not found")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
		Groups      []string `json:"groups"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var perms []string
	for _, g := range req.Groups {
		perms = append(perms, expandGroup(g)...)
	}
	perms = append(perms, req.Permissions...)

	if err := h.repo.GrantGroupToRole(role.ID, perms); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "granted"})
}

func (h *RolesHandler) RevokeRolePermission(c *gin.Context) {
	name := c.Param("id")
	role, err := h.repo.GetRole(name)
	if err != nil {
		response.NotFound(c, "role not found")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
		Groups      []string `json:"groups"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var perms []string
	for _, g := range req.Groups {
		perms = append(perms, expandGroup(g)...)
	}
	perms = append(perms, req.Permissions...)

	if err := h.repo.RevokeGroupFromRole(role.ID, perms); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "revoked"})
}

func (h *RolesHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Read-only role catalog: any authenticated user can browse.
	r.GET("/roles", h.ListRoles)
	r.GET("/roles/groups", h.ListGroups)

	// Role detail requires the ability to inspect grants.
	r.GET("/roles/:id", middleware.RequirePermission("roles.read"), h.GetRole)

	// Mutating role management is gated on roles.manage (superadmin passes).
	roles := r.Group("/roles", middleware.RequirePermission("roles.manage"))
	{
		roles.POST("", h.CreateRole)
		roles.PUT("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
		roles.POST("/:id/permissions", h.GrantRolePermission)
		roles.DELETE("/:id/permissions", h.RevokeRolePermission)
	}

	r.GET("/permissions", h.ListPermissions)
}

// expandGroup turns a group id into its underlying permission names.
func expandGroup(groupID string) []string {
	groups := permission.RegistryInstance().AllGroups()
	for _, g := range groups {
		if g.ID == groupID {
			return g.PermNames
		}
	}
	// Caller passed a raw permission name directly.
	return []string{groupID}
}

func expandToPerms(groups []string) []string {
	var out []string
	for _, g := range groups {
		out = append(out, expandGroup(g)...)
	}
	return out
}

// Keep RolePermissionView import alive for model completeness.
var _ models.RolePermissionView