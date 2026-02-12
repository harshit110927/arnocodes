package assessment

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	TypeDiagnostic = "diagnostic"

	QuestionTypeSlide = "slide"
	QuestionTypeMCQ   = "mcq"

	AttemptInProgress = "in_progress"
	AttemptSubmitted  = "submitted"
	AttemptExpired    = "expired"
)

var (
	ErrTestNotFound         = errors.New("test not found")
	ErrAttemptNotFound      = errors.New("attempt not found")
	ErrAttemptAlreadyDone   = errors.New("diagnostic attempt already submitted")
	ErrAttemptNotActive     = errors.New("attempt not active")
	ErrAttemptExpired       = errors.New("attempt expired")
	ErrNoQuestionsForTopics = errors.New("no questions found for selected topics")
	ErrQuestionOutOfOrder   = errors.New("question submission is out of order")
	ErrInvalidQuestionID    = errors.New("invalid question id")
)

type Test struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	DurationMinutes int    `json:"duration_minutes"`
	TotalMarks      int    `json:"total_marks"`
	Title           string `json:"title"`
}

type Question struct {
	ID            string   `json:"id"`
	TestID        string   `json:"test_id"`
	TopicID       string   `json:"topic_id"`
	QuestionType  string   `json:"question_type"`
	Content       string   `json:"content"`
	Options       []string `json:"options,omitempty"`
	CorrectOption int      `json:"correct_option,omitempty"`
	Marks         int      `json:"marks,omitempty"`
	OrderIndex    int      `json:"order_index"`
}

type QuestionAttempt struct {
	QuestionID      string    `json:"question_id"`
	SelectedOption  int       `json:"selected_option"`
	TimeTakenSecond int       `json:"time_taken_seconds"`
	IsCorrect       bool      `json:"is_correct"`
	AnsweredAt      time.Time `json:"answered_at"`
}

type Attempt struct {
	ID                string
	UserID            string
	TestID            string
	Status            string
	TopicsKnown       []string
	StartedAt         time.Time
	ExpiresAt         time.Time
	SubmittedAt       *time.Time
	CurrentIndex      int
	OrderedQuestionID []string
	QuestionStartedAt time.Time
	Score             int
	Answers           map[string]QuestionAttempt
}

type TestView struct {
	Test      Test       `json:"test"`
	Slides    []Question `json:"slides"`
	Questions []Question `json:"questions"`
}

type SessionView struct {
	AttemptID            string `json:"attempt_id"`
	Status               string `json:"status"`
	CurrentQuestionIndex int    `json:"current_question_index"`
	TotalQuestions       int    `json:"total_questions"`
	RemainingSeconds     int64  `json:"remaining_seconds"`
}

type NextQuestionView struct {
	AttemptID string   `json:"attempt_id"`
	Question  Question `json:"question"`
}

type SubmitAnswerResult struct {
	Accepted      bool   `json:"accepted"`
	NextAvailable bool   `json:"next_available"`
	AttemptStatus string `json:"attempt_status"`
}

type ResultView struct {
	AttemptID    string            `json:"attempt_id"`
	Score        int               `json:"score"`
	TotalMarks   int               `json:"total_marks"`
	Status       string            `json:"status"`
	AnswerReport []QuestionAttempt `json:"answer_report"`
}

type DiagnosticStatus struct {
	DiagnosticCompleted bool `json:"diagnostic_completed"`
	DashboardUnlocked   bool `json:"dashboard_unlocked"`
}

type Service struct {
	mu                sync.RWMutex
	tests             map[string]Test
	questionsByTest   map[string][]Question
	attempts          map[string]*Attempt
	attemptByUserTest map[string]string
	nextAttemptID     int
}

func NewService() *Service {
	s := &Service{
		tests:             make(map[string]Test),
		questionsByTest:   make(map[string][]Question),
		attempts:          make(map[string]*Attempt),
		attemptByUserTest: make(map[string]string),
		nextAttemptID:     1,
	}
	s.seed()
	return s
}

