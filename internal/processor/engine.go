package processor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
)

// DataProcessor handles aggregation and analysis of user activities
type DataProcessor struct {
	logger utils.Logger
}

// NewDataProcessor creates a new data processor instance
func NewDataProcessor(logger utils.Logger) *DataProcessor {
	return &DataProcessor{
		logger: logger,
	}
}

// ProcessingOptions configures how activities are processed
type ProcessingOptions struct {
	IncludeComments     bool
	IncludeWorklogs     bool
	GroupByPriority     bool
	GroupByStatus       bool
	GroupByUser         bool
	CalculateVelocity   bool
	AnalyzeTrends       bool
	CustomTimeRanges    []TimeRange
	MinimumTimeSpent    int64 // Minimum seconds to include activity
}

// TimeRange represents a time period for analysis
type TimeRange struct {
	Start time.Time
	End   time.Time
	Label string
}

// ProcessingResult contains the results of activity processing
type ProcessingResult struct {
	Summary           ProcessingSummary           `json:"summary"`
	UserMetrics       map[string]UserMetrics      `json:"user_metrics"`
	PriorityBreakdown map[string]PriorityMetrics  `json:"priority_breakdown"`
	StatusBreakdown   map[string]StatusMetrics    `json:"status_breakdown"`
	TrendAnalysis     *TrendAnalysis              `json:"trend_analysis,omitempty"`
	VelocityMetrics   *VelocityMetrics            `json:"velocity_metrics,omitempty"`
	ProcessedAt       time.Time                   `json:"processed_at"`
	ProcessingTime    time.Duration               `json:"processing_time"`
}

// ProcessingSummary provides high-level metrics
type ProcessingSummary struct {
	TotalActivities    int           `json:"total_activities"`
	TotalUsers         int           `json:"total_users"`
	TotalTimeSpent     int64         `json:"total_time_spent"` // In seconds
	AverageTimePerUser int64         `json:"average_time_per_user"`
	DateRange          TimeRange     `json:"date_range"`
	MostActiveUser     string        `json:"most_active_user"`
	TopPriority        string        `json:"top_priority"`
	CompletionRate     float64       `json:"completion_rate"`
	ProductivityScore  float64       `json:"productivity_score"`
}

// UserMetrics contains metrics for a specific user
type UserMetrics struct {
	UserID             string            `json:"user_id"`
	DisplayName        string            `json:"display_name"`
	TotalActivities    int               `json:"total_activities"`
	CompletedActivities int              `json:"completed_activities"`
	TotalTimeSpent     int64             `json:"total_time_spent"`
	AverageTimePerTask int64             `json:"average_time_per_task"`
	PriorityDistribution map[string]int  `json:"priority_distribution"`
	StatusDistribution   map[string]int  `json:"status_distribution"`
	CompletionRate     float64           `json:"completion_rate"`
	ProductivityRank   int               `json:"productivity_rank"`
	TopIssues          []string          `json:"top_issues"`
}

// PriorityMetrics contains metrics for a specific priority level
type PriorityMetrics struct {
	Priority        string  `json:"priority"`
	Count           int     `json:"count"`
	TotalTimeSpent  int64   `json:"total_time_spent"`
	CompletedCount  int     `json:"completed_count"`
	CompletionRate  float64 `json:"completion_rate"`
	AverageTimeToComplete int64 `json:"average_time_to_complete"`
}

// StatusMetrics contains metrics for a specific status
type StatusMetrics struct {
	Status         string    `json:"status"`
	Count          int       `json:"count"`
	TotalTimeSpent int64     `json:"total_time_spent"`
	Users          []string  `json:"users"`
	RecentChanges  int       `json:"recent_changes"`
}

// TrendAnalysis contains trend analysis over time
type TrendAnalysis struct {
	TimeRanges        []TimeRangeMetrics `json:"time_ranges"`
	OverallTrend      string             `json:"overall_trend"` // "increasing", "decreasing", "stable"
	VelocityTrend     string             `json:"velocity_trend"`
	ProductivityTrend string             `json:"productivity_trend"`
	Seasonality       map[string]float64 `json:"seasonality"` // Day of week patterns
}

