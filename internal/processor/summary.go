package processor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
)

// SummaryGenerator handles the generation of executive summaries from processed data
type SummaryGenerator struct {
	logger utils.Logger
}

// NewSummaryGenerator creates a new summary generator instance
func NewSummaryGenerator(logger utils.Logger) *SummaryGenerator {
	return &SummaryGenerator{
		logger: logger,
	}
}

// SummaryRequest contains the configuration for summary generation
type SummaryRequest struct {
	Title         string           `json:"title"`
	Period        string           `json:"period"`        // "weekly", "monthly", "quarterly"
	IncludeMetrics bool            `json:"include_metrics"`
	IncludeTrends  bool            `json:"include_trends"`
	IncludeUsers   bool            `json:"include_users"`
	CustomSections []string        `json:"custom_sections"`
	MaxUsers       int             `json:"max_users"`      // Limit number of users to include
	MinTimeSpent   int64           `json:"min_time_spent"` // Minimum time in seconds to include activities
	Format         SummaryFormat   `json:"format"`
}

// SummaryFormat defines the output format for the summary
type SummaryFormat string

const (
	FormatExecutive SummaryFormat = "executive"
	FormatDetailed  SummaryFormat = "detailed"
	FormatBulletPoint SummaryFormat = "bullet_point"
	FormatNarrative SummaryFormat = "narrative"
)

// SummaryResponse contains the generated summary
type SummaryResponse struct {
	Title           string                 `json:"title"`
	Period          string                 `json:"period"`
	GeneratedAt     time.Time              `json:"generated_at"`
	ExecutiveSummary string                `json:"executive_summary"`
	KeyMetrics      SummaryKeyMetrics      `json:"key_metrics"`
	Highlights      []string               `json:"highlights"`
	Concerns        []string               `json:"concerns"`
	Recommendations []string               `json:"recommendations"`
	UserInsights    []UserInsight          `json:"user_insights"`
	TrendAnalysis   *SummaryTrendAnalysis  `json:"trend_analysis,omitempty"`
	Sections        map[string]string      `json:"sections"`
	RawData         *ProcessingResult      `json:"raw_data,omitempty"`
}

// SummaryKeyMetrics contains key performance metrics for the summary
type SummaryKeyMetrics struct {
	TotalActivities    int     `json:"total_activities"`
	CompletedActivities int    `json:"completed_activities"`
	CompletionRate     float64 `json:"completion_rate"`
	TotalTimeSpent     string  `json:"total_time_spent"`
	AverageTimePerTask string  `json:"average_time_per_task"`
	ProductivityScore  float64 `json:"productivity_score"`
	ActiveUsers        int     `json:"active_users"`
	TopPriority        string  `json:"top_priority"`
	MostActiveUser     string  `json:"most_active_user"`
}

// UserInsight contains insights about individual users
type UserInsight struct {
	UserID            string  `json:"user_id"`
	DisplayName       string  `json:"display_name"`
	ProductivityRank  int     `json:"productivity_rank"`
	CompletionRate    float64 `json:"completion_rate"`
	TotalActivities   int     `json:"total_activities"`
	TimeSpent         string  `json:"time_spent"`
	KeyAchievements   []string `json:"key_achievements"`
	AreasForImprovement []string `json:"areas_for_improvement"`
}

// SummaryTrendAnalysis contains trend analysis for the summary
type SummaryTrendAnalysis struct {
	OverallTrend      string             `json:"overall_trend"`
	VelocityTrend     string             `json:"velocity_trend"`
	ProductivityTrend string             `json:"productivity_trend"`
	KeyChanges        []string           `json:"key_changes"`
	Seasonality       map[string]float64 `json:"seasonality"`
}

