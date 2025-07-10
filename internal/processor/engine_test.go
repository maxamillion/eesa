package processor

import (
	"context"
	"testing"
	"time"

	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataProcessor(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	assert.NotNil(t, processor)
	assert.Equal(t, logger, processor.logger)
}

func TestDataProcessor_ProcessActivities_EmptyInput(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	ctx := context.Background()
	
	options := ProcessingOptions{}
	result, err := processor.ProcessActivities(ctx, []models.Activity{}, options)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Summary.TotalActivities)
	assert.Equal(t, 0, result.Summary.TotalUsers)
	assert.True(t, result.ProcessingTime > 0)
}

func TestDataProcessor_ProcessActivities_BasicProcessing(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	ctx := context.Background()
	
	// Create test activities
	activities := createTestActivities()
	
	options := ProcessingOptions{
		GroupByUser:     true,
		GroupByPriority: true,
		GroupByStatus:   true,
	}
	
	result, err := processor.ProcessActivities(ctx, activities, options)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Verify summary metrics
	assert.Equal(t, 4, result.Summary.TotalActivities)
	assert.Equal(t, 2, result.Summary.TotalUsers) // user1 and user2
	assert.True(t, result.Summary.TotalTimeSpent > 0)
	assert.True(t, result.Summary.CompletionRate >= 0)
	assert.True(t, result.Summary.ProductivityScore >= 0)
	assert.NotEmpty(t, result.Summary.MostActiveUser)
	assert.NotEmpty(t, result.Summary.TopPriority)
	
	// Verify user metrics
	assert.Len(t, result.UserMetrics, 2)
	assert.Contains(t, result.UserMetrics, "user1")
	assert.Contains(t, result.UserMetrics, "user2")
	
	user1Metrics := result.UserMetrics["user1"]
	assert.Equal(t, "user1", user1Metrics.UserID)
	assert.Equal(t, "User One", user1Metrics.DisplayName)
	assert.True(t, user1Metrics.TotalActivities > 0)
	assert.True(t, user1Metrics.CompletionRate >= 0)
	
	// Verify priority breakdown
	assert.NotEmpty(t, result.PriorityBreakdown)
	assert.Contains(t, result.PriorityBreakdown, "High")
	
	highPriorityMetrics := result.PriorityBreakdown["High"]
	assert.Equal(t, "High", highPriorityMetrics.Priority)
	assert.True(t, highPriorityMetrics.Count > 0)
	
	// Verify status breakdown
	assert.NotEmpty(t, result.StatusBreakdown)
	assert.Contains(t, result.StatusBreakdown, "Done")
	
	doneStatusMetrics := result.StatusBreakdown["Done"]
	assert.Equal(t, "Done", doneStatusMetrics.Status)
	assert.True(t, doneStatusMetrics.Count > 0)
}

func TestDataProcessor_ProcessActivities_WithTrendAnalysis(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	ctx := context.Background()
	
	activities := createTestActivities()
	
	options := ProcessingOptions{
		AnalyzeTrends: true,
	}
	
	result, err := processor.ProcessActivities(ctx, activities, options)
	
	require.NoError(t, err)
	assert.NotNil(t, result.TrendAnalysis)
	assert.NotEmpty(t, result.TrendAnalysis.OverallTrend)
	assert.NotEmpty(t, result.TrendAnalysis.VelocityTrend)
	assert.NotEmpty(t, result.TrendAnalysis.ProductivityTrend)
	assert.NotEmpty(t, result.TrendAnalysis.Seasonality)
}

func TestDataProcessor_ProcessActivities_WithVelocityMetrics(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	ctx := context.Background()
	
	activities := createTestActivities()
	
	options := ProcessingOptions{
		CalculateVelocity: true,
	}
	
	result, err := processor.ProcessActivities(ctx, activities, options)
	
	require.NoError(t, err)
	assert.NotNil(t, result.VelocityMetrics)
	assert.True(t, result.VelocityMetrics.CurrentVelocity >= 0)
	assert.True(t, result.VelocityMetrics.AverageVelocity >= 0)
	assert.NotEmpty(t, result.VelocityMetrics.VelocityTrend)
	assert.NotEmpty(t, result.VelocityMetrics.UserVelocities)
}