func (s *Service) seed() {
	t := Test{ID: "diagnostic-1", Type: TypeDiagnostic, DurationMinutes: 20, TotalMarks: 10, Title: "Diagnostic Basics"}
	s.tests[t.ID] = t
	s.questionsByTest[t.ID] = []Question{
		{ID: "slide-1", TestID: t.ID, TopicID: "arrays", QuestionType: QuestionTypeSlide, Content: "Arrays: contiguous memory and index-based access.", OrderIndex: 1},
		{ID: "slide-2", TestID: t.ID, TopicID: "strings", QuestionType: QuestionTypeSlide, Content: "Strings: immutable vs mutable handling varies by language.", OrderIndex: 2},
		{ID: "slide-3", TestID: t.ID, TopicID: "trees", QuestionType: QuestionTypeSlide, Content: "Trees: root, nodes, and recursive traversal patterns.", OrderIndex: 3},
		{ID: "q-1", TestID: t.ID, TopicID: "arrays", QuestionType: QuestionTypeMCQ, Content: "Best time complexity to access arr[i]?", Options: []string{"O(n)", "O(1)", "O(log n)", "O(n log n)"}, CorrectOption: 2, Marks: 2, OrderIndex: 10},
		{ID: "q-2", TestID: t.ID, TopicID: "arrays", QuestionType: QuestionTypeMCQ, Content: "Two-pointer works best when?", Options: []string{"Random graph", "Sorted/structured sequence", "Hash-only problems", "Dynamic tree updates"}, CorrectOption: 2, Marks: 2, OrderIndex: 11},
		{ID: "q-3", TestID: t.ID, TopicID: "strings", QuestionType: QuestionTypeMCQ, Content: "Common pattern for substring lookup optimization?", Options: []string{"DFS", "Prefix function / rolling hash", "Disjoint set", "Floyd cycle"}, CorrectOption: 2, Marks: 2, OrderIndex: 12},
		{ID: "q-4", TestID: t.ID, TopicID: "trees", QuestionType: QuestionTypeMCQ, Content: "Traversal that uses queue typically?", Options: []string{"Inorder", "Preorder", "Level-order", "Postorder"}, CorrectOption: 3, Marks: 2, OrderIndex: 13},
		{ID: "q-5", TestID: t.ID, TopicID: "trees", QuestionType: QuestionTypeMCQ, Content: "Balanced BST search average complexity?", Options: []string{"O(1)", "O(log n)", "O(n)", "O(n^2)"}, CorrectOption: 2, Marks: 2, OrderIndex: 14},
	}
}

func normalizeTopics(topics []string) map[string]bool {
	res := map[string]bool{}
	for _, t := range topics {
		k := strings.TrimSpace(strings.ToLower(t))
		if k != "" {
			res[k] = true
		}
	}
	return res
}

func (s *Service) GetTestView(testID string, topics []string) (TestView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	test, ok := s.tests[testID]
	if !ok {
		return TestView{}, ErrTestNotFound
	}

	topicSet := normalizeTopics(topics)
	questions := s.questionsByTest[testID]
	slides := make([]Question, 0)
	mcqs := make([]Question, 0)
	for _, q := range questions {
		if len(topicSet) > 0 && !topicSet[strings.ToLower(q.TopicID)] {
			continue
		}
		if q.QuestionType == QuestionTypeSlide {
			slides = append(slides, q)
			continue
		}
		qCopy := q
		qCopy.CorrectOption = 0
		mcqs = append(mcqs, qCopy)
	}
	sort.Slice(slides, func(i, j int) bool { return slides[i].OrderIndex < slides[j].OrderIndex })
	sort.Slice(mcqs, func(i, j int) bool { return mcqs[i].OrderIndex < mcqs[j].OrderIndex })

	return TestView{Test: test, Slides: slides, Questions: mcqs}, nil
}

