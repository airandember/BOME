package services

import (
	"fmt"
	"log"
	"time"
)

// StripeLogger provides structured logging for Stripe operations
type StripeLogger struct {
	prefix string
}

// NewStripeLogger creates a new logger with a prefix
func NewStripeLogger(prefix string) *StripeLogger {
	return &StripeLogger{prefix: prefix}
}

// LogSyncStart logs the start of a sync operation
func (l *StripeLogger) LogSyncStart(syncType, entityType string, totalItems int) {
	timestamp := time.Now().Format("15:04:05")
	log.Printf("🚀 [%s] %s SYNC START: %s | Items: %d | Time: %s",
		l.prefix, syncType, entityType, totalItems, timestamp)
}

// LogSyncProgress logs sync progress
func (l *StripeLogger) LogSyncProgress(entityType string, processed, total int, duration time.Duration) {
	percentage := float64(processed) / float64(total) * 100
	log.Printf("📊 [%s] PROGRESS: %s | %d/%d (%.1f%%) | Duration: %v",
		l.prefix, entityType, processed, total, percentage, duration)
}

// LogSyncComplete logs sync completion
func (l *StripeLogger) LogSyncComplete(entityType string, processed int, duration time.Duration, errors int) {
	status := "✅ SUCCESS"
	if errors > 0 {
		status = fmt.Sprintf("⚠️ COMPLETED WITH %d ERRORS", errors)
	}

	log.Printf("%s [%s] %s COMPLETE | Processed: %d | Duration: %v | Errors: %d",
		status, l.prefix, entityType, processed, duration, errors)
}

// LogAPICall logs Stripe API calls
func (l *StripeLogger) LogAPICall(endpoint string, params map[string]interface{}, duration time.Duration) {
	log.Printf("📡 [%s] API CALL: %s | Params: %v | Duration: %v",
		l.prefix, endpoint, params, duration)
}

// LogDatabaseOperation logs database operations
func (l *StripeLogger) LogDatabaseOperation(operation, table string, affected int, duration time.Duration) {
	log.Printf("💾 [%s] DB %s: %s | Affected: %d | Duration: %v",
		l.prefix, operation, table, affected, duration)
}

// LogWebhookReceived logs webhook events
func (l *StripeLogger) LogWebhookReceived(eventType, objectID string) {
	timestamp := time.Now().Format("15:04:05")
	log.Printf("📨 [%s] WEBHOOK: %s | Object: %s | Time: %s",
		l.prefix, eventType, objectID, timestamp)
}

// LogError logs errors with context
func (l *StripeLogger) LogError(operation string, err error, context map[string]interface{}) {
	log.Printf("❌ [%s] ERROR in %s: %v | Context: %v",
		l.prefix, operation, err, context)
}

// LogTableStats logs database table statistics
func (l *StripeLogger) LogTableStats(table string, count int) {
	log.Printf("📋 [%s] TABLE STATS: %s has %d records",
		l.prefix, table, count)
}

// LogCronJob logs cron job execution
func (l *StripeLogger) LogCronJob(jobType string, nextRun time.Time) {
	log.Printf("⏰ [%s] CRON: %s scheduled for %s",
		l.prefix, jobType, nextRun.Format("2006-01-02 15:04:05 MST"))
}

// LogSystemStatus logs overall system status
func (l *StripeLogger) LogSystemStatus(enabled bool, keyType string, tablesReady bool) {
	status := "❌ DISABLED"
	if enabled {
		status = "✅ ENABLED"
	}

	tables := "❌ NOT READY"
	if tablesReady {
		tables = "✅ READY"
	}

	log.Printf("🔧 [%s] SYSTEM STATUS: %s | Key: %s | Tables: %s",
		l.prefix, status, keyType, tables)
}