func TestDataProcessor_ProcessActivities_WithMinimumTimeFilter(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	ctx := context.Background()
	
	activities := createTestActivities()
	
	options := ProcessingOptions{
		MinimumTimeSpent: 7200, // 2 hours
		GroupByUser:      true,
	}
	
	result, err := processor.ProcessActivities(ctx, activities, options)
	
	require.NoError(t, err)
	
	// Should filter out activities with less than 2 hours
	assert.True(t, result.Summary.TotalActivities < len(activities))
	
	// All remaining activities should have at least the minimum time
	for _, userMetrics := range result.UserMetrics {
		if userMetrics.TotalActivities > 0 {
			assert.True(t, userMetrics.AverageTimePerTask >= 7200)
		}
	}
}

func TestDataProcessor_CalculateProductivityScore(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	tests := []struct {
		name           string
		activities     []models.Activity
		completionRate float64
		expectedRange  [2]float64 // [min, max]
	}{
		{
			name:           "empty activities",
			activities:     []models.Activity{},
			completionRate: 0,
			expectedRange:  [2]float64{0, 0},
		},
		{
			name: "high completion rate with high priority",
			activities: []models.Activity{
				{
					Priority: "High",
					Status:   "Done",
				},
				{
					Priority: "High",
					Status:   "Done",
				},
			},
			completionRate: 100,
			expectedRange:  [2]float64{90, 100},
		},
		{
			name: "low completion rate",
			activities: []models.Activity{
				{
					Priority: "Low",
					Status:   "In Progress",
				},
				{
					Priority: "Low",
					Status:   "Open",
				},
			},
			completionRate: 0,
			expectedRange:  [2]float64{0, 30},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := processor.calculateProductivityScore(tt.activities, tt.completionRate)
			assert.True(t, score >= tt.expectedRange[0] && score <= tt.expectedRange[1],
				"Score %f not in expected range [%f, %f]", score, tt.expectedRange[0], tt.expectedRange[1])
		})
	}
}