func (s *Service) StartAttempt(userID, testID string, topics []string) (*Attempt, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, false, ErrTestNotFound
	}

	key := userID + "::" + testID
	if existingID, exists := s.attemptByUserTest[key]; exists {
		existing := s.attempts[existingID]
		if existing.Status == AttemptSubmitted {
			return nil, false, ErrAttemptAlreadyDone
		}
		if existing.Status == AttemptExpired {
			return nil, false, ErrAttemptExpired
		}
		return existing, false, nil
	}

	topicSet := normalizeTopics(topics)
	if len(topicSet) == 0 {
		for _, q := range s.questionsByTest[testID] {
			topicSet[strings.ToLower(q.TopicID)] = true
		}
	}

	mcqs := make([]Question, 0)
	for _, q := range s.questionsByTest[testID] {
		if q.QuestionType != QuestionTypeMCQ {
			continue
		}
		if topicSet[strings.ToLower(q.TopicID)] {
			mcqs = append(mcqs, q)
		}
	}
	if len(mcqs) == 0 {
		return nil, false, ErrNoQuestionsForTopics
	}
	sort.Slice(mcqs, func(i, j int) bool { return mcqs[i].OrderIndex < mcqs[j].OrderIndex })
	orderedIDs := make([]string, 0, len(mcqs))
	for _, q := range mcqs {
		orderedIDs = append(orderedIDs, q.ID)
	}

	attemptID := fmt.Sprintf("attempt-%d", s.nextAttemptID)
	s.nextAttemptID++
	now := time.Now().UTC()
	a := &Attempt{
		ID:                attemptID,
		UserID:            userID,
		TestID:            testID,
		Status:            AttemptInProgress,
		TopicsKnown:       topics,
		StartedAt:         now,
		ExpiresAt:         now.Add(time.Duration(test.DurationMinutes) * time.Minute),
		CurrentIndex:      0,
		OrderedQuestionID: orderedIDs,
		QuestionStartedAt: now,
		Answers:           map[string]QuestionAttempt{},
	}
	s.attempts[attemptID] = a
	s.attemptByUserTest[key] = attemptID
	return a, true, nil
}

func (s *Service) getQuestion(testID, questionID string) (Question, bool) {
	for _, q := range s.questionsByTest[testID] {
		if q.ID == questionID {
			return q, true
		}
	}
	return Question{}, false
}

func (s *Service) GetSession(userID, attemptID string) (SessionView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return SessionView{}, ErrAttemptNotFound
	}
	if a.Status == AttemptInProgress && time.Now().UTC().After(a.ExpiresAt) {
		a.Status = AttemptExpired
	}
	remaining := int64(time.Until(a.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return SessionView{AttemptID: a.ID, Status: a.Status, CurrentQuestionIndex: a.CurrentIndex, TotalQuestions: len(a.OrderedQuestionID), RemainingSeconds: remaining}, nil
}

func (s *Service) NextQuestion(userID, attemptID string) (NextQuestionView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return NextQuestionView{}, ErrAttemptNotFound
	}
	if a.Status != AttemptInProgress {
		return NextQuestionView{}, ErrAttemptNotActive
	}
	if time.Now().UTC().After(a.ExpiresAt) {
		a.Status = AttemptExpired
		return NextQuestionView{}, ErrAttemptExpired
	}
	if a.CurrentIndex >= len(a.OrderedQuestionID) {
		return NextQuestionView{}, ErrAttemptNotActive
	}
	qid := a.OrderedQuestionID[a.CurrentIndex]
	q, ok := s.getQuestion(a.TestID, qid)
	if !ok {
		return NextQuestionView{}, ErrInvalidQuestionID
	}
	q.CorrectOption = 0
	return NextQuestionView{AttemptID: a.ID, Question: q}, nil
}

