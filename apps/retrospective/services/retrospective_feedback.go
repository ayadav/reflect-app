package services

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strconv"
)

// RetrospectiveFeedbackService ...
type RetrospectiveFeedbackService struct {
	DB *gorm.DB
}

// Add ...
func (service RetrospectiveFeedbackService) Add(userID uint, sprintID string, retroID string,
	feedbackType models.RetrospectiveFeedbackType,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackCreateSerializer) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	error) {
	db := service.DB

	retroIDInt, err := strconv.Atoi(retroID)
	if err != nil {
		return nil, errors.New("invalid retrospective id")
	}
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	retroFeedback := models.RetrospectiveFeedback{
		RetrospectiveID: uint(retroIDInt),
		SubType:         feedbackData.SubType,
		Type:            feedbackType,
		AddedAt:         sprint.StartDate,
		CreatedByID:     userID,
		AssigneeID:      nil,
		ExpectedAt:      nil,
		ResolvedAt:      nil,
	}

	if feedbackType != models.GoalType {
		retroFeedback.ResolvedAt = sprint.EndDate
	}

	err = db.Create(&retroFeedback).Error
	if err != nil {
		return nil, err
	}

	return service.getRetrospectiveFeedback(retroFeedback.ID)

}

// Update ...
func (service RetrospectiveFeedbackService) Update(userID uint, retroID string,
	feedbackID string,
	feedbackData *retrospectiveSerializers.RetrospectiveFeedbackUpdateSerializer) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	error) {
	db := service.DB

	retroFeedback := models.RetrospectiveFeedback{}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", feedbackID).
		First(&retroFeedback).Error; err != nil {
		return nil, err
	}

	if retroFeedback.Type == models.GoalType && retroFeedback.ResolvedAt != nil {
		return nil, errors.New("can not updated resolved goal")
	}

	if feedbackData.Scope != nil {
		retroFeedback.Scope = models.RetrospectiveFeedbackScope(*feedbackData.Scope)
	}

	if feedbackData.Text != nil {
		retroFeedback.Text = *feedbackData.Text
	}

	if feedbackData.ExpectedAt != nil {
		if retroFeedback.Type != models.GoalType {
			return nil, errors.New("expectedAt can be updated only for goal " +
				"type retrospective feedback")
		}
		retroFeedback.ExpectedAt = feedbackData.ExpectedAt
	}

	retroFeedback.AssigneeID = feedbackData.AssigneeID

	err := db.Save(&retroFeedback).Error
	if err != nil {
		return nil, err
	}

	return service.getRetrospectiveFeedback(retroFeedback.ID)
}

// Resolve ...
func (service RetrospectiveFeedbackService) Resolve(userID uint, sprintID string, retroID string,
	feedbackID string,
	markResolved bool) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	error) {
	db := service.DB

	retroFeedback := models.RetrospectiveFeedback{}

	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", feedbackID).
		First(&retroFeedback).Error; err != nil {
		return nil, err
	}

	if retroFeedback.Type != models.GoalType {
		return nil, errors.New("only goal typed retrospective feedback could" +
			" be resolved or unresolved")
	}

	if markResolved && retroFeedback.ResolvedAt == nil {
		retroFeedback.ResolvedAt = sprint.EndDate
	}

	if !markResolved && retroFeedback.ResolvedAt != nil {
		retroFeedback.ResolvedAt = nil
	}

	err := db.Save(&retroFeedback).Error
	if err != nil {
		return nil, err
	}

	return service.getRetrospectiveFeedback(retroFeedback.ID)
}

// List ...
func (service RetrospectiveFeedbackService) List(userID uint, sprintID string, retroID string,
	feedbackType models.RetrospectiveFeedbackType) (
	feedbackList *retrospectiveSerializers.RetrospectiveFeedbackListSerializer,
	err error) {
	db := service.DB
	feedbackList = new(retrospectiveSerializers.RetrospectiveFeedbackListSerializer)
	sprint := models.Sprint{}

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&models.RetrospectiveFeedback{}).
		Where("retrospective_id = ? AND type = ?", retroID, feedbackType).
		Where("added_at >= ? AND added_at <= ?", *sprint.StartDate, *sprint.EndDate).
		Preload("Assignee").
		Preload("CreatedBy").
		Find(&feedbackList.Feedbacks).Error; err != nil {
		return nil, err
	}

	return feedbackList, nil
}

// ListGoal ...
func (service RetrospectiveFeedbackService) ListGoal(userID uint, sprintID string,
	retroID string, goalType string) (
	feedbackList *retrospectiveSerializers.RetrospectiveFeedbackListSerializer,
	err error) {
	db := service.DB

	sprint := models.Sprint{}
	feedbackList = new(retrospectiveSerializers.RetrospectiveFeedbackListSerializer)

	if err := db.Model(&models.Sprint{}).
		Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, err
	}

	query := db.Model(&models.RetrospectiveFeedback{}).
		Where("retrospective_id = ? AND type = ?", retroID, models.GoalType)

	switch goalType {
	case "added":
		query = query.Where("resolved_at IS NULL").
			Where("added_at >= ? AND added_at <= ?", sprint.StartDate, sprint.EndDate)
	case "completed":
		query = query.
			Where("resolved_at >= ? AND resolved_at <= ?", sprint.StartDate, sprint.EndDate)
	case "pending":
		query = query.
			Where("resolved_at IS NULL").
			Where("added_at < ?", sprint.EndDate)
	default:
		return nil, errors.New("invalid goal type")
	}

	if err := query.
		Preload("Assignee").
		Preload("CreatedBy").
		Find(&feedbackList.Feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbackList, nil
}

func (service RetrospectiveFeedbackService) getRetrospectiveFeedback(retroFeedbackID uint) (
	*retrospectiveSerializers.RetrospectiveFeedback,
	error) {
	db := service.DB
	feedback := retrospectiveSerializers.RetrospectiveFeedback{}
	err := db.Model(&models.RetrospectiveFeedback{}).
		Where("id = ?", retroFeedbackID).
		Preload("CreatedBy").
		Preload("Assignee").
		First(&feedback).Error
	if err != nil {
		return nil, err
	}

	return &feedback, nil

}
