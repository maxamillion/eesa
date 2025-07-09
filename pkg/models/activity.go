package models

import (
	"errors"
	"fmt"
	"time"
)

// Activity represents a user activity from Jira
type Activity struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Reporter    User      `json:"reporter"`
	Assignee    User      `json:"assignee"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Project     Project   `json:"project"`
	TimeSpent   int64     `json:"time_spent"` // In seconds
	Comments    []Comment `json:"comments"`
	Worklog     []Worklog `json:"worklog"`
}

// User represents a Jira user
type User struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	EmailAddress string `json:"email_address"`
	Active      bool   `json:"active"`
}

// Project represents a Jira project
type Project struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lead        User   `json:"lead"`
}

// Comment represents a comment on a Jira issue
type Comment struct {
	ID       string    `json:"id"`
	Author   User      `json:"author"`
	Body     string    `json:"body"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	Visibility string  `json:"visibility"`
}

// Worklog represents a work log entry
type Worklog struct {
	ID          string    `json:"id"`
	Author      User      `json:"author"`
	TimeSpent   int64     `json:"time_spent"` // In seconds
	Description string    `json:"description"`
	Started     time.Time `json:"started"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// ActivityFilter represents filters for querying activities
type ActivityFilter struct {
	Users       []string  `json:"users"`
	Projects    []string  `json:"projects"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      []string  `json:"status"`
	Priority    []string  `json:"priority"`
	Type        []string  `json:"type"`
	MaxResults  int       `json:"max_results"`
	StartAt     int       `json:"start_at"`
}

// DefaultActivityFilter returns a default activity filter
func DefaultActivityFilter() *ActivityFilter {
	return &ActivityFilter{
		Users:      []string{},
		Projects:   []string{},
		Status:     []string{},
		Priority:   []string{},
		Type:       []string{},
		MaxResults: 100,
		StartAt:    0,
	}
}

// ActivitySummary represents a summary of activities
type ActivitySummary struct {
	TotalIssues       int                    `json:"total_issues"`
	CompletedIssues   int                    `json:"completed_issues"`
	InProgressIssues  int                    `json:"in_progress_issues"`
	TotalTimeSpent    int64                  `json:"total_time_spent"` // In seconds
	UserSummaries     []UserActivitySummary  `json:"user_summaries"`
	ProjectSummaries  []ProjectActivitySummary `json:"project_summaries"`
	PrioritySummaries []PriorityActivitySummary `json:"priority_summaries"`
	GeneratedAt       time.Time              `json:"generated_at"`
}

// UserActivitySummary represents activity summary for a user
type UserActivitySummary struct {
	User            User  `json:"user"`
	IssuesCreated   int   `json:"issues_created"`
	IssuesAssigned  int   `json:"issues_assigned"`
	IssuesCompleted int   `json:"issues_completed"`
	TimeSpent       int64 `json:"time_spent"` // In seconds
	CommentsAdded   int   `json:"comments_added"`
}

// ProjectActivitySummary represents activity summary for a project
type ProjectActivitySummary struct {
	Project         Project `json:"project"`
	TotalIssues     int     `json:"total_issues"`
	CompletedIssues int     `json:"completed_issues"`
	TimeSpent       int64   `json:"time_spent"` // In seconds
}

// PriorityActivitySummary represents activity summary by priority
type PriorityActivitySummary struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
	TimeSpent int64 `json:"time_spent"` // In seconds
}

