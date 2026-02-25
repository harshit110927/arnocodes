package course

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type CourseService struct {
	repo      *CourseRepository
	statusAPI DiagnosticStatusProvider
}

type DiagnosticUserStatus struct {
	DiagnosticCompleted bool
}

type DiagnosticStatusProvider interface {
	GetUserStatus(ctx context.Context, userID string) (DiagnosticUserStatus, error)
}

type CourseTopic struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Prerequisites     []string `json:"prerequisites"`
	UnlockStatus      string   `json:"unlock_status"`
	MasteryScore      float64  `json:"mastery_score"`
	CompletionStatus  string   `json:"completion_status"`
	NumberOfSubtopics int      `json:"number_of_subtopics"`
}

type TopicDetailsResponse struct {
	Topic             CourseTopic            `json:"topic"`
	Subtopics         []SubtopicProgressRow  `json:"subtopics"`
	LearningQuestions []LearningQuestionLink `json:"learning_questions"`
}

type SubtopicDetailsResponse struct {
	Subtopic          SubtopicProgressRow    `json:"subtopic"`
	ParentTopic       CourseTopic            `json:"parent_topic"`
	LearningQuestions []LearningQuestionLink `json:"learning_questions"`
}

func NewCourseService(repo *CourseRepository, statusAPI DiagnosticStatusProvider) *CourseService {
	return &CourseService{repo: repo, statusAPI: statusAPI}
}

func (s *CourseService) ensureDiagnosticSubmitted(ctx context.Context, userID string) error {
	if s.statusAPI == nil {
		return fmt.Errorf("diagnostic status provider is not initialized")
	}
	status, err := s.statusAPI.GetUserStatus(ctx, userID)
	if err != nil {
		return err
	}
	if !status.DiagnosticCompleted {
		return ErrDiagnosticNotCompleted
	}
	return nil
}

func (s *CourseService) GetCourse(ctx context.Context, userID string) ([]CourseTopic, error) {
	if err := s.ensureDiagnosticSubmitted(ctx, userID); err != nil {
		return nil, err
	}
	topicRows, err := s.repo.ListCourseTopicRows(ctx, userID)
	if err != nil {
		return nil, err
	}
	prereqs, err := s.repo.ListTopicPrerequisites(ctx)
	if err != nil {
		return nil, err
	}
	return BuildCourseTopics(topicRows, prereqs), nil
}

func (s *CourseService) GetTopic(ctx context.Context, userID, topicID string) (TopicDetailsResponse, error) {
	if err := s.ensureDiagnosticSubmitted(ctx, userID); err != nil {
		return TopicDetailsResponse{}, err
	}
	base, err := s.repo.GetTopicBase(ctx, topicID)
	if err != nil {
		return TopicDetailsResponse{}, err
	}
	course, err := s.GetCourse(ctx, userID)
	if err != nil {
		return TopicDetailsResponse{}, err
	}
	var current *CourseTopic
	for i := range course {
		if course[i].ID == topicID {
			current = &course[i]
			break
		}
	}
	if current == nil {
		return TopicDetailsResponse{}, pgx.ErrNoRows
	}
	if current.UnlockStatus == "locked" {
		return TopicDetailsResponse{}, ErrTopicLocked
	}
	subtopics, err := s.repo.ListSubtopicsWithProgressByTopic(ctx, userID, topicID)
	if err != nil {
		return TopicDetailsResponse{}, err
	}
	questions, err := s.repo.ListLearningQuestionsByTopic(ctx, topicID)
	if err != nil {
		return TopicDetailsResponse{}, err
	}
	current.Name = base.Name
	return TopicDetailsResponse{Topic: *current, Subtopics: subtopics, LearningQuestions: questions}, nil
}

func (s *CourseService) GetSubtopic(ctx context.Context, userID, subtopicID string) (SubtopicDetailsResponse, error) {
	if err := s.ensureDiagnosticSubmitted(ctx, userID); err != nil {
		return SubtopicDetailsResponse{}, err
	}
	subtopic, err := s.repo.GetSubtopicWithProgress(ctx, userID, subtopicID)
	if err != nil {
		return SubtopicDetailsResponse{}, err
	}
	topic, err := s.GetTopic(ctx, userID, subtopic.TopicID)
	if err != nil {
		return SubtopicDetailsResponse{}, err
	}
	return SubtopicDetailsResponse{Subtopic: subtopic, ParentTopic: topic.Topic, LearningQuestions: topic.LearningQuestions}, nil
}

func BuildCourseTopics(topicRows []TopicRow, prereqs []TopicPrerequisite) []CourseTopic {
	mastery := make(map[string]float64, len(topicRows))
	prereqMap := make(map[string][]string)
	for _, p := range prereqs {
		prereqMap[p.TopicID] = append(prereqMap[p.TopicID], p.PrerequisiteID)
	}
	for _, t := range topicRows {
		if t.MasteryScore != nil {
			mastery[t.TopicID] = *t.MasteryScore
		} else {
			mastery[t.TopicID] = 0
		}
	}
	out := make([]CourseTopic, 0, len(topicRows))
	for _, t := range topicRows {
		m := mastery[t.TopicID]
		unlock := "locked"
		if m >= 80 {
			unlock = "completed"
		} else {
			reqs := prereqMap[t.TopicID]
			if len(reqs) == 0 {
				unlock = "unlocked"
			} else {
				allDone := true
				for _, p := range reqs {
					if mastery[p] < 80 {
						allDone = false
						break
					}
				}
				if allDone {
					unlock = "unlocked"
				}
			}
		}
		completion := t.CompletionStatus
		out = append(out, CourseTopic{
			ID:                t.TopicID,
			Name:              t.Name,
			Prerequisites:     prereqMap[t.TopicID],
			UnlockStatus:      unlock,
			MasteryScore:      m,
			CompletionStatus:  completion,
			NumberOfSubtopics: t.SubtopicCount,
		})
	}
	return out
}
