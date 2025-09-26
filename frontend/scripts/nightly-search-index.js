#!/usr/bin/env node

/**
 * Nightly Search Index Generator
 * Runs as a scheduled job to regenerate the search index
 */

import cron from 'node-cron';
import fetch from 'node-fetch';
import fs from 'fs/promises';
import path from 'path';

// Configuration
const CONFIG = {
    // Cron schedule: "0 0 * * *" = every day at midnight
    schedule: process.env.SEARCH_INDEX_SCHEDULE || '0 0 * * *',
    apiUrl: process.env.API_URL || process.env.VITE_API_URL || 'http://localhost:8080/api/v1',
    outputFile: process.env.SEARCH_INDEX_OUTPUT || 'static/search-index.json',
    timezone: process.env.TZ || 'America/Denver', // MST timezone
    enableBackup: process.env.ENABLE_BACKUP_SCHEDULE !== 'false',
    backupSchedule: '0 6 * * *', // 6 AM backup
};

class NightlySearchIndexGenerator {
    constructor() {
        this.isRunning = false;
        this.lastRun = null;
        this.nextRun = null;
        this.stats = {
            totalRuns: 0,
            successfulRuns: 0,
            failedRuns: 0,
            lastError: null
        };
    }

    start() {
        console.log('🌙 Starting nightly search index generator...');
        console.log(`📅 Schedule: ${CONFIG.schedule} (${CONFIG.timezone})`);
        console.log(`🎯 API URL: ${CONFIG.apiUrl}`);
        console.log(`📁 Output: ${CONFIG.outputFile}`);

        // Main midnight schedule
        cron.schedule(CONFIG.schedule, () => {
            this.generateSearchIndex('midnight');
        }, {
            scheduled: true,
            timezone: CONFIG.timezone
        });

        // Backup morning schedule
        if (CONFIG.enableBackup) {
            cron.schedule(CONFIG.backupSchedule, () => {
                this.generateSearchIndex('morning-backup');
            }, {
                scheduled: true,
                timezone: CONFIG.timezone
            });
            console.log(`🌅 Backup schedule: ${CONFIG.backupSchedule}`);
        }

        this.isRunning = true;
        console.log('✅ Nightly search index generator started');

        // Log next run time
        this.logNextRun();
    }

    async generateSearchIndex(trigger = 'manual') {
        if (this.isGenerating) {
            console.log('⚠️ Search index generation already in progress, skipping...');
            return;
        }

        this.isGenerating = true;
        this.stats.totalRuns++;
        
        const startTime = Date.now();
        console.log(`🚀 [${trigger.toUpperCase()}] Starting search index generation...`);

        try {
            // Fetch all videos from API
            const videos = await this.fetchAllVideos();
            
            // Generate search index
            const searchIndex = {
                version: '1.0',
                generatedAt: new Date().toISOString(),
                totalVideos: videos.length,
                videos: videos,
                metadata: {
                    trigger: trigger,
                    generationTime: null, // Will be set below
                    timezone: CONFIG.timezone
                }
            };

            // Write to file
            await this.writeSearchIndex(searchIndex);

            const duration = Date.now() - startTime;
            searchIndex.metadata.generationTime = `${duration}ms`;

            this.lastRun = new Date();
            this.stats.successfulRuns++;
            this.stats.lastError = null;

            console.log(`✅ [${trigger.toUpperCase()}] Search index generated successfully!`);
            console.log(`📊 Stats: ${videos.length} videos, ${Math.round(duration/1000)}s generation time`);

            // Log next run
            this.logNextRun();

        } catch (error) {
            this.stats.failedRuns++;
            this.stats.lastError = error.message;
            
            console.error(`❌ [${trigger.toUpperCase()}] Search index generation failed:`, error);
            
            // Don't throw - let the scheduler continue
        } finally {
            this.isGenerating = false;
        }
    }