// GenerateSummary generates a comprehensive summary from processing results
func (sg *SummaryGenerator) GenerateSummary(ctx context.Context, data *ProcessingResult, request SummaryRequest) (*SummaryResponse, error) {
	sg.logger.Info("Starting summary generation",
		utils.NewField("title", request.Title),
		utils.NewField("period", request.Period),
		utils.NewField("format", string(request.Format)),
	)

	if data == nil {
		return nil, fmt.Errorf("processing data is required")
	}

	// Build the summary response
	response := &SummaryResponse{
		Title:       request.Title,
		Period:      request.Period,
		GeneratedAt: time.Now(),
		Sections:    make(map[string]string),
	}

	// Include raw data if requested
	if request.Format == FormatDetailed {
		response.RawData = data
	}

	// Generate key metrics
	response.KeyMetrics = sg.generateKeyMetrics(data)

	// Generate executive summary
	response.ExecutiveSummary = sg.generateExecutiveSummary(data, request)

	// Generate highlights and concerns
	response.Highlights = sg.generateHighlights(data)
	response.Concerns = sg.generateConcerns(data)

	// Generate recommendations
	response.Recommendations = sg.generateRecommendations(data)

	// Generate user insights
	if request.IncludeUsers {
		response.UserInsights = sg.generateUserInsights(data, request.MaxUsers)
	}

	// Generate trend analysis
	if request.IncludeTrends && data.TrendAnalysis != nil {
		response.TrendAnalysis = sg.generateTrendAnalysis(data.TrendAnalysis)
	}

	// Generate custom sections
	for _, section := range request.CustomSections {
		response.Sections[section] = sg.generateCustomSection(data, section)
	}

	sg.logger.Info("Summary generation completed",
		utils.NewField("highlights_count", len(response.Highlights)),
		utils.NewField("concerns_count", len(response.Concerns)),
		utils.NewField("recommendations_count", len(response.Recommendations)),
		utils.NewField("user_insights_count", len(response.UserInsights)),
	)

	return response, nil
}

// generateKeyMetrics creates key performance metrics
func (sg *SummaryGenerator) generateKeyMetrics(data *ProcessingResult) SummaryKeyMetrics {
	avgTimePerTask := int64(0)
	if data.Summary.TotalActivities > 0 {
		avgTimePerTask = data.Summary.TotalTimeSpent / int64(data.Summary.TotalActivities)
	}

	return SummaryKeyMetrics{
		TotalActivities:     data.Summary.TotalActivities,
		CompletedActivities: int(float64(data.Summary.TotalActivities) * data.Summary.CompletionRate / 100),
		CompletionRate:      data.Summary.CompletionRate,
		TotalTimeSpent:      models.FormatTimeSpent(data.Summary.TotalTimeSpent),
		AverageTimePerTask:  models.FormatTimeSpent(avgTimePerTask),
		ProductivityScore:   data.Summary.ProductivityScore,
		ActiveUsers:         data.Summary.TotalUsers,
		TopPriority:         data.Summary.TopPriority,
		MostActiveUser:      data.Summary.MostActiveUser,
	}
}

// generateExecutiveSummary creates the main executive summary text
func (sg *SummaryGenerator) generateExecutiveSummary(data *ProcessingResult, request SummaryRequest) string {
	var summary strings.Builder

	// Opening statement
	summary.WriteString(fmt.Sprintf("During the %s period from %s, the team completed %d activities with a %s completion rate. ",
		request.Period,
		data.Summary.DateRange.Label,
		data.Summary.TotalActivities,
		sg.formatPercentage(data.Summary.CompletionRate),
	))

	// Time investment
	summary.WriteString(fmt.Sprintf("A total of %s was invested across %d team members, with an average of %s per task. ",
		models.FormatTimeSpent(data.Summary.TotalTimeSpent),
		data.Summary.TotalUsers,
		models.FormatTimeSpent(data.Summary.AverageTimePerUser),
	))

	// Productivity assessment
	productivityLevel := sg.getProductivityLevel(data.Summary.ProductivityScore)
	summary.WriteString(fmt.Sprintf("The team achieved a %s productivity score of %s, indicating %s performance. ",
		productivityLevel,
		sg.formatPercentage(data.Summary.ProductivityScore),
		productivityLevel,
	))

	// Priority focus
	if data.Summary.TopPriority != "" {
		summary.WriteString(fmt.Sprintf("The primary focus was on %s priority items, ",
			strings.ToLower(data.Summary.TopPriority),
		))
	}

	// Most active contributor
	if data.Summary.MostActiveUser != "" {
		if userMetrics, exists := data.UserMetrics[data.Summary.MostActiveUser]; exists {
			summary.WriteString(fmt.Sprintf("with %s leading the effort by contributing %s across %d activities.",
				userMetrics.DisplayName,
				models.FormatTimeSpent(userMetrics.TotalTimeSpent),
				userMetrics.TotalActivities,
			))
		}
	}

	return summary.String()
}

