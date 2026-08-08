package services

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"

	"gorm.io/gorm"
)

const (
	DeveloperResourceMessages             = "messages"
	DeveloperResourceScheduledTasks       = "scheduled-tasks"
	DeveloperResourceUserProfileSummaries = "user-profile-summaries"
	DeveloperResourceConversationArchives = "conversation-archives"
	DeveloperResourceUserSettings         = "user-settings"
	DeveloperResourceUsers                = "users"
	DeveloperResourceVoiceFiles           = "voice-files"
)

type ArchiveVectorStore interface {
	DeleteArchives(ctx context.Context, embeddingIDs []string, dimension int) error
}

type archiveVectorRef struct {
	EmbeddingID        string
	EmbeddingDimension int
}

type DeveloperDataService struct {
	archiveVectors ArchiveVectorStore
}

type DeveloperPage struct {
	Data  any   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

type DeveloperDashboard struct {
	Messages             int64       `json:"messages"`
	CompletedTasks       int64       `json:"completed_tasks"`
	WaitingTasks         int64       `json:"waiting_tasks"`
	CurrentUser          models.User `json:"current_user"`
	UserProfileSummaries int64       `json:"user_profile_summaries"`
	ConversationArchives int64       `json:"conversation_archives"`
	VoiceFiles           int64       `json:"voice_files"`
	VoiceFileBytes       int64       `json:"voice_file_bytes"`
}

func NewDeveloperDataService(archiveVectors ArchiveVectorStore) *DeveloperDataService {
	return &DeveloperDataService{archiveVectors: archiveVectors}
}

func (s *DeveloperDataService) Dashboard(userID int64) (*DeveloperDashboard, error) {
	dashboard := &DeveloperDashboard{}
	if err := database.DB.First(&dashboard.CurrentUser, userID).Error; err != nil {
		return nil, err
	}
	counts := []struct {
		model  any
		target *int64
	}{
		{&models.Message{}, &dashboard.Messages},
		{&models.UserProfileSummary{}, &dashboard.UserProfileSummaries},
		{&models.ConversationArchive{}, &dashboard.ConversationArchives},
		{&models.VoiceFile{}, &dashboard.VoiceFiles},
	}
	for _, count := range counts {
		if err := database.DB.Model(count.model).Count(count.target).Error; err != nil {
			return nil, err
		}
	}
	if err := database.DB.Model(&models.ScheduledTask{}).Where("status = ?", "completed").Count(&dashboard.CompletedTasks).Error; err != nil {
		return nil, err
	}
	if err := database.DB.Model(&models.ScheduledTask{}).Where("status = ?", "pending").Count(&dashboard.WaitingTasks).Error; err != nil {
		return nil, err
	}
	if err := database.DB.Model(&models.VoiceFile{}).Select("COALESCE(SUM(file_size), 0)").Scan(&dashboard.VoiceFileBytes).Error; err != nil {
		return nil, err
	}
	return dashboard, nil
}

func (s *DeveloperDataService) List(resource string, page, size int) (*DeveloperPage, error) {
	model, err := developerModel(resource)
	if err != nil {
		return nil, err
	}
	page, size = normalizeDeveloperPage(page, size)

	var total int64
	if err := database.DB.Model(model).Count(&total).Error; err != nil {
		return nil, err
	}

	data, err := developerListTarget(resource)
	if err != nil {
		return nil, err
	}
	if err := database.DB.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(data).Error; err != nil {
		return nil, err
	}

	return &DeveloperPage{Data: data, Total: total, Page: page, Size: size}, nil
}

func (s *DeveloperDataService) Get(resource string, id int64) (any, error) {
	model, err := developerModel(resource)
	if err != nil {
		return nil, err
	}
	if err := database.DB.First(model, id).Error; err != nil {
		return nil, err
	}
	return model, nil
}

func (s *DeveloperDataService) Create(resource string, values map[string]any) (any, error) {
	model, err := developerModel(resource)
	if err != nil {
		return nil, err
	}
	values = allowedDeveloperValues(resource, values)
	if len(values) == 0 {
		return nil, fmt.Errorf("at least one writable field is required")
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(payload, model); err != nil {
		return nil, err
	}
	if err := database.DB.Create(model).Error; err != nil {
		return nil, err
	}
	return s.Get(resource, developerModelID(model))
}

func (s *DeveloperDataService) Update(resource string, id int64, values map[string]any) (any, error) {
	model, err := developerModel(resource)
	if err != nil {
		return nil, err
	}
	values = allowedDeveloperValues(resource, values)
	if len(values) == 0 {
		return nil, fmt.Errorf("at least one writable field is required")
	}
	result := database.DB.Model(model).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return s.Get(resource, id)
}

func (s *DeveloperDataService) Delete(ctx context.Context, resource string, id int64) error {
	model, err := developerModel(resource)
	if err != nil {
		return err
	}
	vectorRefs, err := s.archiveVectorRefs(ctx, resource, []int64{id})
	if err != nil {
		return err
	}
	result := database.DB.WithContext(ctx).Delete(model, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.deleteArchiveVectors(ctx, vectorRefs)
}

func (s *DeveloperDataService) DeleteMany(ctx context.Context, resource string, ids []int64) (int64, error) {
	model, err := developerModel(resource)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("ids are required")
	}
	vectorRefs, err := s.archiveVectorRefs(ctx, resource, ids)
	if err != nil {
		return 0, err
	}
	result := database.DB.WithContext(ctx).Where("id IN ?", ids).Delete(model)
	if result.Error != nil {
		return result.RowsAffected, result.Error
	}
	if err := s.deleteArchiveVectors(ctx, vectorRefs); err != nil {
		return result.RowsAffected, err
	}
	return result.RowsAffected, nil
}

func (s *DeveloperDataService) archiveVectorRefs(ctx context.Context, resource string, ids []int64) ([]archiveVectorRef, error) {
	if resource != DeveloperResourceConversationArchives {
		return nil, nil
	}
	var refs []archiveVectorRef
	err := database.DB.WithContext(ctx).
		Model(&models.ConversationArchive{}).
		Where("id IN ? AND embedding_id <> ''", ids).
		Select("embedding_id, embedding_dimension").
		Find(&refs).Error
	return refs, err
}

func (s *DeveloperDataService) deleteArchiveVectors(ctx context.Context, refs []archiveVectorRef) error {
	if len(refs) == 0 || s.archiveVectors == nil {
		return nil
	}
	idsByDimension := make(map[int][]string)
	for _, ref := range refs {
		idsByDimension[ref.EmbeddingDimension] = append(idsByDimension[ref.EmbeddingDimension], ref.EmbeddingID)
	}
	for dimension, embeddingIDs := range idsByDimension {
		if err := s.archiveVectors.DeleteArchives(ctx, embeddingIDs, dimension); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDeveloperPage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func developerModel(resource string) (any, error) {
	switch resource {
	case DeveloperResourceMessages:
		return &models.Message{}, nil
	case DeveloperResourceScheduledTasks:
		return &models.ScheduledTask{}, nil
	case DeveloperResourceUserProfileSummaries:
		return &models.UserProfileSummary{}, nil
	case DeveloperResourceConversationArchives:
		return &models.ConversationArchive{}, nil
	case DeveloperResourceUserSettings:
		return &models.UserSettings{}, nil
	case DeveloperResourceUsers:
		return &models.User{}, nil
	case DeveloperResourceVoiceFiles:
		return &models.VoiceFile{}, nil
	default:
		return nil, fmt.Errorf("unknown developer resource %q", resource)
	}
}

func developerListTarget(resource string) (any, error) {
	switch resource {
	case DeveloperResourceMessages:
		return &[]models.Message{}, nil
	case DeveloperResourceScheduledTasks:
		return &[]models.ScheduledTask{}, nil
	case DeveloperResourceUserProfileSummaries:
		return &[]models.UserProfileSummary{}, nil
	case DeveloperResourceConversationArchives:
		return &[]models.ConversationArchive{}, nil
	case DeveloperResourceUserSettings:
		return &[]models.UserSettings{}, nil
	case DeveloperResourceUsers:
		return &[]models.User{}, nil
	case DeveloperResourceVoiceFiles:
		return &[]models.VoiceFile{}, nil
	default:
		return nil, fmt.Errorf("unknown developer resource %q", resource)
	}
}

func developerModelID(model any) int64 {
	switch value := model.(type) {
	case *models.Message:
		return value.ID
	case *models.ScheduledTask:
		return value.ID
	case *models.UserProfileSummary:
		return value.ID
	case *models.ConversationArchive:
		return value.ID
	case *models.UserSettings:
		return value.ID
	case *models.User:
		return value.ID
	case *models.VoiceFile:
		return value.ID
	default:
		return 0
	}
}

func allowedDeveloperValues(resource string, values map[string]any) map[string]any {
	allowed := developerWritableFields[resource]
	filtered := make(map[string]any, len(values))
	for field, value := range values {
		field = strings.TrimSpace(field)
		if allowed[field] {
			filtered[field] = value
		}
	}
	return filtered
}

var developerWritableFields = map[string]map[string]bool{
	DeveloperResourceMessages:             {"user_id": true, "role": true, "text_content": true, "translation": true, "voice_file_id": true, "input_type": true},
	DeveloperResourceScheduledTasks:       {"user_id": true, "function_name": true, "args": true, "scheduled_at": true, "status": true},
	DeveloperResourceUserProfileSummaries: {"user_id": true, "summary": true},
	DeveloperResourceConversationArchives: {"user_id": true, "message_ids": true, "summary": true, "message_count": true, "embedding_id": true, "embedding_dimension": true},
	DeveloperResourceUserSettings:         {"user_id": true, "api_base_url": true, "api_key": true, "model": true, "llm_type": true, "embedding_api_base_url": true, "embedding_api_key": true, "embedding_model": true, "embedding_dimension": true, "stt_api_base_url": true, "stt_api_key": true, "stt_model": true, "tts_api_base_url": true, "tts_api_key": true, "tts_model": true, "tts_voice": true, "tts_type": true, "theme": true, "language": true, "native_language": true, "target_language": true, "timezone": true, "default_vocabulary_id": true},
	DeveloperResourceUsers:                {"email": true, "phone": true, "username": true, "password_hash": true, "avatar_url": true, "last_active_at": true, "role": true},
	DeveloperResourceVoiceFiles:           {"user_id": true, "voice_role": true, "voice_url": true, "duration": true, "file_size": true},
}