    async fetchAllVideos() {
        console.log('🔍 Fetching all videos from API...');
        
        let allVideos = [];
        let page = 1;
        let hasMore = true;
        
        while (hasMore) {
            const response = await fetch(`${CONFIG.apiUrl}/bunny-videos?page=${page}&limit=100`);
            
            if (!response.ok) {
                throw new Error(`API request failed: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            const videos = data.videos || [];
            
            if (videos.length === 0) {
                hasMore = false;
                break;
            }
            
            // Process videos for search index
            const searchableVideos = videos.map(video => ({
                id: video.id,
                title: video.title || '',
                description: video.description || '',
                category: video.category || '',
                tags: video.tags || [],
                duration: video.duration || 0,
                createdAt: video.createdAt || '',
                thumbnail: video.thumbnailUrl || video.thumbnail || '',
                thumbnailUrl: video.thumbnailUrl || '',
                bunny: {
                    guid: video.bunnyVideoId || video.id || '',
                    videoLibraryId: video.videoLibraryId || '',
                    thumbnailFileName: video.thumbnailFileName || '',
                    previewImageUrl: video.previewImageUrl || '',
                    width: video.width || 0,
                    height: video.height || 0,
                    length: video.duration || 0
                },
                views: video.viewCount || video.views || 0,
                status: video.status || 'unknown',
                videoUrl: video.videoUrl || video.playbackUrl || '',
                iframeSrc: video.iframeSrc || ''
            }));
            
            allVideos.push(...searchableVideos);
            console.log(`📥 Page ${page}: ${videos.length} videos (Total: ${allVideos.length})`);
            
            page++;
            
            // Safety check
            if (page > 100) {
                console.warn('⚠️ Reached page limit, stopping...');
                break;
            }
        }
        
        return allVideos;
    }

    async writeSearchIndex(searchIndex) {
        console.log(`📝 Writing search index to: ${CONFIG.outputFile}`);
        
        // Ensure directory exists
        const dir = path.dirname(CONFIG.outputFile);
        await fs.mkdir(dir, { recursive: true });
        
        // Write with pretty formatting
        const jsonData = JSON.stringify(searchIndex, null, 2);
        await fs.writeFile(CONFIG.outputFile, jsonData);
        
        // Log file size
        const stats = await fs.stat(CONFIG.outputFile);
        const fileSizeKB = Math.round(stats.size / 1024);
        console.log(`📊 File written: ${fileSizeKB} KB`);
    }

    logNextRun() {
        // This is a simplified next run calculation
        // In a real implementation, you'd use the cron library's next run calculation
        const now = new Date();
        const tomorrow = new Date(now);
        tomorrow.setDate(tomorrow.getDate() + 1);
        tomorrow.setHours(0, 0, 0, 0);
        
        console.log(`⏰ Next run scheduled for: ${tomorrow.toLocaleString('en-US', { timeZone: CONFIG.timezone })}`);
    }

    getStatus() {
        return {
            isRunning: this.isRunning,
            isGenerating: this.isGenerating || false,
            lastRun: this.lastRun,
            schedule: CONFIG.schedule,
            timezone: CONFIG.timezone,
            stats: this.stats,
            config: {
                apiUrl: CONFIG.apiUrl,
                outputFile: CONFIG.outputFile,
                enableBackup: CONFIG.enableBackup
            }
        };
    }

    // Manual trigger for testing
    async triggerManual() {
        console.log('🔄 Manual trigger requested...');
        await this.generateSearchIndex('manual');
    }
}

// Create and start the generator
const generator = new NightlySearchIndexGenerator();

// Handle different run modes
const args = process.argv.slice(2);
const command = args[0];

switch (command) {
    case 'start':
        generator.start();
        break;
    
    case 'generate':
        console.log('🔄 Running one-time generation...');
        generator.triggerManual().then(() => {
            console.log('✅ One-time generation complete');
            process.exit(0);
        }).catch(error => {
            console.error('❌ Generation failed:', error);
            process.exit(1);
        });
        break;
    
    case 'status':
        console.log('📊 Generator Status:', JSON.stringify(generator.getStatus(), null, 2));
        break;
    
    default:
        console.log('🌙 Nightly Search Index Generator');
        console.log('');
        console.log('Usage:');
        console.log('  node nightly-search-index.js start     - Start the scheduled generator');
        console.log('  node nightly-search-index.js generate  - Run one-time generation');
        console.log('  node nightly-search-index.js status    - Show current status');
        console.log('');
        console.log('Environment Variables:');
        console.log('  SEARCH_INDEX_SCHEDULE - Cron schedule (default: "0 0 * * *")');
        console.log('  API_URL              - API base URL');
        console.log('  SEARCH_INDEX_OUTPUT  - Output file path');
        console.log('  TZ                   - Timezone (default: America/Denver)');
        process.exit(0);
}

// Graceful shutdown
process.on('SIGTERM', () => {
    console.log('📴 Received SIGTERM, shutting down gracefully...');
    process.exit(0);
});

process.on('SIGINT', () => {
    console.log('📴 Received SIGINT, shutting down gracefully...');
    process.exit(0);
});
