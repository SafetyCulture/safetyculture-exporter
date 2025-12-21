package feed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"go.uber.org/zap"
)

// BlockProcessorConfig configures parallel block processing
type BlockProcessorConfig struct {
	FeedName      string
	OrgID         string
	NumWorkers    int
	MaxRetries    int
	RetryBackoff  time.Duration
	StopOnFailure bool // Stop all workers if one block fails repeatedly
}

// BlockProcessor manages parallel processing of time blocks
type BlockProcessor struct {
	config     BlockProcessorConfig
	repository *BlockProgressRepository
	logger     *zap.SugaredLogger

	// Synchronization
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	errChan     chan error
	globalError error
	errorMu     sync.Mutex
}

// NewBlockProcessor creates a new block processor
func NewBlockProcessor(config BlockProcessorConfig, repository *BlockProgressRepository) *BlockProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &BlockProcessor{
		config:     config,
		repository: repository,
		logger:     logger.GetLogger(),
		ctx:        ctx,
		cancel:     cancel,
		errChan:    make(chan error, config.NumWorkers),
	}
}

// ProcessBlocksInParallel processes blocks using a worker pool
func (bp *BlockProcessor) ProcessBlocksInParallel(
	parentCtx context.Context,
	apiClient *httpapi.Client,
	exporter Exporter,
	blockProcessFn func(context.Context, *httpapi.Client, Exporter, util.TimeBlock, int) error,
) error {

	bp.logger.With(
		"feed", bp.config.FeedName,
		"org_id", bp.config.OrgID,
		"workers", bp.config.NumWorkers,
	).Info("starting parallel block processing")

	// Start error monitor goroutine
	go bp.monitorErrors()

	// Start worker pool
	for i := 0; i < bp.config.NumWorkers; i++ {
		bp.wg.Add(1)
		go bp.worker(i, parentCtx, apiClient, exporter, blockProcessFn)
	}

	// Wait for all workers to complete
	bp.wg.Wait()
	close(bp.errChan)

	// Check for global error
	bp.errorMu.Lock()
	defer bp.errorMu.Unlock()

	if bp.globalError != nil {
		return bp.globalError
	}

	// Final stats
	total, completed, failed, pending, err := bp.repository.GetProgressStats(bp.config.FeedName, bp.config.OrgID)
	if err != nil {
		bp.logger.Warnf("failed to get progress stats: %v", err)
	} else {
		bp.logger.With(
			"total", total,
			"completed", completed,
			"failed", failed,
			"pending", pending,
		).Info("block processing complete")
	}

	if failed > 0 {
		return fmt.Errorf("%d blocks failed after max retries", failed)
	}

	return nil
}

// worker processes blocks from the queue
func (bp *BlockProcessor) worker(
	workerID int,
	parentCtx context.Context,
	apiClient *httpapi.Client,
	exporter Exporter,
	blockProcessFn func(context.Context, *httpapi.Client, Exporter, util.TimeBlock, int) error,
) {
	defer bp.wg.Done()

	logger := bp.logger.With("worker", workerID)
	logger.Debug("worker started")

	for {
		// Check for cancellation
		select {
		case <-bp.ctx.Done():
			logger.Info("worker stopping due to cancellation")
			return
		case <-parentCtx.Done():
			logger.Info("worker stopping due to parent context cancellation")
			return
		default:
		}

		// Get next block to process
		block, err := bp.repository.GetNextPendingBlock(
			bp.config.FeedName,
			bp.config.OrgID,
			bp.config.MaxRetries,
		)
		if err != nil {
			bp.reportError(fmt.Errorf("worker %d: get next block: %w", workerID, err))
			return
		}

		// No more blocks - worker done
		if block == nil {
			logger.Debug("no more blocks to process")
			return
		}

		// Mark block as in progress
		if err := bp.repository.MarkBlockInProgress(block.ID); err != nil {
			bp.reportError(fmt.Errorf("worker %d: mark block in progress: %w", workerID, err))
			return
		}

		logger.With(
			"block_id", block.ID,
			"block_start", block.BlockStart.Format(time.RFC3339),
			"block_end", block.BlockEnd.Format(time.RFC3339),
			"retry", block.RetryCount,
		).Info("processing block")

		// Process the block
		timeBlock := util.TimeBlock{
			Start: block.BlockStart,
			End:   block.BlockEnd,
		}

		err = blockProcessFn(parentCtx, apiClient, exporter, timeBlock, block.RetryCount)

		if err != nil {
			// Mark as failed
			errMsg := err.Error()
			if len(errMsg) > 1000 {
				errMsg = errMsg[:1000] // Truncate for DB
			}

			if err := bp.repository.MarkBlockFailed(block.ID, errMsg); err != nil {
				logger.Warnf("failed to mark block as failed: %v", err)
			}

			logger.With(
				"block_id", block.ID,
				"retry_count", block.RetryCount+1,
				"error", err,
			).Error("block processing failed")

			// Check if we should stop all workers
			if bp.config.StopOnFailure && block.RetryCount+1 >= bp.config.MaxRetries {
				bp.reportError(fmt.Errorf("block %d failed after %d retries: %w",
					block.ID, block.RetryCount+1, err))
				return
			}

			// Exponential backoff before next block
			if block.RetryCount > 0 {
				backoff := bp.config.RetryBackoff * time.Duration(1<<uint(block.RetryCount))
				logger.With("backoff", backoff).Info("waiting before next block")
				time.Sleep(backoff)
			}

			continue
		}

		// Mark as completed
		if err := bp.repository.MarkBlockCompleted(block.ID); err != nil {
			bp.reportError(fmt.Errorf("worker %d: mark block completed: %w", workerID, err))
			return
		}

		logger.With("block_id", block.ID).Info("block completed successfully")
	}
}

// monitorErrors handles error reporting and cancellation
func (bp *BlockProcessor) monitorErrors() {
	for err := range bp.errChan {
		bp.errorMu.Lock()
		if bp.globalError == nil {
			bp.globalError = err
			bp.cancel() // Stop all workers
		}
		bp.errorMu.Unlock()
	}
}

// reportError sends an error to the error channel
func (bp *BlockProcessor) reportError(err error) {
	select {
	case bp.errChan <- err:
	default:
		// Channel full, error already reported
	}
}
