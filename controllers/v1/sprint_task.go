package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retroServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// SprintTaskController ...
type SprintTaskController struct {
	SprintTaskService retroServices.SprintTaskService
	PermissionService retroServices.PermissionService
	TrailService      retroServices.TrailService
}

// Routes for Tasks
func (ctrl SprintTaskController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.GET("/:sprintTaskID/", ctrl.Get)
	r.PATCH("/:sprintTaskID/", ctrl.Update)
	r.POST("/:sprintTaskID/done/", ctrl.MarkDone)
	r.DELETE("/:sprintTaskID/done/", ctrl.MarkUndone)
}

// List ...
func (ctrl SprintTaskController) List(c *gin.Context) {
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	tasks, status, err := ctrl.SprintTaskService.List(retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, tasks)
}

// Get ...
func (ctrl SprintTaskController) Get(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanAccessSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, status, err := ctrl.SprintTaskService.Get(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// Update ...
func (ctrl SprintTaskController) Update(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	var data retroSerializers.SprintTaskUpdate
	err := c.BindJSON(&data)

	task, status, err := ctrl.SprintTaskService.Update(id, retroID, sprintID, data)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// MarkDone ...
func (ctrl SprintTaskController) MarkDone(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, status, err := ctrl.SprintTaskService.MarkDone(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}

// MarkUndone ...
func (ctrl SprintTaskController) MarkUndone(c *gin.Context) {
	id := c.Param("sprintTaskID")
	retroID := c.Param("retroID")
	sprintID := c.Param("sprintID")
	userID, _ := c.Get("userID")

	if !ctrl.PermissionService.UserCanEditSprintTask(retroID, sprintID, id, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	task, status, err := ctrl.SprintTaskService.MarkUndone(id, retroID, sprintID)

	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, task)
}
