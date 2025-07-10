package processor

import (
	"context"
	"testing"
	"time"

	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSummaryGenerator(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	assert.NotNil(t, generator)
	assert.Equal(t, logger, generator.logger)
}

func TestSummaryGenerator_GenerateSummary(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)
	ctx := context.Background()

	// Create test data
	testData := createTestProcessingResult()

	t.Run("Basic Summary Generation", func(t *testing.T) {
		request := SummaryRequest{
			Title:          "Weekly Team Summary",
			Period:         "weekly",
			IncludeMetrics: true,
			IncludeTrends:  false,
			IncludeUsers:   true,
			MaxUsers:       5,
			Format:         FormatExecutive,
		}

		summary, err := generator.GenerateSummary(ctx, testData, request)

		require.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, "Weekly Team Summary", summary.Title)
		assert.Equal(t, "weekly", summary.Period)
		assert.NotEmpty(t, summary.ExecutiveSummary)
		assert.True(t, len(summary.Highlights) > 0)
		assert.True(t, len(summary.Recommendations) > 0)
		assert.True(t, len(summary.UserInsights) > 0)
		assert.Nil(t, summary.TrendAnalysis) // Not requested
	})

	t.Run("Summary with Trends", func(t *testing.T) {
		request := SummaryRequest{
			Title:         "Monthly Analysis",
			Period:        "monthly",
			IncludeTrends: true,
			Format:        FormatDetailed,
		}

		summary, err := generator.GenerateSummary(ctx, testData, request)

		require.NoError(t, err)
		assert.NotNil(t, summary.TrendAnalysis)
		assert.NotNil(t, summary.RawData) // Included in detailed format
		assert.Equal(t, testData.TrendAnalysis.OverallTrend, summary.TrendAnalysis.OverallTrend)
	})

	t.Run("Custom Sections", func(t *testing.T) {
		request := SummaryRequest{
			Title:          "Custom Report",
			Period:         "quarterly",
			CustomSections: []string{"priority_breakdown", "status_summary"},
			Format:         FormatExecutive,
		}

		summary, err := generator.GenerateSummary(ctx, testData, request)

		require.NoError(t, err)
		assert.Contains(t, summary.Sections, "priority_breakdown")
		assert.Contains(t, summary.Sections, "status_summary")
		assert.NotEmpty(t, summary.Sections["priority_breakdown"])
		assert.NotEmpty(t, summary.Sections["status_summary"])
	})

	t.Run("Error Cases", func(t *testing.T) {
		request := SummaryRequest{
			Title:  "Test",
			Period: "weekly",
		}

		// Test with nil data
		_, err := generator.GenerateSummary(ctx, nil, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "processing data is required")
	})
}

func TestSummaryGenerator_GenerateKeyMetrics(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	testData := createTestProcessingResult()
	metrics := generator.generateKeyMetrics(testData)

	assert.Equal(t, testData.Summary.TotalActivities, metrics.TotalActivities)
	assert.Equal(t, testData.Summary.CompletionRate, metrics.CompletionRate)
	assert.Equal(t, testData.Summary.ProductivityScore, metrics.ProductivityScore)
	assert.Equal(t, testData.Summary.TotalUsers, metrics.ActiveUsers)
	assert.Equal(t, testData.Summary.TopPriority, metrics.TopPriority)
	assert.Equal(t, testData.Summary.MostActiveUser, metrics.MostActiveUser)
	assert.NotEmpty(t, metrics.TotalTimeSpent)
	assert.NotEmpty(t, metrics.AverageTimePerTask)
}

func TestSummaryGenerator_GenerateExecutiveSummary(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	testData := createTestProcessingResult()
	request := SummaryRequest{
		Period: "weekly",
		Format: FormatExecutive,
	}

	summary := generator.generateExecutiveSummary(testData, request)

	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "weekly period")
	assert.Contains(t, summary, "4 activities")
	assert.Contains(t, summary, "75.0% completion rate")
	assert.Contains(t, summary, "2 team members")
}