// generateHighlights creates a list of positive highlights
func (sg *SummaryGenerator) generateHighlights(data *ProcessingResult) []string {
	highlights := []string{}

	// High completion rate
	if data.Summary.CompletionRate >= 80 {
		highlights = append(highlights, fmt.Sprintf("Excellent completion rate of %s demonstrates strong execution capability",
			sg.formatPercentage(data.Summary.CompletionRate)))
	}

	// High productivity score
	if data.Summary.ProductivityScore >= 75 {
		highlights = append(highlights, fmt.Sprintf("Strong productivity score of %s indicates effective team performance",
			sg.formatPercentage(data.Summary.ProductivityScore)))
	}

	// High-priority focus
	if highPriorityMetrics, exists := data.PriorityBreakdown["High"]; exists {
		if highPriorityMetrics.CompletionRate >= 75 {
			highlights = append(highlights, fmt.Sprintf("High-priority items completed at %s rate, showing good prioritization",
				sg.formatPercentage(highPriorityMetrics.CompletionRate)))
		}
	}

	// Top performers
	topPerformers := sg.getTopPerformers(data.UserMetrics, 2)
	if len(topPerformers) > 0 {
		names := make([]string, len(topPerformers))
		for i, user := range topPerformers {
			names[i] = user.DisplayName
		}
		highlights = append(highlights, fmt.Sprintf("Outstanding contributions from %s with consistently high performance",
			strings.Join(names, " and ")))
	}

	// Trend improvements
	if data.TrendAnalysis != nil {
		if data.TrendAnalysis.OverallTrend == "increasing" {
			highlights = append(highlights, "Positive trend in overall team performance and productivity")
		}
		if data.TrendAnalysis.VelocityTrend == "increasing" {
			highlights = append(highlights, "Improving velocity indicates enhanced team efficiency")
		}
	}

	return highlights
}

// generateConcerns creates a list of areas needing attention
func (sg *SummaryGenerator) generateConcerns(data *ProcessingResult) []string {
	concerns := []string{}

	// Low completion rate
	if data.Summary.CompletionRate < 60 {
		concerns = append(concerns, fmt.Sprintf("Completion rate of %s is below optimal levels and requires attention",
			sg.formatPercentage(data.Summary.CompletionRate)))
	}

	// Low productivity score
	if data.Summary.ProductivityScore < 50 {
		concerns = append(concerns, fmt.Sprintf("Productivity score of %s indicates potential process inefficiencies",
			sg.formatPercentage(data.Summary.ProductivityScore)))
	}

	// High-priority backlog
	if highPriorityMetrics, exists := data.PriorityBreakdown["High"]; exists {
		if highPriorityMetrics.CompletionRate < 60 {
			concerns = append(concerns, fmt.Sprintf("High-priority items only %s completed, potentially impacting critical objectives",
				sg.formatPercentage(highPriorityMetrics.CompletionRate)))
		}
	}

	// Underperforming users
	underPerformers := sg.getUnderPerformers(data.UserMetrics)
	if len(underPerformers) > 0 {
		concerns = append(concerns, fmt.Sprintf("%d team members showing below-average performance metrics",
			len(underPerformers)))
	}

	// Declining trends
	if data.TrendAnalysis != nil {
		if data.TrendAnalysis.OverallTrend == "decreasing" {
			concerns = append(concerns, "Declining trend in overall team performance requires investigation")
		}
		if data.TrendAnalysis.VelocityTrend == "decreasing" {
			concerns = append(concerns, "Decreasing velocity trend may indicate capacity or process issues")
		}
	}

	// Workload imbalance
	if sg.hasWorkloadImbalance(data.UserMetrics) {
		concerns = append(concerns, "Uneven workload distribution may lead to burnout and reduced efficiency")
	}

	return concerns
}

