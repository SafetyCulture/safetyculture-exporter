package feed_test

import (
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// getTestDB creates an in-memory SQLite database for testing
// Each call creates a new isolated database
func getTestDB(t *testing.T) *gorm.DB {
	// Use unique connection string for each test to avoid sharing state
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NotNil(t, db)

	// Auto-migrate the BlockProgress table
	err = db.AutoMigrate(&feed.BlockProgress{})
	require.NoError(t, err)

	return db
}

func TestBlockProgressRepository_InitializeBlocks_FirstRun(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	// Create some time blocks
	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	isFirstRun, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)
	assert.True(t, isFirstRun)

	// Verify blocks were created
	var count int64
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify all blocks are pending
	var pendingCount int64
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ? AND status = ?", feedName, orgID, feed.BlockStatusPending).
		Count(&pendingCount)
	assert.Equal(t, int64(2), pendingCount)
}

func TestBlockProgressRepository_InitializeBlocks_Resume(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	// First run
	isFirstRun, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)
	assert.True(t, isFirstRun)

	// Mark first block as completed
	var firstBlock feed.BlockProgress
	err = db.Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Order("block_start ASC").
		First(&firstBlock).Error
	require.NoError(t, err)

	err = repo.MarkBlockCompleted(firstBlock.ID)
	require.NoError(t, err)

	// Second run (resume)
	isFirstRun, err = repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)
	assert.False(t, isFirstRun)

	// Verify still only 2 blocks
	var count int64
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify one is completed, one is pending
	var completedCount int64
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ? AND status = ?", feedName, orgID, feed.BlockStatusCompleted).
		Count(&completedCount)
	assert.Equal(t, int64(1), completedCount)
}

func TestBlockProgressRepository_InitializeBlocks_BlockSizeChange(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"

	now := time.Now()
	blocks7d := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	// Initialize with 7d block size
	isFirstRun, err := repo.InitializeBlocks(feedName, orgID, blocks7d, "7d")
	require.NoError(t, err)
	assert.True(t, isFirstRun)

	// Verify 2 blocks created
	var count int64
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Count(&count)
	assert.Equal(t, int64(2), count)

	// Change to 1d block size
	blocks1d := []util.TimeBlock{
		{Start: now.Add(-3 * 24 * time.Hour), End: now.Add(-2 * 24 * time.Hour)},
		{Start: now.Add(-2 * 24 * time.Hour), End: now.Add(-1 * 24 * time.Hour)},
		{Start: now.Add(-1 * 24 * time.Hour), End: now},
	}

	isFirstRun, err = repo.InitializeBlocks(feedName, orgID, blocks1d, "1d")
	require.NoError(t, err)
	assert.True(t, isFirstRun) // Should be treated as first run after deletion

	// Verify old blocks deleted and new blocks created
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Count(&count)
	assert.Equal(t, int64(3), count)

	// Verify all are 1d blocks
	var blocks []feed.BlockProgress
	db.Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Find(&blocks)
	for _, block := range blocks {
		assert.Equal(t, "1d", block.BlockSize)
	}
}

func TestBlockProgressRepository_GetNextPendingBlock_Success(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get next pending block
	block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block)

	// Should return the earliest block
	assert.Equal(t, blocks[0].Start.Format(time.RFC3339), block.BlockStart.Format(time.RFC3339))
	assert.Equal(t, feed.BlockStatusInProgress, block.Status)
}

func TestBlockProgressRepository_GetNextPendingBlock_NoBlocksAvailable(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"

	// No blocks exist
	block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	assert.Nil(t, block)
}

func TestBlockProgressRepository_GetNextPendingBlock_SkipsInProgress(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get first block (marks as in_progress)
	firstBlock, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, firstBlock)

	// Get next block (should skip in_progress and return second block)
	secondBlock, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, secondBlock)

	assert.NotEqual(t, firstBlock.ID, secondBlock.ID)
	assert.Equal(t, blocks[1].Start.Format(time.RFC3339), secondBlock.BlockStart.Format(time.RFC3339))
}

func TestBlockProgressRepository_GetNextPendingBlock_RespectsMaxRetries(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get block and mark as failed with retry count = 2
	block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block)

	// Manually set retry count to 2 and mark as failed
	db.Model(&feed.BlockProgress{}).
		Where("id = ?", block.ID).
		Updates(map[string]interface{}{
			"status":      feed.BlockStatusFailed,
			"retry_count": 2,
		})

	// Should still get the block since retry_count (2) < maxRetries (3)
	retriedBlock, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, retriedBlock)
	assert.Equal(t, block.ID, retriedBlock.ID)

	// Set retry count to 3
	db.Model(&feed.BlockProgress{}).
		Where("id = ?", block.ID).
		Updates(map[string]interface{}{
			"status":      feed.BlockStatusFailed,
			"retry_count": 3,
		})

	// Should not get the block anymore
	retriedBlock, err = repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	assert.Nil(t, retriedBlock)
}

func TestBlockProgressRepository_MarkBlockInProgress(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get the block ID
	var block feed.BlockProgress
	err = db.Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		First(&block).Error
	require.NoError(t, err)

	// Mark as in progress
	err = repo.MarkBlockInProgress(block.ID)
	require.NoError(t, err)

	// Verify status changed
	err = db.Where("id = ?", block.ID).First(&block).Error
	require.NoError(t, err)
	assert.Equal(t, feed.BlockStatusInProgress, block.Status)
}