// FormatTimeSpent formats time spent in seconds to a human-readable string
func FormatTimeSpent(seconds int64) string {
	if seconds == 0 {
		return "0m"
	}
	
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// GetFormattedTimeSpent returns formatted time spent string
func (a *Activity) GetFormattedTimeSpent() string {
	return FormatTimeSpent(a.TimeSpent)
}

// GetFormattedTimeSpent returns formatted time spent string
func (w *Worklog) GetFormattedTimeSpent() string {
	return FormatTimeSpent(w.TimeSpent)
}

// GetFormattedTimeSpent returns formatted time spent string
func (s *ActivitySummary) GetFormattedTimeSpent() string {
	return FormatTimeSpent(s.TotalTimeSpent)
}

// GetFormattedTimeSpent returns formatted time spent string
func (u *UserActivitySummary) GetFormattedTimeSpent() string {
	return FormatTimeSpent(u.TimeSpent)
}

// GetFormattedTimeSpent returns formatted time spent string
func (p *ProjectActivitySummary) GetFormattedTimeSpent() string {
	return FormatTimeSpent(p.TimeSpent)
}

// GetFormattedTimeSpent returns formatted time spent string
func (p *PriorityActivitySummary) GetFormattedTimeSpent() string {
	return FormatTimeSpent(p.TimeSpent)
}

// IsCompleted returns true if the activity is completed
func (a *Activity) IsCompleted() bool {
	completedStatuses := []string{
		"Done",
		"Completed",
		"Closed",
		"Resolved",
	}
	
	for _, status := range completedStatuses {
		if a.Status == status {
			return true
		}
	}
	return false
}

// IsInProgress returns true if the activity is in progress
func (a *Activity) IsInProgress() bool {
	inProgressStatuses := []string{
		"In Progress",
		"In Development",
		"In Review",
		"In Testing",
	}
	
	for _, status := range inProgressStatuses {
		if a.Status == status {
			return true
		}
	}
	return false
}

// GetPriorityWeight returns a numeric weight for priority sorting
func (a *Activity) GetPriorityWeight() int {
	switch a.Priority {
	case "Highest":
		return 5
	case "High":
		return 4
	case "Medium":
		return 3
	case "Low":
		return 2
	case "Lowest":
		return 1
	default:
		return 0
	}
}

// AddComment adds a comment to the activity
func (a *Activity) AddComment(comment Comment) {
	a.Comments = append(a.Comments, comment)
}

// AddWorklog adds a worklog entry to the activity
func (a *Activity) AddWorklog(worklog Worklog) {
	a.Worklog = append(a.Worklog, worklog)
	a.TimeSpent += worklog.TimeSpent
}

// GetCommentsByUser returns comments by a specific user
func (a *Activity) GetCommentsByUser(userID string) []Comment {
	var comments []Comment
	for _, comment := range a.Comments {
		if comment.Author.AccountID == userID {
			comments = append(comments, comment)
		}
	}
	return comments
}

// GetWorklogByUser returns worklog entries by a specific user
func (a *Activity) GetWorklogByUser(userID string) []Worklog {
	var worklog []Worklog
	for _, entry := range a.Worklog {
		if entry.Author.AccountID == userID {
			worklog = append(worklog, entry)
		}
	}
	return worklog
}

// GetUserTimeSpent returns total time spent by a specific user
func (a *Activity) GetUserTimeSpent(userID string) int64 {
	var totalTime int64
	for _, entry := range a.Worklog {
		if entry.Author.AccountID == userID {
			totalTime += entry.TimeSpent
		}
	}
	return totalTime
}

// GetCommentCount returns the number of comments
func (a *Activity) GetCommentCount() int {
	return len(a.Comments)
}

// GetWorklogCount returns the number of worklog entries
func (a *Activity) GetWorklogCount() int {
	return len(a.Worklog)
}

// GetLastUpdated returns the most recent update time
func (a *Activity) GetLastUpdated() time.Time {
	lastUpdated := a.Updated
	
	for _, comment := range a.Comments {
		if comment.Updated.After(lastUpdated) {
			lastUpdated = comment.Updated
		}
	}
	
	for _, worklog := range a.Worklog {
		if worklog.Updated.After(lastUpdated) {
			lastUpdated = worklog.Updated
		}
	}
	
	return lastUpdated
}

// ToSummary creates a summary representation of the activity
func (a *Activity) ToSummary() string {
	return fmt.Sprintf("%s: %s [%s] - %s", a.Key, a.Summary, a.Status, a.Priority)
}

// Validate validates the activity data
func (a *Activity) Validate() error {
	if a.ID == "" {
		return errors.New("activity ID is required")
	}
	
	if a.Key == "" {
		return errors.New("activity key is required")
	}
	
	if a.Summary == "" {
		return errors.New("activity summary is required")
	}
	
	if a.Type == "" {
		return errors.New("activity type is required")
	}
	
	if a.Status == "" {
		return errors.New("activity status is required")
	}
	
	if a.Created.IsZero() {
		return errors.New("activity created date is required")
	}
	
	if a.Updated.IsZero() {
		return errors.New("activity updated date is required")
	}
	
	return nil
}

// GetAge returns the age of the activity in days
func (a *Activity) GetAge() int {
	return int(time.Since(a.Created).Hours() / 24)
}

// GetTimeSinceLastUpdate returns the time since last update in days
func (a *Activity) GetTimeSinceLastUpdate() int {
	return int(time.Since(a.GetLastUpdated()).Hours() / 24)
}

// IsStale returns true if the activity hasn't been updated in the specified days
func (a *Activity) IsStale(days int) bool {
	return a.GetTimeSinceLastUpdate() > days
}

// IsOverdue returns true if the activity is overdue based on priority
func (a *Activity) IsOverdue() bool {
	daysSinceCreated := a.GetAge()
	
	switch a.Priority {
	case "Highest":
		return daysSinceCreated > 1
	case "High":
		return daysSinceCreated > 3
	case "Medium":
		return daysSinceCreated > 7
	case "Low":
		return daysSinceCreated > 14
	case "Lowest":
		return daysSinceCreated > 30
	default:
		return daysSinceCreated > 14
	}
}