// generateRecommendations creates actionable recommendations
func (sg *SummaryGenerator) generateRecommendations(data *ProcessingResult) []string {
	recommendations := []string{}

	// Based on completion rate
	if data.Summary.CompletionRate < 70 {
		recommendations = append(recommendations, "Implement daily standups and sprint reviews to improve task completion tracking")
		recommendations = append(recommendations, "Consider reducing work-in-progress limits to focus on completing current tasks")
	}

	// Based on productivity score
	if data.Summary.ProductivityScore < 60 {
		recommendations = append(recommendations, "Conduct process review to identify and eliminate bottlenecks in the workflow")
		recommendations = append(recommendations, "Provide additional training or resources to team members with lower productivity scores")
	}

	// Based on priority distribution
	if sg.hasHighPriorityBacklog(data.PriorityBreakdown) {
		recommendations = append(recommendations, "Prioritize high-priority items and consider resource reallocation")
		recommendations = append(recommendations, "Review and refine prioritization process to ensure critical work gets adequate attention")
	}

	// Based on workload distribution
	if sg.hasWorkloadImbalance(data.UserMetrics) {
		recommendations = append(recommendations, "Redistribute workload to balance team capacity and prevent burnout")
		recommendations = append(recommendations, "Cross-train team members to provide better coverage and flexibility")
	}

	// Based on trends
	if data.TrendAnalysis != nil && data.TrendAnalysis.OverallTrend == "decreasing" {
		recommendations = append(recommendations, "Investigate root causes of declining performance trends")
		recommendations = append(recommendations, "Implement regular retrospectives to identify improvement opportunities")
	}

	// General recommendations
	recommendations = append(recommendations, "Continue monitoring key metrics and adjust strategies based on performance data")
	recommendations = append(recommendations, "Recognize and celebrate high performers to maintain team motivation")

	return recommendations
}

// generateUserInsights creates insights for individual users
func (sg *SummaryGenerator) generateUserInsights(data *ProcessingResult, maxUsers int) []UserInsight {
	insights := []UserInsight{}

	// Sort users by productivity rank
	userList := make([]UserMetrics, 0, len(data.UserMetrics))
	for _, user := range data.UserMetrics {
		userList = append(userList, user)
	}

	sort.Slice(userList, func(i, j int) bool {
		return userList[i].ProductivityRank < userList[j].ProductivityRank
	})

	// Limit to maxUsers if specified
	if maxUsers > 0 && len(userList) > maxUsers {
		userList = userList[:maxUsers]
	}

	for _, user := range userList {
		insight := UserInsight{
			UserID:           user.UserID,
			DisplayName:      user.DisplayName,
			ProductivityRank: user.ProductivityRank,
			CompletionRate:   user.CompletionRate,
			TotalActivities:  user.TotalActivities,
			TimeSpent:        models.FormatTimeSpent(user.TotalTimeSpent),
		}

		// Generate achievements
		insight.KeyAchievements = sg.generateUserAchievements(user)

		// Generate improvement areas
		insight.AreasForImprovement = sg.generateUserImprovements(user)

		insights = append(insights, insight)
	}

	return insights
}

