#!/usr/bin/env node

/**
 * Build Search Index Script
 * Generates a comprehensive search index JSON file for instant video search
 */

import fetch from 'node-fetch';
import fs from 'fs/promises';
import path from 'path';

const API_BASE_URL = process.env.API_URL || 'http://localhost:8080/api/v1';
const OUTPUT_FILE = 'static/search-index.json';

async function fetchAllVideos() {
    console.log('🔍 Fetching all videos for search index...');
    
    let allVideos = [];
    let page = 1;
    let hasMore = true;
    
    while (hasMore) {
        try {
            console.log(`📥 Fetching page ${page}...`);
            
            const response = await fetch(`${API_BASE_URL}/bunny-videos?page=${page}&limit=100`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const videos = data.videos || [];
            
            if (videos.length === 0) {
                hasMore = false;
                break;
            }
            
            // Extract only the fields we need for search and display (optimized for instant loading)
            const searchableVideos = videos.map(video => {
                // Generate fallback thumbnail if none provided
                const primaryThumbnail = video.thumbnailUrl || video.thumbnail || '';
                const fallbackThumbnail = video.id ? 
                    `https://vz-f75053f7-465.b-cdn.net/${video.id}/thumbnail.jpg` : '';
                
                return {
                    id: video.id,
                    title: video.title || '',
                    description: video.description || '',
                    category: video.category || '',
                    tags: video.tags || [],
                    duration: video.duration || 0,
                    createdAt: video.createdAt || '',
                    // Thumbnail data for instant image loading (use API field names with fallbacks)
                    thumbnail: primaryThumbnail || fallbackThumbnail,
                    thumbnailUrl: primaryThumbnail || fallbackThumbnail,
                    // Bunny CDN data for video playback and thumbnails
                    bunny: {
                        guid: video.bunnyVideoId || video.id || '',
                        videoLibraryId: video.videoLibraryId || '',
                        thumbnailFileName: video.thumbnailFileName || '',
                        // Include preview thumbnails if available
                        previewImageUrl: video.previewImageUrl || fallbackThumbnail,
                        // Video metadata for display
                        width: video.width || 0,
                        height: video.height || 0,
                        length: video.duration || 0
                    },
                    // Additional display fields
                    views: video.viewCount || video.views || 0,
                    status: video.status || 'unknown',
                    // Video URLs for playback
                    videoUrl: video.videoUrl || video.playbackUrl || '',
                    iframeSrc: video.iframeSrc || ''
                };
            });
            
            allVideos.push(...searchableVideos);
            console.log(`✅ Page ${page}: ${videos.length} videos (Total: ${allVideos.length})`);
            
            page++;
            
            // Safety check to prevent infinite loops
            if (page > 100) {
                console.warn('⚠️ Reached page limit (100), stopping...');
                break;
            }
            
        } catch (error) {
            console.error(`❌ Error fetching page ${page}:`, error.message);
            hasMore = false;
        }
    }
    
    return allVideos;
}

async function buildSearchIndex() {
    try {
        console.log('🚀 Building comprehensive search index...');
        
        const videos = await fetchAllVideos();
        
        if (videos.length === 0) {
            console.warn('⚠️ No videos found, creating empty index');
        }
        
        const searchIndex = {
            version: '1.0',
            generatedAt: new Date().toISOString(),
            totalVideos: videos.length,
            videos: videos
        };
        
        // Ensure the static directory exists
        const outputDir = path.dirname(OUTPUT_FILE);
        await fs.mkdir(outputDir, { recursive: true });
        
        // Write the search index
        await fs.writeFile(OUTPUT_FILE, JSON.stringify(searchIndex, null, 2));
        
        const stats = await fs.stat(OUTPUT_FILE);
        const fileSizeKB = Math.round(stats.size / 1024);
        
        console.log('✅ Search index built successfully!');
        console.log(`📊 Stats:`);
        console.log(`   - Total videos: ${videos.length}`);
        console.log(`   - File size: ${fileSizeKB} KB`);
        console.log(`   - Output: ${OUTPUT_FILE}`);
        
    } catch (error) {
        console.error('❌ Failed to build search index:', error);
        process.exit(1);
    }
}

// Run the build
buildSearchIndex();