// TimeRangeMetrics contains metrics for a specific time range
type TimeRangeMetrics struct {
	Range              TimeRange `json:"range"`
	ActivityCount      int       `json:"activity_count"`
	CompletionCount    int       `json:"completion_count"`
	TotalTimeSpent     int64     `json:"total_time_spent"`
	AverageVelocity    float64   `json:"average_velocity"`
	ProductivityScore  float64   `json:"productivity_score"`
}

// VelocityMetrics contains velocity-related metrics
type VelocityMetrics struct {
	CurrentVelocity     float64           `json:"current_velocity"`
	AverageVelocity     float64           `json:"average_velocity"`
	VelocityTrend       string            `json:"velocity_trend"`
	BurndownRate        float64           `json:"burndown_rate"`
	SprintMetrics       []SprintMetrics   `json:"sprint_metrics,omitempty"`
	UserVelocities      map[string]float64 `json:"user_velocities"`
}

// SprintMetrics contains sprint-specific metrics
type SprintMetrics struct {
	SprintName        string    `json:"sprint_name"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	PlannedStoryPoints int      `json:"planned_story_points"`
	CompletedStoryPoints int    `json:"completed_story_points"`
	Velocity          float64   `json:"velocity"`
	BurndownRate      float64   `json:"burndown_rate"`
}

// ProcessActivities processes a collection of activities and returns aggregated metrics
func (dp *DataProcessor) ProcessActivities(ctx context.Context, activities []models.Activity, options ProcessingOptions) (*ProcessingResult, error) {
	startTime := time.Now()
	
	dp.logger.Info("Starting activity processing",
		utils.NewField("activity_count", len(activities)),
		utils.NewField("include_comments", options.IncludeComments),
		utils.NewField("include_worklogs", options.IncludeWorklogs),
	)
	
	if len(activities) == 0 {
		return &ProcessingResult{
			ProcessedAt:    time.Now(),
			ProcessingTime: time.Since(startTime),
		}, nil
	}
	
	// Filter activities based on minimum time spent
	filteredActivities := dp.filterActivities(activities, options)
	
	// Build result structure
	result := &ProcessingResult{
		UserMetrics:       make(map[string]UserMetrics),
		PriorityBreakdown: make(map[string]PriorityMetrics),
		StatusBreakdown:   make(map[string]StatusMetrics),
		ProcessedAt:       time.Now(),
	}
	
	// Process basic metrics
	dp.processSummaryMetrics(filteredActivities, &result.Summary)
	
	// Process user metrics
	if options.GroupByUser {
		dp.processUserMetrics(filteredActivities, result.UserMetrics)
	}
	
	// Process priority breakdown
	if options.GroupByPriority {
		dp.processPriorityMetrics(filteredActivities, result.PriorityBreakdown)
	}
	
	// Process status breakdown
	if options.GroupByStatus {
		dp.processStatusMetrics(filteredActivities, result.StatusBreakdown)
	}
	
	// Process trend analysis
	if options.AnalyzeTrends {
		trendAnalysis := dp.analyzeTrends(filteredActivities, options.CustomTimeRanges)
		result.TrendAnalysis = trendAnalysis
	}
	
	// Process velocity metrics
	if options.CalculateVelocity {
		velocityMetrics := dp.calculateVelocityMetrics(filteredActivities)
		result.VelocityMetrics = velocityMetrics
	}
	
	result.ProcessingTime = time.Since(startTime)
	
	dp.logger.Info("Activity processing completed",
		utils.NewField("total_activities", result.Summary.TotalActivities),
		utils.NewField("total_users", result.Summary.TotalUsers),
		utils.NewField("processing_time", result.ProcessingTime),
	)
	
	return result, nil
}

// filterActivities filters activities based on processing options
func (dp *DataProcessor) filterActivities(activities []models.Activity, options ProcessingOptions) []models.Activity {
	if options.MinimumTimeSpent <= 0 {
		return activities
	}
	
	filtered := make([]models.Activity, 0, len(activities))
	for _, activity := range activities {
		if activity.TimeSpent >= options.MinimumTimeSpent {
			filtered = append(filtered, activity)
		}
	}
	
	dp.logger.Debug("Filtered activities",
		utils.NewField("original_count", len(activities)),
		utils.NewField("filtered_count", len(filtered)),
		utils.NewField("minimum_time_spent", options.MinimumTimeSpent),
	)
	
	return filtered
}

// processSummaryMetrics calculates high-level summary metrics
func (dp *DataProcessor) processSummaryMetrics(activities []models.Activity, summary *ProcessingSummary) {
	if len(activities) == 0 {
		return
	}
	
	userSet := make(map[string]bool)
	priorityCount := make(map[string]int)
	userTimeSpent := make(map[string]int64)
	completedCount := 0
	totalTimeSpent := int64(0)
	
	// Find date range
	minDate := activities[0].Created
	maxDate := activities[0].Updated
	
	for _, activity := range activities {
		// Track unique users
		userSet[activity.Assignee.AccountID] = true
		
		// Track priority distribution
		priorityCount[activity.Priority]++
		
		// Track user time spent
		userTimeSpent[activity.Assignee.AccountID] += activity.TimeSpent
		
		// Track completion
		if dp.isCompleted(activity.Status) {
			completedCount++
		}
		
		// Track total time spent
		totalTimeSpent += activity.TimeSpent
		
		// Track date range
		if activity.Created.Before(minDate) {
			minDate = activity.Created
		}
		if activity.Updated.After(maxDate) {
			maxDate = activity.Updated
		}
	}
	
	// Find most active user
	mostActiveUser := ""
	maxTime := int64(0)
	for userID, timeSpent := range userTimeSpent {
		if timeSpent > maxTime {
			maxTime = timeSpent
			mostActiveUser = userID
		}
	}
	
	// Find top priority
	topPriority := ""
	maxCount := 0
	for priority, count := range priorityCount {
		if count > maxCount {
			maxCount = count
			topPriority = priority
		}
	}
	
	// Calculate metrics
	totalUsers := len(userSet)
	completionRate := float64(completedCount) / float64(len(activities)) * 100
	averageTimePerUser := int64(0)
	if totalUsers > 0 {
		averageTimePerUser = totalTimeSpent / int64(totalUsers)
	}
	
	// Calculate productivity score (0-100 based on completion rate and time efficiency)
	productivityScore := dp.calculateProductivityScore(activities, completionRate)
	
	*summary = ProcessingSummary{
		TotalActivities:    len(activities),
		TotalUsers:         totalUsers,
		TotalTimeSpent:     totalTimeSpent,
		AverageTimePerUser: averageTimePerUser,
		DateRange: TimeRange{
			Start: minDate,
			End:   maxDate,
			Label: fmt.Sprintf("%s to %s", minDate.Format("2006-01-02"), maxDate.Format("2006-01-02")),
		},
		MostActiveUser:    mostActiveUser,
		TopPriority:       topPriority,
		CompletionRate:    completionRate,
		ProductivityScore: productivityScore,
	}
}

// processUserMetrics calculates per-user metrics
func (dp *DataProcessor) processUserMetrics(activities []models.Activity, userMetrics map[string]UserMetrics) {
	userActivities := make(map[string][]models.Activity)
	
	// Group activities by user
	for _, activity := range activities {
		userID := activity.Assignee.AccountID
		userActivities[userID] = append(userActivities[userID], activity)
	}
	
	// Calculate metrics for each user
	userProductivity := make(map[string]float64)
	for userID, activities := range userActivities {
		metrics := dp.calculateUserMetrics(userID, activities)
		userMetrics[userID] = metrics
		userProductivity[userID] = metrics.CompletionRate
	}
	
	// Rank users by productivity
	dp.rankUsersByProductivity(userMetrics, userProductivity)
}

// calculateUserMetrics calculates metrics for a single user
func (dp *DataProcessor) calculateUserMetrics(userID string, activities []models.Activity) UserMetrics {
	if len(activities) == 0 {
		return UserMetrics{UserID: userID}
	}
	
	priorityDist := make(map[string]int)
	statusDist := make(map[string]int)
	totalTimeSpent := int64(0)
	completedCount := 0
	topIssues := make([]string, 0)
	
	displayName := activities[0].Assignee.DisplayName
	
	for _, activity := range activities {
		priorityDist[activity.Priority]++
		statusDist[activity.Status]++
		totalTimeSpent += activity.TimeSpent
		
		if dp.isCompleted(activity.Status) {
			completedCount++
		}
		
		// Track top issues (those with high time investment)
		if activity.TimeSpent >= 7200 { // 2 hours or more
			topIssues = append(topIssues, activity.Key)
		}
	}
	
	averageTimePerTask := int64(0)
	if len(activities) > 0 {
		averageTimePerTask = totalTimeSpent / int64(len(activities))
	}
	
	completionRate := float64(completedCount) / float64(len(activities)) * 100
	
	return UserMetrics{
		UserID:               userID,
		DisplayName:          displayName,
		TotalActivities:      len(activities),
		CompletedActivities:  completedCount,
		TotalTimeSpent:       totalTimeSpent,
		AverageTimePerTask:   averageTimePerTask,
		PriorityDistribution: priorityDist,
		StatusDistribution:   statusDist,
		CompletionRate:       completionRate,
		TopIssues:            topIssues,
	}
}

// processPriorityMetrics calculates priority-based metrics
func (dp *DataProcessor) processPriorityMetrics(activities []models.Activity, priorityMetrics map[string]PriorityMetrics) {
	priorityActivities := make(map[string][]models.Activity)
	
	// Group activities by priority
	for _, activity := range activities {
		priority := activity.Priority
		if priority == "" {
			priority = "None"
		}
		priorityActivities[priority] = append(priorityActivities[priority], activity)
	}
	
	// Calculate metrics for each priority
	for priority, activities := range priorityActivities {
		metrics := dp.calculatePriorityMetrics(priority, activities)
		priorityMetrics[priority] = metrics
	}
}

// calculatePriorityMetrics calculates metrics for a specific priority level
func (dp *DataProcessor) calculatePriorityMetrics(priority string, activities []models.Activity) PriorityMetrics {
	if len(activities) == 0 {
		return PriorityMetrics{Priority: priority}
	}
	
	totalTimeSpent := int64(0)
	completedCount := 0
	completionTimes := make([]int64, 0)
	
	for _, activity := range activities {
		totalTimeSpent += activity.TimeSpent
		
		if dp.isCompleted(activity.Status) {
			completedCount++
			// Calculate time to complete (simplified: use time spent as proxy)
			completionTimes = append(completionTimes, activity.TimeSpent)
		}
	}
	
	completionRate := float64(completedCount) / float64(len(activities)) * 100
	
	averageTimeToComplete := int64(0)
	if len(completionTimes) > 0 {
		total := int64(0)
		for _, time := range completionTimes {
			total += time
		}
		averageTimeToComplete = total / int64(len(completionTimes))
	}
	
	return PriorityMetrics{
		Priority:              priority,
		Count:                 len(activities),
		TotalTimeSpent:        totalTimeSpent,
		CompletedCount:        completedCount,
		CompletionRate:        completionRate,
		AverageTimeToComplete: averageTimeToComplete,
	}
}

// processStatusMetrics calculates status-based metrics
func (dp *DataProcessor) processStatusMetrics(activities []models.Activity, statusMetrics map[string]StatusMetrics) {
	statusActivities := make(map[string][]models.Activity)
	
	// Group activities by status
	for _, activity := range activities {
		status := activity.Status
		if status == "" {
			status = "Unknown"
		}
		statusActivities[status] = append(statusActivities[status], activity)
	}
	
	// Calculate metrics for each status
	for status, activities := range statusActivities {
		metrics := dp.calculateStatusMetrics(status, activities)
		statusMetrics[status] = metrics
	}
}

// calculateStatusMetrics calculates metrics for a specific status
func (dp *DataProcessor) calculateStatusMetrics(status string, activities []models.Activity) StatusMetrics {
	if len(activities) == 0 {
		return StatusMetrics{Status: status}
	}
	
	userSet := make(map[string]bool)
	totalTimeSpent := int64(0)
	recentChanges := 0
	recentCutoff := time.Now().Add(-7 * 24 * time.Hour) // Last 7 days
	
	for _, activity := range activities {
		userSet[activity.Assignee.AccountID] = true
		totalTimeSpent += activity.TimeSpent
		
		if activity.Updated.After(recentCutoff) {
			recentChanges++
		}
	}
	
	users := make([]string, 0, len(userSet))
	for userID := range userSet {
		users = append(users, userID)
	}
	
	return StatusMetrics{
		Status:         status,
		Count:          len(activities),
		TotalTimeSpent: totalTimeSpent,
		Users:          users,
		RecentChanges:  recentChanges,
	}
}

// Helper methods

// isCompleted checks if a status indicates completion
func (dp *DataProcessor) isCompleted(status string) bool {
	completedStatuses := map[string]bool{
		"Done":     true,
		"Closed":   true,
		"Resolved": true,
		"Complete": true,
		"Finished": true,
	}
	return completedStatuses[status]
}

// calculateProductivityScore calculates a productivity score based on various factors
func (dp *DataProcessor) calculateProductivityScore(activities []models.Activity, completionRate float64) float64 {
	if len(activities) == 0 {
		return 0
	}
	
	// Base score from completion rate (0-50 points)
	score := completionRate * 0.5
	
	// Bonus for high-priority completion (0-25 points)
	highPriorityCompleted := 0
	highPriorityTotal := 0
	
	// Bonus for time efficiency (0-25 points)
	totalEfficiency := 0.0
	
	for _, activity := range activities {
		if activity.Priority == "High" || activity.Priority == "Critical" {
			highPriorityTotal++
			if dp.isCompleted(activity.Status) {
				highPriorityCompleted++
			}
		}
		
		// Simple efficiency metric: completed items get higher score
		if dp.isCompleted(activity.Status) {
			totalEfficiency += 1.0
		} else {
			totalEfficiency += 0.5 // Partial credit for in-progress work
		}
	}
	
	// Add high-priority bonus
	if highPriorityTotal > 0 {
		highPriorityRate := float64(highPriorityCompleted) / float64(highPriorityTotal)
		score += highPriorityRate * 25
	}
	
	// Add efficiency bonus
	efficiencyScore := (totalEfficiency / float64(len(activities))) * 25
	score += efficiencyScore
	
	// Cap at 100
	if score > 100 {
		score = 100
	}
	
	return score
}

// rankUsersByProductivity assigns productivity ranks to users
func (dp *DataProcessor) rankUsersByProductivity(userMetrics map[string]UserMetrics, userProductivity map[string]float64) {
	// Sort users by productivity score
	type userScore struct {
		userID string
		score  float64
	}
	
	scores := make([]userScore, 0, len(userProductivity))
	for userID, score := range userProductivity {
		scores = append(scores, userScore{userID: userID, score: score})
	}
	
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Assign ranks
	for rank, userScore := range scores {
		if metrics, exists := userMetrics[userScore.userID]; exists {
			metrics.ProductivityRank = rank + 1
			userMetrics[userScore.userID] = metrics
		}
	}
}

// analyzeTrends performs trend analysis over time
func (dp *DataProcessor) analyzeTrends(activities []models.Activity, timeRanges []TimeRange) *TrendAnalysis {
	// Implementation for trend analysis
	// This is a simplified version - can be expanded based on needs
	
	analysis := &TrendAnalysis{
		TimeRanges:        make([]TimeRangeMetrics, 0),
		OverallTrend:      "stable",
		VelocityTrend:     "stable",
		ProductivityTrend: "stable",
		Seasonality:       make(map[string]float64),
	}
	
	// If no custom time ranges provided, create weekly ranges
	if len(timeRanges) == 0 {
		timeRanges = dp.generateWeeklyRanges(activities)
	}
	
	// Analyze each time range
	for _, timeRange := range timeRanges {
		rangeActivities := dp.filterActivitiesByTimeRange(activities, timeRange)
		rangeMetrics := dp.calculateTimeRangeMetrics(timeRange, rangeActivities)
		analysis.TimeRanges = append(analysis.TimeRanges, rangeMetrics)
	}
	
	// Calculate overall trends
	dp.calculateTrends(analysis)
	
	// Calculate seasonality patterns
	dp.calculateSeasonality(activities, analysis)
	
	return analysis
}

// calculateVelocityMetrics calculates velocity-related metrics
func (dp *DataProcessor) calculateVelocityMetrics(activities []models.Activity) *VelocityMetrics {
	// Simplified velocity calculation
	// In a real implementation, this would integrate with sprint data
	
	userVelocities := make(map[string]float64)
	userActivities := make(map[string][]models.Activity)
	
	// Group by user
	for _, activity := range activities {
		userID := activity.Assignee.AccountID
		userActivities[userID] = append(userActivities[userID], activity)
	}
	
	// Calculate velocity per user (activities completed per day)
	totalVelocity := 0.0
	for userID, userActs := range userActivities {
		if len(userActs) == 0 {
			continue
		}
		
		// Calculate date range for user
		minDate := userActs[0].Created
		maxDate := userActs[0].Updated
		completedCount := 0
		
		for _, activity := range userActs {
			if activity.Created.Before(minDate) {
				minDate = activity.Created
			}
			if activity.Updated.After(maxDate) {
				maxDate = activity.Updated
			}
			if dp.isCompleted(activity.Status) {
				completedCount++
			}
		}
		
		days := maxDate.Sub(minDate).Hours() / 24
		if days > 0 {
			velocity := float64(completedCount) / days
			userVelocities[userID] = velocity
			totalVelocity += velocity
		}
	}
	
	averageVelocity := 0.0
	if len(userVelocities) > 0 {
		averageVelocity = totalVelocity / float64(len(userVelocities))
	}
	
	return &VelocityMetrics{
		CurrentVelocity: averageVelocity,
		AverageVelocity: averageVelocity,
		VelocityTrend:   "stable",
		BurndownRate:    0.8, // Placeholder
		UserVelocities:  userVelocities,
	}
}

// Helper methods for trend analysis

func (dp *DataProcessor) generateWeeklyRanges(activities []models.Activity) []TimeRange {
	if len(activities) == 0 {
		return []TimeRange{}
	}
	
	// Find overall date range
	minDate := activities[0].Created
	maxDate := activities[0].Updated
	
	for _, activity := range activities {
		if activity.Created.Before(minDate) {
			minDate = activity.Created
		}
		if activity.Updated.After(maxDate) {
			maxDate = activity.Updated
		}
	}
	
	// Generate weekly ranges
	ranges := make([]TimeRange, 0)
	current := minDate
	weekNum := 1
	
	for current.Before(maxDate) {
		weekEnd := current.Add(7 * 24 * time.Hour)
		if weekEnd.After(maxDate) {
			weekEnd = maxDate
		}
		
		ranges = append(ranges, TimeRange{
			Start: current,
			End:   weekEnd,
			Label: fmt.Sprintf("Week %d", weekNum),
		})
		
		current = weekEnd
		weekNum++
	}
	
	return ranges
}

func (dp *DataProcessor) filterActivitiesByTimeRange(activities []models.Activity, timeRange TimeRange) []models.Activity {
	filtered := make([]models.Activity, 0)
	
	for _, activity := range activities {
		if (activity.Created.After(timeRange.Start) || activity.Created.Equal(timeRange.Start)) &&
			(activity.Updated.Before(timeRange.End) || activity.Updated.Equal(timeRange.End)) {
			filtered = append(filtered, activity)
		}
	}
	
	return filtered
}

func (dp *DataProcessor) calculateTimeRangeMetrics(timeRange TimeRange, activities []models.Activity) TimeRangeMetrics {
	completionCount := 0
	totalTimeSpent := int64(0)
	
	for _, activity := range activities {
		if dp.isCompleted(activity.Status) {
			completionCount++
		}
		totalTimeSpent += activity.TimeSpent
	}
	
	// Calculate velocity (completed items per day)
	days := timeRange.End.Sub(timeRange.Start).Hours() / 24
	velocity := 0.0
	if days > 0 {
		velocity = float64(completionCount) / days
	}
	
	// Calculate productivity score for this range
	productivityScore := 0.0
	if len(activities) > 0 {
		completionRate := float64(completionCount) / float64(len(activities)) * 100
		productivityScore = dp.calculateProductivityScore(activities, completionRate)
	}
	
	return TimeRangeMetrics{
		Range:             timeRange,
		ActivityCount:     len(activities),
		CompletionCount:   completionCount,
		TotalTimeSpent:    totalTimeSpent,
		AverageVelocity:   velocity,
		ProductivityScore: productivityScore,
	}
}

func (dp *DataProcessor) calculateTrends(analysis *TrendAnalysis) {
	if len(analysis.TimeRanges) < 2 {
		return
	}
	
	// Analyze velocity trend
	velocities := make([]float64, len(analysis.TimeRanges))
	productivities := make([]float64, len(analysis.TimeRanges))
	
	for i, timeRange := range analysis.TimeRanges {
		velocities[i] = timeRange.AverageVelocity
		productivities[i] = timeRange.ProductivityScore
	}
	
	// Simple trend analysis: compare first half with second half
	midpoint := len(velocities) / 2
	
	firstHalfVel := dp.average(velocities[:midpoint])
	secondHalfVel := dp.average(velocities[midpoint:])
	
	if secondHalfVel > firstHalfVel*1.1 {
		analysis.VelocityTrend = "increasing"
	} else if secondHalfVel < firstHalfVel*0.9 {
		analysis.VelocityTrend = "decreasing"
	}
	
	firstHalfProd := dp.average(productivities[:midpoint])
	secondHalfProd := dp.average(productivities[midpoint:])
	
	if secondHalfProd > firstHalfProd*1.1 {
		analysis.ProductivityTrend = "increasing"
		analysis.OverallTrend = "increasing"
	} else if secondHalfProd < firstHalfProd*0.9 {
		analysis.ProductivityTrend = "decreasing"
		analysis.OverallTrend = "decreasing"
	}
}

func (dp *DataProcessor) calculateSeasonality(activities []models.Activity, analysis *TrendAnalysis) {
	dayOfWeekCounts := make(map[string]int)
	dayOfWeekCompleted := make(map[string]int)
	
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for _, day := range days {
		dayOfWeekCounts[day] = 0
		dayOfWeekCompleted[day] = 0
	}
	
	for _, activity := range activities {
		dayName := activity.Created.Weekday().String()
		dayOfWeekCounts[dayName]++
		
		if dp.isCompleted(activity.Status) {
			dayOfWeekCompleted[dayName]++
		}
	}
	
	// Calculate productivity ratio for each day
	for _, day := range days {
		if dayOfWeekCounts[day] > 0 {
			ratio := float64(dayOfWeekCompleted[day]) / float64(dayOfWeekCounts[day])
			analysis.Seasonality[day] = ratio
		}
	}
}

func (dp *DataProcessor) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	
	return sum / float64(len(values))
}