// generateTrendAnalysis creates trend analysis summary
func (sg *SummaryGenerator) generateTrendAnalysis(trends *TrendAnalysis) *SummaryTrendAnalysis {
	keyChanges := []string{}

	if trends.OverallTrend == "increasing" {
		keyChanges = append(keyChanges, "Overall team performance showing positive improvement")
	} else if trends.OverallTrend == "decreasing" {
		keyChanges = append(keyChanges, "Overall team performance showing concerning decline")
	}

	if trends.VelocityTrend == "increasing" {
		keyChanges = append(keyChanges, "Team velocity increasing, indicating improved efficiency")
	} else if trends.VelocityTrend == "decreasing" {
		keyChanges = append(keyChanges, "Team velocity decreasing, may indicate capacity issues")
	}

	return &SummaryTrendAnalysis{
		OverallTrend:      trends.OverallTrend,
		VelocityTrend:     trends.VelocityTrend,
		ProductivityTrend: trends.ProductivityTrend,
		KeyChanges:        keyChanges,
		Seasonality:       trends.Seasonality,
	}
}

// generateCustomSection creates content for custom sections
func (sg *SummaryGenerator) generateCustomSection(data *ProcessingResult, section string) string {
	switch strings.ToLower(section) {
	case "priority_breakdown":
		return sg.generatePriorityBreakdown(data.PriorityBreakdown)
	case "status_summary":
		return sg.generateStatusSummary(data.StatusBreakdown)
	case "velocity_analysis":
		if data.VelocityMetrics != nil {
			return sg.generateVelocityAnalysis(data.VelocityMetrics)
		}
		return "Velocity analysis not available"
	case "time_analysis":
		return sg.generateTimeAnalysis(data)
	default:
		return fmt.Sprintf("Custom section '%s' not implemented", section)
	}
}

// Helper methods

func (sg *SummaryGenerator) formatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func (sg *SummaryGenerator) getProductivityLevel(score float64) string {
	if score >= 80 {
		return "excellent"
	} else if score >= 60 {
		return "good"
	} else if score >= 40 {
		return "average"
	} else {
		return "concerning"
	}
}

func (sg *SummaryGenerator) getTopPerformers(userMetrics map[string]UserMetrics, limit int) []UserMetrics {
	users := make([]UserMetrics, 0, len(userMetrics))
	for _, user := range userMetrics {
		if user.CompletionRate >= 80 && user.ProductivityRank <= 3 {
			users = append(users, user)
		}
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ProductivityRank < users[j].ProductivityRank
	})

	if len(users) > limit {
		users = users[:limit]
	}

	return users
}

func (sg *SummaryGenerator) getUnderPerformers(userMetrics map[string]UserMetrics) []UserMetrics {
	users := make([]UserMetrics, 0)
	for _, user := range userMetrics {
		if user.CompletionRate < 50 || user.ProductivityRank > len(userMetrics)*3/4 {
			users = append(users, user)
		}
	}
	return users
}

func (sg *SummaryGenerator) hasWorkloadImbalance(userMetrics map[string]UserMetrics) bool {
	if len(userMetrics) < 2 {
		return false
	}

	var totalTime int64
	var maxTime int64
	var minTime int64

	first := true
	for _, user := range userMetrics {
		totalTime += user.TotalTimeSpent
		if first {
			maxTime = user.TotalTimeSpent
			minTime = user.TotalTimeSpent
			first = false
		} else {
			if user.TotalTimeSpent > maxTime {
				maxTime = user.TotalTimeSpent
			}
			if user.TotalTimeSpent < minTime {
				minTime = user.TotalTimeSpent
			}
		}
	}

	avgTime := totalTime / int64(len(userMetrics))
	
	// Consider imbalanced if max is more than 2x the average or min is less than 50% of average
	return maxTime > avgTime*2 || minTime < avgTime/2
}

