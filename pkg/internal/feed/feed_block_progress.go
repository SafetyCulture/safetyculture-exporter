package feed

import (
	"time"
)

// BlockProgressStatus represents the status of a time block processing
type BlockProgressStatus string

const (
	BlockStatusPending    BlockProgressStatus = "pending"
	BlockStatusInProgress BlockProgressStatus = "in_progress"
	BlockStatusCompleted  BlockProgressStatus = "completed"
	BlockStatusFailed     BlockProgressStatus = "failed"
)

// BlockProgress tracks the progress of individual time blocks during export
type BlockProgress struct {
	ID             uint                `gorm:"primarykey;autoIncrement"`
	FeedName       string              `gorm:"size:50;uniqueIndex:idx_block_progress_composite;not null"`
	OrganisationID string              `gorm:"size:37;uniqueIndex:idx_block_progress_composite;not null"`
	BlockStart     time.Time           `gorm:"uniqueIndex:idx_block_progress_composite;not null"`
	BlockEnd       time.Time           `gorm:"not null"`
	BlockSize      string              `gorm:"size:10;not null"`
	Status         BlockProgressStatus `gorm:"size:20;index:idx_block_progress_status;not null;default:'pending'"`
	RetryCount     int                 `gorm:"not null;default:0"`
	LastError      string              `gorm:"type:text"`
	CreatedAt      time.Time           `gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time           `gorm:"autoUpdateTime;not null"`
	CompletedAt    *time.Time          `gorm:"index:idx_block_progress_completed"`
}

// TableName overrides the default table name
func (BlockProgress) TableName() string {
	return "block_progress"
}
