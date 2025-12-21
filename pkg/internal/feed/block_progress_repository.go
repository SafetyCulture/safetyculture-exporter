package feed

import (
	"fmt"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlockProgressRepository handles database operations for block progress
type BlockProgressRepository struct {
	db *gorm.DB
}

// NewBlockProgressRepository creates a new repository instance
func NewBlockProgressRepository(db *gorm.DB) *BlockProgressRepository {
	return &BlockProgressRepository{db: db}
}

// InitializeBlocks creates or updates block progress records for a feed
// Returns true if blocks were newly created (first run), false if resuming
func (r *BlockProgressRepository) InitializeBlocks(
	feedName string,
	orgID string,
	blocks []util.TimeBlock,
	blockSize string,
) (isFirstRun bool, err error) {

	// Check if any blocks already exist for this feed+org combination
	var existingCount int64
	err = r.db.Model(&BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Count(&existingCount).Error
	if err != nil {
		return false, fmt.Errorf("check existing blocks: %w", err)
	}

	// If blocks exist, check if block size has changed
	if existingCount > 0 {
		var existingBlockSize string
		err = r.db.Model(&BlockProgress{}).
			Select("block_size").
			Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
			Limit(1).
			Pluck("block_size", &existingBlockSize).Error
		if err != nil {
			return false, fmt.Errorf("check block size: %w", err)
		}

		// Block size changed - delete ALL blocks and start fresh
		if existingBlockSize != blockSize {
			err = r.db.
				Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
				Delete(&BlockProgress{}).Error
			if err != nil {
				return false, fmt.Errorf("cleanup blocks after size change: %w", err)
			}
			existingCount = 0
		}
	}

	isFirstRun = existingCount == 0

	// Create block records
	var records []BlockProgress
	for _, block := range blocks {
		records = append(records, BlockProgress{
			FeedName:       feedName,
			OrganisationID: orgID,
			BlockStart:     block.Start,
			BlockEnd:       block.End,
			BlockSize:      blockSize,
			Status:         BlockStatusPending,
			RetryCount:     0,
		})
	}

	// Use upsert to handle both first run and resume scenarios
	// On conflict (same feed_name + org_id + block_start), do nothing
	err = r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "feed_name"},
			{Name: "organisation_id"},
			{Name: "block_start"},
		},
		DoNothing: true,
	}).Create(&records).Error

	if err != nil {
		return false, fmt.Errorf("create block records: %w", err)
	}

	return isFirstRun, nil
}

// GetNextPendingBlock retrieves the next block to process (pending or failed with retries remaining)
// Uses optimistic locking approach that works across all database backends
func (r *BlockProgressRepository) GetNextPendingBlock(
	feedName string,
	orgID string,
	maxRetries int,
) (*BlockProgress, error) {

	// Find a pending or failed block (without locking)
	var block BlockProgress
	err := r.db.
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Where("status IN ?", []BlockProgressStatus{BlockStatusPending, BlockStatusFailed}).
		Where("retry_count < ?", maxRetries).
		Order("block_start ASC").
		Limit(1).
		First(&block).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No more blocks to process
		}
		return nil, fmt.Errorf("get next block: %w", err)
	}

	// Atomically claim the block by updating its status
	// Only succeeds if the block is still in pending/failed status (optimistic locking)
	result := r.db.Model(&BlockProgress{}).
		Where("id = ?", block.ID).
		Where("status IN ?", []BlockProgressStatus{BlockStatusPending, BlockStatusFailed}).
		Updates(map[string]interface{}{
			"status":     BlockStatusInProgress,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return nil, fmt.Errorf("claim block: %w", result.Error)
	}

	// If no rows were updated, another worker claimed it - try again
	if result.RowsAffected == 0 {
		return r.GetNextPendingBlock(feedName, orgID, maxRetries)
	}

	// Update the block's status field to reflect the database state
	block.Status = BlockStatusInProgress
	block.UpdatedAt = time.Now()

	return &block, nil
}

// MarkBlockInProgress updates block status to in_progress
func (r *BlockProgressRepository) MarkBlockInProgress(blockID uint) error {
	return r.db.Model(&BlockProgress{}).
		Where("id = ?", blockID).
		Updates(map[string]interface{}{
			"status":     BlockStatusInProgress,
			"updated_at": time.Now(),
		}).Error
}

// MarkBlockCompleted marks a block as successfully completed
func (r *BlockProgressRepository) MarkBlockCompleted(blockID uint) error {
	now := time.Now()
	return r.db.Model(&BlockProgress{}).
		Where("id = ?", blockID).
		Updates(map[string]interface{}{
			"status":       BlockStatusCompleted,
			"completed_at": &now,
			"last_error":   "",
			"updated_at":   now,
		}).Error
}

// MarkBlockFailed marks a block as failed and increments retry count
func (r *BlockProgressRepository) MarkBlockFailed(blockID uint, errMsg string) error {
	// Truncate error message to 1000 chars to fit in database
	if len(errMsg) > 1000 {
		errMsg = errMsg[:1000]
	}

	return r.db.Model(&BlockProgress{}).
		Where("id = ?", blockID).
		Updates(map[string]interface{}{
			"status":      BlockStatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  errMsg,
			"updated_at":  time.Now(),
		}).Error
}

// GetProgressStats returns statistics about block processing
func (r *BlockProgressRepository) GetProgressStats(feedName string, orgID string) (total, completed, failed, pending int64, err error) {
	baseWhere := r.db.Model(&BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID)

	if err = baseWhere.Count(&total).Error; err != nil {
		return
	}

	completedQuery := r.db.Model(&BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Where("status = ?", BlockStatusCompleted)
	if err = completedQuery.Count(&completed).Error; err != nil {
		return
	}

	failedQuery := r.db.Model(&BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Where("status = ?", BlockStatusFailed)
	if err = failedQuery.Count(&failed).Error; err != nil {
		return
	}

	// Pending includes both pending and in_progress blocks
	pendingQuery := r.db.Model(&BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Where("status IN ?", []BlockProgressStatus{BlockStatusPending, BlockStatusInProgress})
	if err = pendingQuery.Count(&pending).Error; err != nil {
		return
	}

	return
}

// DeleteCompletedBlocks removes completed blocks (for cleanup)
func (r *BlockProgressRepository) DeleteCompletedBlocks(feedName string, orgID string) error {
	return r.db.
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Where("status = ?", BlockStatusCompleted).
		Delete(&BlockProgress{}).Error
}

// HasInProgressBlocks checks if there are any blocks currently being processed
func (r *BlockProgressRepository) HasInProgressBlocks(feedName string, orgID string) (bool, error) {
	var count int64
	err := r.db.Model(&BlockProgress{}).
		Where("feed_name = ? AND organisation_id = ?", feedName, orgID).
		Where("status = ?", BlockStatusInProgress).
		Count(&count).Error

	return count > 0, err
}