func (sg *SummaryGenerator) hasHighPriorityBacklog(priorityBreakdown map[string]PriorityMetrics) bool {
	if highPriority, exists := priorityBreakdown["High"]; exists {
		return highPriority.CompletionRate < 70
	}
	return false
}

func (sg *SummaryGenerator) generateUserAchievements(user UserMetrics) []string {
	achievements := []string{}

	if user.CompletionRate >= 90 {
		achievements = append(achievements, "Exceptional completion rate above 90%")
	}

	if user.ProductivityRank <= 2 {
		achievements = append(achievements, "Top performer in team productivity rankings")
	}

	if len(user.TopIssues) > 0 {
		achievements = append(achievements, fmt.Sprintf("Successfully handled %d complex, high-impact issues", len(user.TopIssues)))
	}

	return achievements
}

func (sg *SummaryGenerator) generateUserImprovements(user UserMetrics) []string {
	improvements := []string{}

	if user.CompletionRate < 60 {
		improvements = append(improvements, "Focus on improving task completion rate")
	}

	if user.AverageTimePerTask > 14400 { // More than 4 hours
		improvements = append(improvements, "Consider breaking down large tasks into smaller, manageable pieces")
	}

	return improvements
}

func (sg *SummaryGenerator) generatePriorityBreakdown(breakdown map[string]PriorityMetrics) string {
	var content strings.Builder
	content.WriteString("Priority Distribution Analysis:\n")

	priorities := []string{"High", "Medium", "Low"}
	for _, priority := range priorities {
		if metrics, exists := breakdown[priority]; exists {
			content.WriteString(fmt.Sprintf("- %s Priority: %d items (%s completion rate, %s total time)\n",
				priority,
				metrics.Count,
				sg.formatPercentage(metrics.CompletionRate),
				models.FormatTimeSpent(metrics.TotalTimeSpent),
			))
		}
	}

	return content.String()
}

func (sg *SummaryGenerator) generateStatusSummary(breakdown map[string]StatusMetrics) string {
	var content strings.Builder
	content.WriteString("Status Distribution Summary:\n")

	for status, metrics := range breakdown {
		content.WriteString(fmt.Sprintf("- %s: %d items (%s total time, %d recent changes)\n",
			status,
			metrics.Count,
			models.FormatTimeSpent(metrics.TotalTimeSpent),
			metrics.RecentChanges,
		))
	}

	return content.String()
}

func (sg *SummaryGenerator) generateVelocityAnalysis(velocity *VelocityMetrics) string {
	var content strings.Builder
	content.WriteString("Velocity Analysis:\n")
	content.WriteString(fmt.Sprintf("- Current Velocity: %.2f items/day\n", velocity.CurrentVelocity))
	content.WriteString(fmt.Sprintf("- Average Velocity: %.2f items/day\n", velocity.AverageVelocity))
	content.WriteString(fmt.Sprintf("- Velocity Trend: %s\n", velocity.VelocityTrend))
	content.WriteString(fmt.Sprintf("- Burndown Rate: %.1f%%\n", velocity.BurndownRate*100))

	return content.String()
}

func (sg *SummaryGenerator) generateTimeAnalysis(data *ProcessingResult) string {
	var content strings.Builder
	content.WriteString("Time Investment Analysis:\n")
	content.WriteString(fmt.Sprintf("- Total Time Invested: %s\n", models.FormatTimeSpent(data.Summary.TotalTimeSpent)))
	content.WriteString(fmt.Sprintf("- Average Time per User: %s\n", models.FormatTimeSpent(data.Summary.AverageTimePerUser)))
	content.WriteString(fmt.Sprintf("- Average Time per Task: %s\n", models.FormatTimeSpent(data.Summary.TotalTimeSpent/int64(data.Summary.TotalActivities))))
	content.WriteString(fmt.Sprintf("- Period: %s\n", data.Summary.DateRange.Label))

	return content.String()
}