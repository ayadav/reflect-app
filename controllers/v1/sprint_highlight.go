package v1

import (
	"fmt"
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"net/http"

	"github.com/gin-gonic/gin"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
)

// SprintHighlightController ...
type SprintHighlightController struct {
	RetrospectiveFeedbackService retrospectiveServices.RetrospectiveFeedbackService
	PermissionService            retrospectiveServices.PermissionService
	TrailService                 retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintHighlightController) Routes(r *gin.RouterGroup) {
	r.POST("/", ctrl.Add)
	r.GET("/", ctrl.List)
	r.PUT("/:highlightID/", ctrl.Update)
}

// Add Highlight to sprint's retrospective
func (ctrl SprintHighlightController) Add(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	feedbackData := serializers.RetrospectiveFeedbackCreateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.Add(
		userID.(uint),
		sprintID,
		retroID,
		models.HighlightType,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to create highlight",
			"error":   err.Error()})
		return
	}

	ctrl.TrailService.Add("Added Highlight", "Retrospective Feedback",
		fmt.Sprint(response.ID),
		userID.(uint))
	c.JSON(http.StatusCreated, response)
}

// List highlights associated to sprint
func (ctrl SprintHighlightController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.List(
		userID.(uint),
		sprintID,
		retroID,
		models.HighlightType)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to fetch highlights",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Update highlight associated to a sprint
func (ctrl SprintHighlightController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	highlightID := c.Param("highlightID")
	feedbackData := serializers.RetrospectiveFeedbackUpdateSerializer{}

	if err := c.BindJSON(&feedbackData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	if !ctrl.PermissionService.CanAccessRetrospectiveFeedback(sprintID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, err := ctrl.RetrospectiveFeedbackService.Update(
		userID.(uint),
		retroID,
		highlightID,
		&feedbackData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to update highlight",
			"error":   err.Error()})
		return
	}

	ctrl.TrailService.Add("Updated Highlight", "Retrospective Feedback",
		highlightID,
		userID.(uint))

	c.JSON(http.StatusOK, response)
}