func TestDataProcessor_IsCompleted(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	tests := []struct {
		status   string
		expected bool
	}{
		{"Done", true},
		{"Closed", true},
		{"Resolved", true},
		{"Complete", true},
		{"Finished", true},
		{"In Progress", false},
		{"Open", false},
		{"To Do", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := processor.isCompleted(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataProcessor_FilterActivities(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	activities := []models.Activity{
		{TimeSpent: 1800},  // 30 minutes
		{TimeSpent: 3600},  // 1 hour
		{TimeSpent: 7200},  // 2 hours
		{TimeSpent: 10800}, // 3 hours
	}
	
	tests := []struct {
		name             string
		minimumTimeSpent int64
		expectedCount    int
	}{
		{
			name:             "no filter",
			minimumTimeSpent: 0,
			expectedCount:    4,
		},
		{
			name:             "filter 1 hour minimum",
			minimumTimeSpent: 3600,
			expectedCount:    3,
		},
		{
			name:             "filter 2 hours minimum",
			minimumTimeSpent: 7200,
			expectedCount:    2,
		},
		{
			name:             "filter very high minimum",
			minimumTimeSpent: 20000,
			expectedCount:    0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := ProcessingOptions{
				MinimumTimeSpent: tt.minimumTimeSpent,
			}
			
			filtered := processor.filterActivities(activities, options)
			assert.Len(t, filtered, tt.expectedCount)
			
			// Verify all filtered activities meet the minimum
			for _, activity := range filtered {
				assert.True(t, activity.TimeSpent >= tt.minimumTimeSpent)
			}
		})
	}
}

func TestDataProcessor_CalculateUserMetrics(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	activities := []models.Activity{
		{
			Key:      "TEST-1",
			Priority: "High",
			Status:   "Done",
			TimeSpent: 7200,
			Assignee: models.User{
				AccountID:   "user1",
				DisplayName: "Test User",
			},
		},
		{
			Key:      "TEST-2",
			Priority: "Medium",
			Status:   "In Progress",
			TimeSpent: 3600,
			Assignee: models.User{
				AccountID:   "user1",
				DisplayName: "Test User",
			},
		},
	}
	
	metrics := processor.calculateUserMetrics("user1", activities)
	
	assert.Equal(t, "user1", metrics.UserID)
	assert.Equal(t, "Test User", metrics.DisplayName)
	assert.Equal(t, 2, metrics.TotalActivities)
	assert.Equal(t, 1, metrics.CompletedActivities)
	assert.Equal(t, int64(10800), metrics.TotalTimeSpent)
	assert.Equal(t, int64(5400), metrics.AverageTimePerTask)
	assert.Equal(t, 50.0, metrics.CompletionRate)
	assert.Equal(t, 1, metrics.PriorityDistribution["High"])
	assert.Equal(t, 1, metrics.PriorityDistribution["Medium"])
	assert.Equal(t, 1, metrics.StatusDistribution["Done"])
	assert.Equal(t, 1, metrics.StatusDistribution["In Progress"])
	assert.Contains(t, metrics.TopIssues, "TEST-1") // >2 hours
}

func TestDataProcessor_CalculatePriorityMetrics(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	activities := []models.Activity{
		{
			Priority:  "High",
			Status:    "Done",
			TimeSpent: 7200,
		},
		{
			Priority:  "High",
			Status:    "In Progress",
			TimeSpent: 3600,
		},
	}
	
	metrics := processor.calculatePriorityMetrics("High", activities)
	
	assert.Equal(t, "High", metrics.Priority)
	assert.Equal(t, 2, metrics.Count)
	assert.Equal(t, int64(10800), metrics.TotalTimeSpent)
	assert.Equal(t, 1, metrics.CompletedCount)
	assert.Equal(t, 50.0, metrics.CompletionRate)
	assert.Equal(t, int64(7200), metrics.AverageTimeToComplete)
}

func TestDataProcessor_CalculateStatusMetrics(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	now := time.Now()
	activities := []models.Activity{
		{
			Status:    "Done",
			TimeSpent: 7200,
			Updated:   now.Add(-1 * time.Hour), // Recent
			Assignee: models.User{
				AccountID: "user1",
			},
		},
		{
			Status:    "Done",
			TimeSpent: 3600,
			Updated:   now.Add(-10 * 24 * time.Hour), // Old
			Assignee: models.User{
				AccountID: "user2",
			},
		},
	}
	
	metrics := processor.calculateStatusMetrics("Done", activities)
	
	assert.Equal(t, "Done", metrics.Status)
	assert.Equal(t, 2, metrics.Count)
	assert.Equal(t, int64(10800), metrics.TotalTimeSpent)
	assert.Len(t, metrics.Users, 2)
	assert.Contains(t, metrics.Users, "user1")
	assert.Contains(t, metrics.Users, "user2")
	assert.Equal(t, 1, metrics.RecentChanges) // Only the recent one
}

func TestDataProcessor_GenerateWeeklyRanges(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(15 * 24 * time.Hour) // 15 days
	
	activities := []models.Activity{
		{
			Created: startDate,
			Updated: endDate,
		},
	}
	
	ranges := processor.generateWeeklyRanges(activities)
	
	assert.Len(t, ranges, 3) // Should create 3 weekly ranges for 15 days
	assert.Equal(t, "Week 1", ranges[0].Label)
	assert.Equal(t, "Week 2", ranges[1].Label)
	assert.Equal(t, "Week 3", ranges[2].Label)
	
	// Verify ranges are continuous
	assert.Equal(t, startDate, ranges[0].Start)
	assert.Equal(t, ranges[0].End, ranges[1].Start)
	assert.Equal(t, ranges[1].End, ranges[2].Start)
}

func TestDataProcessor_FilterActivitiesByTimeRange(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	
	activities := []models.Activity{
		{
			Created: baseTime.Add(1 * 24 * time.Hour),
			Updated: baseTime.Add(2 * 24 * time.Hour),
		},
		{
			Created: baseTime.Add(3 * 24 * time.Hour),
			Updated: baseTime.Add(4 * 24 * time.Hour),
		},
		{
			Created: baseTime.Add(8 * 24 * time.Hour),
			Updated: baseTime.Add(9 * 24 * time.Hour),
		},
	}
	
	timeRange := TimeRange{
		Start: baseTime,
		End:   baseTime.Add(5 * 24 * time.Hour),
		Label: "Test Range",
	}
	
	filtered := processor.filterActivitiesByTimeRange(activities, timeRange)
	
	// Should include first two activities, exclude the third
	assert.Len(t, filtered, 2)
}

func TestDataProcessor_CalculateTimeRangeMetrics(t *testing.T) {
	logger := utils.NewMockLogger()
	processor := NewDataProcessor(logger)
	
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	timeRange := TimeRange{
		Start: baseTime,
		End:   baseTime.Add(7 * 24 * time.Hour),
		Label: "Week 1",
	}
	
	activities := []models.Activity{
		{
			Status:    "Done",
			TimeSpent: 7200,
		},
		{
			Status:    "In Progress",
			TimeSpent: 3600,
		},
	}
	
	metrics := processor.calculateTimeRangeMetrics(timeRange, activities)
	
	assert.Equal(t, timeRange, metrics.Range)
	assert.Equal(t, 2, metrics.ActivityCount)
	assert.Equal(t, 1, metrics.CompletionCount)
	assert.Equal(t, int64(10800), metrics.TotalTimeSpent)
	assert.True(t, metrics.AverageVelocity > 0)
	assert.True(t, metrics.ProductivityScore > 0)
}

// Helper function to create test activities
func createTestActivities() []models.Activity {
	now := time.Now()
	
	return []models.Activity{
		{
			ID:          "1",
			Key:         "TEST-1",
			Summary:     "Test Issue 1",
			Priority:    "High",
			Status:      "Done",
			TimeSpent:   7200, // 2 hours
			Created:     now.Add(-7 * 24 * time.Hour),
			Updated:     now.Add(-1 * 24 * time.Hour),
			Assignee: models.User{
				AccountID:   "user1",
				DisplayName: "User One",
			},
		},
		{
			ID:          "2",
			Key:         "TEST-2",
			Summary:     "Test Issue 2",
			Priority:    "Medium",
			Status:      "In Progress",
			TimeSpent:   3600, // 1 hour
			Created:     now.Add(-6 * 24 * time.Hour),
			Updated:     now.Add(-2 * time.Hour),
			Assignee: models.User{
				AccountID:   "user1",
				DisplayName: "User One",
			},
		},
		{
			ID:          "3",
			Key:         "TEST-3",
			Summary:     "Test Issue 3",
			Priority:    "High",
			Status:      "Done",
			TimeSpent:   5400, // 1.5 hours
			Created:     now.Add(-5 * 24 * time.Hour),
			Updated:     now.Add(-1 * time.Hour),
			Assignee: models.User{
				AccountID:   "user2",
				DisplayName: "User Two",
			},
		},
		{
			ID:          "4",
			Key:         "TEST-4",
			Summary:     "Test Issue 4",
			Priority:    "Low",
			Status:      "Open",
			TimeSpent:   1800, // 30 minutes
			Created:     now.Add(-4 * 24 * time.Hour),
			Updated:     now.Add(-30 * time.Minute),
			Assignee: models.User{
				AccountID:   "user2",
				DisplayName: "User Two",
			},
		},
	}
}