func TestSummaryGenerator_GenerateHighlights(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	t.Run("High Performance Data", func(t *testing.T) {
		testData := createTestProcessingResult()
		// Set high performance metrics
		testData.Summary.CompletionRate = 85.0
		testData.Summary.ProductivityScore = 80.0

		highlights := generator.generateHighlights(testData)

		assert.True(t, len(highlights) > 0)
		assert.Contains(t, highlights[0], "Excellent completion rate")
		assert.Contains(t, highlights[1], "Strong productivity score")
	})

	t.Run("With Trend Analysis", func(t *testing.T) {
		testData := createTestProcessingResult()
		testData.TrendAnalysis.OverallTrend = "increasing"
		testData.TrendAnalysis.VelocityTrend = "increasing"

		highlights := generator.generateHighlights(testData)

		assert.True(t, len(highlights) > 0)
		// Should include trend-related highlights
		found := false
		for _, highlight := range highlights {
			if contains(highlight, "Positive trend") || contains(highlight, "Improving velocity") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestSummaryGenerator_GenerateConcerns(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	t.Run("Low Performance Data", func(t *testing.T) {
		testData := createTestProcessingResult()
		// Set low performance metrics
		testData.Summary.CompletionRate = 45.0
		testData.Summary.ProductivityScore = 35.0

		concerns := generator.generateConcerns(testData)

		assert.True(t, len(concerns) > 0)
		assert.Contains(t, concerns[0], "below optimal levels")
		assert.Contains(t, concerns[1], "process inefficiencies")
	})

	t.Run("With Declining Trends", func(t *testing.T) {
		testData := createTestProcessingResult()
		testData.TrendAnalysis.OverallTrend = "decreasing"
		testData.TrendAnalysis.VelocityTrend = "decreasing"

		concerns := generator.generateConcerns(testData)

		assert.True(t, len(concerns) > 0)
		// Should include trend-related concerns
		found := false
		for _, concern := range concerns {
			if contains(concern, "Declining trend") || contains(concern, "Decreasing velocity") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestSummaryGenerator_GenerateRecommendations(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	t.Run("Low Performance Recommendations", func(t *testing.T) {
		testData := createTestProcessingResult()
		// Set low performance metrics
		testData.Summary.CompletionRate = 50.0
		testData.Summary.ProductivityScore = 40.0

		recommendations := generator.generateRecommendations(testData)

		assert.True(t, len(recommendations) > 0)
		// Should include process improvement recommendations
		found := false
		for _, rec := range recommendations {
			if contains(rec, "standups") || contains(rec, "process review") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("Always Include General Recommendations", func(t *testing.T) {
		testData := createTestProcessingResult()

		recommendations := generator.generateRecommendations(testData)

		assert.True(t, len(recommendations) > 0)
		// Should always include monitoring and recognition
		found := false
		for _, rec := range recommendations {
			if contains(rec, "monitoring key metrics") || contains(rec, "celebrate high performers") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestSummaryGenerator_GenerateUserInsights(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	testData := createTestProcessingResult()

	t.Run("All Users", func(t *testing.T) {
		insights := generator.generateUserInsights(testData, 0)

		assert.Len(t, insights, 2) // Should have 2 users
		assert.Equal(t, "user1", insights[0].UserID)
		assert.Equal(t, "User One", insights[0].DisplayName)
		assert.NotEmpty(t, insights[0].TimeSpent)
	})

	t.Run("Limited Users", func(t *testing.T) {
		insights := generator.generateUserInsights(testData, 1)

		assert.Len(t, insights, 1) // Should be limited to 1
		assert.Equal(t, "user1", insights[0].UserID) // Should be top performer
	})

	t.Run("User Achievements", func(t *testing.T) {
		// Create modified user metrics for testing
		modifiedUserMetrics := make(map[string]UserMetrics)
		for k, v := range testData.UserMetrics {
			modifiedUserMetrics[k] = v
		}
		
		// Set high performance for user1
		user1 := modifiedUserMetrics["user1"]
		user1.CompletionRate = 95.0
		user1.ProductivityRank = 1
		modifiedUserMetrics["user1"] = user1
		
		modifiedData := *testData
		modifiedData.UserMetrics = modifiedUserMetrics

		insights := generator.generateUserInsights(&modifiedData, 1)

		assert.Len(t, insights, 1)
		assert.True(t, len(insights[0].KeyAchievements) > 0)
	})
}

func TestSummaryGenerator_GenerateTrendAnalysis(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	trendData := &TrendAnalysis{
		OverallTrend:      "increasing",
		VelocityTrend:     "stable",
		ProductivityTrend: "increasing",
		Seasonality:       map[string]float64{"Monday": 0.8, "Friday": 0.6},
	}

	analysis := generator.generateTrendAnalysis(trendData)

	assert.NotNil(t, analysis)
	assert.Equal(t, "increasing", analysis.OverallTrend)
	assert.Equal(t, "stable", analysis.VelocityTrend)
	assert.Equal(t, "increasing", analysis.ProductivityTrend)
	assert.True(t, len(analysis.KeyChanges) > 0)
	assert.Equal(t, trendData.Seasonality, analysis.Seasonality)
}

func TestSummaryGenerator_GenerateCustomSection(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	testData := createTestProcessingResult()

	t.Run("Priority Breakdown", func(t *testing.T) {
		content := generator.generateCustomSection(testData, "priority_breakdown")

		assert.NotEmpty(t, content)
		assert.Contains(t, content, "Priority Distribution Analysis")
		assert.Contains(t, content, "High Priority")
		assert.Contains(t, content, "Medium Priority")
	})

	t.Run("Status Summary", func(t *testing.T) {
		content := generator.generateCustomSection(testData, "status_summary")

		assert.NotEmpty(t, content)
		assert.Contains(t, content, "Status Distribution Summary")
		assert.Contains(t, content, "Done")
		assert.Contains(t, content, "In Progress")
	})

	t.Run("Velocity Analysis", func(t *testing.T) {
		content := generator.generateCustomSection(testData, "velocity_analysis")

		assert.NotEmpty(t, content)
		assert.Contains(t, content, "Velocity Analysis")
		assert.Contains(t, content, "Current Velocity")
		assert.Contains(t, content, "Average Velocity")
	})

	t.Run("Time Analysis", func(t *testing.T) {
		content := generator.generateCustomSection(testData, "time_analysis")

		assert.NotEmpty(t, content)
		assert.Contains(t, content, "Time Investment Analysis")
		assert.Contains(t, content, "Total Time Invested")
		assert.Contains(t, content, "Average Time per User")
	})

	t.Run("Unknown Section", func(t *testing.T) {
		content := generator.generateCustomSection(testData, "unknown_section")

		assert.NotEmpty(t, content)
		assert.Contains(t, content, "not implemented")
	})
}

func TestSummaryGenerator_HelperMethods(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	t.Run("FormatPercentage", func(t *testing.T) {
		assert.Equal(t, "75.0%", generator.formatPercentage(75.0))
		assert.Equal(t, "100.0%", generator.formatPercentage(100.0))
		assert.Equal(t, "0.0%", generator.formatPercentage(0.0))
	})

	t.Run("GetProductivityLevel", func(t *testing.T) {
		assert.Equal(t, "excellent", generator.getProductivityLevel(85.0))
		assert.Equal(t, "good", generator.getProductivityLevel(65.0))
		assert.Equal(t, "average", generator.getProductivityLevel(45.0))
		assert.Equal(t, "concerning", generator.getProductivityLevel(25.0))
	})

	t.Run("GetTopPerformers", func(t *testing.T) {
		testData := createTestProcessingResult()
		performers := generator.getTopPerformers(testData.UserMetrics, 2)

		assert.True(t, len(performers) <= 2)
		for _, performer := range performers {
			assert.True(t, performer.CompletionRate >= 80 && performer.ProductivityRank <= 3)
		}
	})

	t.Run("GetUnderPerformers", func(t *testing.T) {
		testData := createTestProcessingResult()
		// Create modified user metrics for testing
		modifiedUserMetrics := make(map[string]UserMetrics)
		for k, v := range testData.UserMetrics {
			modifiedUserMetrics[k] = v
		}
		
		// Set low performance for user2
		user2 := modifiedUserMetrics["user2"]
		user2.CompletionRate = 40.0
		user2.ProductivityRank = 2
		modifiedUserMetrics["user2"] = user2

		underPerformers := generator.getUnderPerformers(modifiedUserMetrics)

		assert.True(t, len(underPerformers) > 0)
		for _, performer := range underPerformers {
			assert.True(t, performer.CompletionRate < 50 || performer.ProductivityRank > 1)
		}
	})

	t.Run("HasWorkloadImbalance", func(t *testing.T) {
		testData := createTestProcessingResult()
		
		// Create balanced workload
		balancedUserMetrics := make(map[string]UserMetrics)
		for k, v := range testData.UserMetrics {
			balancedUserMetrics[k] = v
		}
		user1 := balancedUserMetrics["user1"]
		user1.TotalTimeSpent = 10000
		balancedUserMetrics["user1"] = user1
		user2 := balancedUserMetrics["user2"]
		user2.TotalTimeSpent = 12000
		balancedUserMetrics["user2"] = user2
		assert.False(t, generator.hasWorkloadImbalance(balancedUserMetrics))

		// Create imbalanced workload
		imbalancedUserMetrics := make(map[string]UserMetrics)
		for k, v := range testData.UserMetrics {
			imbalancedUserMetrics[k] = v
		}
		user1 = imbalancedUserMetrics["user1"]
		user1.TotalTimeSpent = 20000
		imbalancedUserMetrics["user1"] = user1
		user2 = imbalancedUserMetrics["user2"]
		user2.TotalTimeSpent = 5000
		imbalancedUserMetrics["user2"] = user2
		assert.True(t, generator.hasWorkloadImbalance(imbalancedUserMetrics))
	})

	t.Run("HasHighPriorityBacklog", func(t *testing.T) {
		testData := createTestProcessingResult()
		
		// Set high completion rate
		highCompletionBreakdown := make(map[string]PriorityMetrics)
		for k, v := range testData.PriorityBreakdown {
			highCompletionBreakdown[k] = v
		}
		highPriority := highCompletionBreakdown["High"]
		highPriority.CompletionRate = 80.0
		highCompletionBreakdown["High"] = highPriority
		assert.False(t, generator.hasHighPriorityBacklog(highCompletionBreakdown))

		// Set low completion rate
		lowCompletionBreakdown := make(map[string]PriorityMetrics)
		for k, v := range testData.PriorityBreakdown {
			lowCompletionBreakdown[k] = v
		}
		highPriority = lowCompletionBreakdown["High"]
		highPriority.CompletionRate = 50.0
		lowCompletionBreakdown["High"] = highPriority
		assert.True(t, generator.hasHighPriorityBacklog(lowCompletionBreakdown))
	})
}

func TestSummaryGenerator_UserAchievementsAndImprovements(t *testing.T) {
	logger := utils.NewMockLogger()
	generator := NewSummaryGenerator(logger)

	t.Run("User Achievements", func(t *testing.T) {
		user := UserMetrics{
			UserID:           "user1",
			DisplayName:      "Test User",
			CompletionRate:   95.0,
			ProductivityRank: 1,
			TopIssues:        []string{"ISSUE-1", "ISSUE-2"},
		}

		achievements := generator.generateUserAchievements(user)

		assert.True(t, len(achievements) > 0)
		assert.Contains(t, achievements[0], "Exceptional completion rate")
		assert.Contains(t, achievements[1], "Top performer")
		assert.Contains(t, achievements[2], "complex, high-impact issues")
	})

	t.Run("User Improvements", func(t *testing.T) {
		user := UserMetrics{
			UserID:             "user1",
			DisplayName:        "Test User",
			CompletionRate:     50.0,
			AverageTimePerTask: 18000, // 5 hours
		}

		improvements := generator.generateUserImprovements(user)

		assert.True(t, len(improvements) > 0)
		assert.Contains(t, improvements[0], "improving task completion rate")
		assert.Contains(t, improvements[1], "breaking down large tasks")
	})
}

// Helper functions for tests

func createTestProcessingResult() *ProcessingResult {
	now := time.Now()
	
	return &ProcessingResult{
		Summary: ProcessingSummary{
			TotalActivities:    4,
			TotalUsers:         2,
			TotalTimeSpent:     21600, // 6 hours
			AverageTimePerUser: 10800, // 3 hours
			DateRange: TimeRange{
				Start: now.Add(-7 * 24 * time.Hour),
				End:   now,
				Label: "2023-01-01 to 2023-01-07",
			},
			MostActiveUser:    "user1",
			TopPriority:       "High",
			CompletionRate:    75.0,
			ProductivityScore: 70.0,
		},
		UserMetrics: map[string]UserMetrics{
			"user1": {
				UserID:             "user1",
				DisplayName:        "User One",
				TotalActivities:    3,
				CompletedActivities: 2,
				TotalTimeSpent:     14400, // 4 hours
				AverageTimePerTask: 4800,  // 1.33 hours
				CompletionRate:     66.7,
				ProductivityRank:   1,
				TopIssues:          []string{"TEST-1", "TEST-3"},
				PriorityDistribution: map[string]int{"High": 2, "Medium": 1},
				StatusDistribution:   map[string]int{"Done": 2, "In Progress": 1},
			},
			"user2": {
				UserID:             "user2",
				DisplayName:        "User Two",
				TotalActivities:    1,
				CompletedActivities: 1,
				TotalTimeSpent:     7200, // 2 hours
				AverageTimePerTask: 7200, // 2 hours
				CompletionRate:     100.0,
				ProductivityRank:   2,
				TopIssues:          []string{},
				PriorityDistribution: map[string]int{"High": 1},
				StatusDistribution:   map[string]int{"Done": 1},
			},
		},
		PriorityBreakdown: map[string]PriorityMetrics{
			"High": {
				Priority:              "High",
				Count:                 3,
				TotalTimeSpent:        18000, // 5 hours
				CompletedCount:        2,
				CompletionRate:        66.7,
				AverageTimeToComplete: 9000, // 2.5 hours
			},
			"Medium": {
				Priority:              "Medium",
				Count:                 1,
				TotalTimeSpent:        3600, // 1 hour
				CompletedCount:        1,
				CompletionRate:        100.0,
				AverageTimeToComplete: 3600, // 1 hour
			},
		},
		StatusBreakdown: map[string]StatusMetrics{
			"Done": {
				Status:         "Done",
				Count:          3,
				TotalTimeSpent: 18000, // 5 hours
				Users:          []string{"user1", "user2"},
				RecentChanges:  2,
			},
			"In Progress": {
				Status:         "In Progress",
				Count:          1,
				TotalTimeSpent: 3600, // 1 hour
				Users:          []string{"user1"},
				RecentChanges:  1,
			},
		},
		TrendAnalysis: &TrendAnalysis{
			OverallTrend:      "stable",
			VelocityTrend:     "increasing",
			ProductivityTrend: "stable",
			Seasonality:       map[string]float64{"Monday": 0.8, "Tuesday": 0.7, "Wednesday": 0.9},
		},
		VelocityMetrics: &VelocityMetrics{
			CurrentVelocity: 2.5,
			AverageVelocity: 2.3,
			VelocityTrend:   "increasing",
			BurndownRate:    0.8,
			UserVelocities:  map[string]float64{"user1": 2.8, "user2": 2.2},
		},
		ProcessedAt:    now,
		ProcessingTime: 100 * time.Millisecond,
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}