func (s *Service) SubmitAnswer(userID, attemptID, questionID string, selectedOption int) (SubmitAnswerResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return SubmitAnswerResult{}, ErrAttemptNotFound
	}
	if a.Status != AttemptInProgress {
		return SubmitAnswerResult{}, ErrAttemptNotActive
	}
	if time.Now().UTC().After(a.ExpiresAt) {
		a.Status = AttemptExpired
		return SubmitAnswerResult{}, ErrAttemptExpired
	}
	if a.CurrentIndex >= len(a.OrderedQuestionID) {
		return SubmitAnswerResult{}, ErrQuestionOutOfOrder
	}
	currentID := a.OrderedQuestionID[a.CurrentIndex]
	if currentID != questionID {
		return SubmitAnswerResult{}, ErrQuestionOutOfOrder
	}
	q, ok := s.getQuestion(a.TestID, questionID)
	if !ok {
		return SubmitAnswerResult{}, ErrInvalidQuestionID
	}
	taken := int(time.Since(a.QuestionStartedAt).Seconds())
	isCorrect := selectedOption == q.CorrectOption
	a.Answers[questionID] = QuestionAttempt{QuestionID: questionID, SelectedOption: selectedOption, TimeTakenSecond: taken, IsCorrect: isCorrect, AnsweredAt: time.Now().UTC()}
	a.CurrentIndex++
	a.QuestionStartedAt = time.Now().UTC()
	return SubmitAnswerResult{Accepted: true, NextAvailable: a.CurrentIndex < len(a.OrderedQuestionID), AttemptStatus: a.Status}, nil
}

func (s *Service) SubmitAttempt(userID, attemptID string) (ResultView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return ResultView{}, ErrAttemptNotFound
	}
	if a.Status != AttemptInProgress && a.Status != AttemptExpired {
		return ResultView{}, ErrAttemptNotActive
	}
	score := 0
	report := make([]QuestionAttempt, 0, len(a.Answers))
	for _, qid := range a.OrderedQuestionID {
		q, _ := s.getQuestion(a.TestID, qid)
		if ans, ok := a.Answers[qid]; ok {
			report = append(report, ans)
			if ans.IsCorrect {
				score += q.Marks
			}
		}
	}
	now := time.Now().UTC()
	a.SubmittedAt = &now
	a.Status = AttemptSubmitted
	a.Score = score
	test := s.tests[a.TestID]
	return ResultView{AttemptID: a.ID, Score: score, TotalMarks: test.TotalMarks, Status: a.Status, AnswerReport: report}, nil
}

func (s *Service) GetResult(userID, attemptID string) (ResultView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return ResultView{}, ErrAttemptNotFound
	}
	report := make([]QuestionAttempt, 0, len(a.Answers))
	for _, v := range a.Answers {
		report = append(report, v)
	}
	t := s.tests[a.TestID]
	return ResultView{AttemptID: a.ID, Score: a.Score, TotalMarks: t.TotalMarks, Status: a.Status, AnswerReport: report}, nil
}

func (s *Service) ExpireAttempt(userID, attemptID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return ErrAttemptNotFound
	}
	if a.Status == AttemptSubmitted {
		return ErrAttemptNotActive
	}
	a.Status = AttemptExpired
	return nil
}

func (s *Service) ResumeAttempt(userID, attemptID string) (*Attempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attempts[attemptID]
	if !ok || a.UserID != userID {
		return nil, ErrAttemptNotFound
	}
	if time.Now().UTC().After(a.ExpiresAt) {
		a.Status = AttemptExpired
		return nil, ErrAttemptExpired
	}
	if a.Status == AttemptSubmitted {
		return nil, ErrAttemptNotActive
	}
	a.Status = AttemptInProgress
	return a, nil
}

func (s *Service) DiagnosticStatusForUser(userID string) DiagnosticStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := userID + "::" + "diagnostic-1"
	attemptID, ok := s.attemptByUserTest[key]
	if !ok {
		return DiagnosticStatus{DiagnosticCompleted: false, DashboardUnlocked: false}
	}

	attempt := s.attempts[attemptID]
	if attempt == nil || attempt.Status != AttemptSubmitted {
		return DiagnosticStatus{DiagnosticCompleted: false, DashboardUnlocked: false}
	}
	return DiagnosticStatus{DiagnosticCompleted: true, DashboardUnlocked: true}
}