func TestBlockProgressRepository_MarkBlockCompleted(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get the block
	block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block)

	// Mark as completed
	err = repo.MarkBlockCompleted(block.ID)
	require.NoError(t, err)

	// Verify status changed
	var completedBlock feed.BlockProgress
	err = db.Where("id = ?", block.ID).First(&completedBlock).Error
	require.NoError(t, err)
	assert.Equal(t, feed.BlockStatusCompleted, completedBlock.Status)
	assert.NotNil(t, completedBlock.CompletedAt)
	assert.Empty(t, completedBlock.LastError)
}

func TestBlockProgressRepository_MarkBlockFailed(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get the block
	block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block)

	initialRetryCount := block.RetryCount

	// Mark as failed
	errorMsg := "API request failed: 500 Internal Server Error"
	err = repo.MarkBlockFailed(block.ID, errorMsg)
	require.NoError(t, err)

	// Verify status changed and error stored
	var failedBlock feed.BlockProgress
	err = db.Where("id = ?", block.ID).First(&failedBlock).Error
	require.NoError(t, err)
	assert.Equal(t, feed.BlockStatusFailed, failedBlock.Status)
	assert.Equal(t, errorMsg, failedBlock.LastError)
	assert.Equal(t, initialRetryCount+1, failedBlock.RetryCount)
}

func TestBlockProgressRepository_MarkBlockFailed_TruncatesLongError(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Get the block
	block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block)

	// Create a very long error message (> 1000 chars)
	longError := ""
	for i := 0; i < 1100; i++ {
		longError += "x"
	}

	// Mark as failed with long error
	err = repo.MarkBlockFailed(block.ID, longError)
	require.NoError(t, err)

	// Verify error was truncated to 1000 chars
	var failedBlock feed.BlockProgress
	err = db.Where("id = ?", block.ID).First(&failedBlock).Error
	require.NoError(t, err)
	assert.Equal(t, 1000, len(failedBlock.LastError))
}

func TestBlockProgressRepository_GetProgressStats(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-28 * 24 * time.Hour), End: now.Add(-21 * 24 * time.Hour)},
		{Start: now.Add(-21 * 24 * time.Hour), End: now.Add(-14 * 24 * time.Hour)},
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Mark some blocks with different statuses
	var allBlocks []feed.BlockProgress
	db.Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Order("block_start ASC").
		Find(&allBlocks)

	// Mark first as completed
	err = repo.MarkBlockCompleted(allBlocks[0].ID)
	require.NoError(t, err)

	// Mark second as in_progress
	err = repo.MarkBlockInProgress(allBlocks[1].ID)
	require.NoError(t, err)

	// Mark third as failed
	err = repo.MarkBlockFailed(allBlocks[2].ID, "test error")
	require.NoError(t, err)

	// Fourth remains pending

	// Get stats
	total, completed, failed, pending, err := repo.GetProgressStats(feedName, orgID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, int64(1), completed)
	assert.Equal(t, int64(1), failed)
	assert.Equal(t, int64(2), pending) // includes in_progress
}

func TestBlockProgressRepository_DeleteCompletedBlocks(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Mark first block as completed
	var firstBlock feed.BlockProgress
	err = db.Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Order("block_start ASC").
		First(&firstBlock).Error
	require.NoError(t, err)

	err = repo.MarkBlockCompleted(firstBlock.ID)
	require.NoError(t, err)

	// Delete completed blocks
	err = repo.DeleteCompletedBlocks(feedName, orgID)
	require.NoError(t, err)

	// Verify only pending block remains
	var count int64
	db.Model(&feed.BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Count(&count)
	assert.Equal(t, int64(1), count)

	var remainingBlock feed.BlockProgress
	err = db.Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		First(&remainingBlock).Error
	require.NoError(t, err)
	assert.Equal(t, feed.BlockStatusPending, remainingBlock.Status)
}

func TestBlockProgressRepository_Concurrency_MultipleWorkers(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	// Create 10 blocks
	now := time.Now()
	blocks := make([]util.TimeBlock, 10)
	for i := 0; i < 10; i++ {
		blocks[i] = util.TimeBlock{
			Start: now.Add(time.Duration(-24*(10-i)) * time.Hour),
			End:   now.Add(time.Duration(-24*(9-i)) * time.Hour),
		}
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// Sequentially claim blocks (simulating workers)
	// Note: True concurrent testing is difficult with in-memory SQLite
	// This tests that GetNextPendingBlock returns different blocks
	claimedBlocks := make([]uint, 3)
	for i := 0; i < 3; i++ {
		block, err := repo.GetNextPendingBlock(feedName, orgID, 3)
		require.NoError(t, err)
		require.NotNil(t, block)
		claimedBlocks[i] = block.ID
	}

	// Verify all claimed block IDs are unique
	seen := make(map[uint]bool)
	for _, blockID := range claimedBlocks {
		assert.False(t, seen[blockID], "Block %d was claimed multiple times", blockID)
		seen[blockID] = true
	}
}

func TestBlockProgressRepository_OptimisticLocking_SequentialClaiming(t *testing.T) {
	db := getTestDB(t)
	repo := feed.NewBlockProgressRepository(db)

	feedName := "inspections"
	orgID := "org_123"
	blockSize := "7d"

	now := time.Now()
	blocks := []util.TimeBlock{
		{Start: now.Add(-14 * 24 * time.Hour), End: now.Add(-7 * 24 * time.Hour)},
		{Start: now.Add(-7 * 24 * time.Hour), End: now},
	}

	_, err := repo.InitializeBlocks(feedName, orgID, blocks, blockSize)
	require.NoError(t, err)

	// First worker claims first block
	block1, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block1)

	// Second worker should get the second block (not the first)
	block2, err := repo.GetNextPendingBlock(feedName, orgID, 3)
	require.NoError(t, err)
	require.NotNil(t, block2)

	// Verify different blocks were claimed
	assert.NotEqual(t, block1.ID, block2.ID, "Same block was claimed twice")
}
