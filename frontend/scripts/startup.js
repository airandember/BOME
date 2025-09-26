#!/usr/bin/env node

/**
 * Production Startup Script
 * Generates search index and starts the server
 */

import { spawn } from 'child_process';
import fs from 'fs/promises';
import path from 'path';

const SEARCH_INDEX_PATH = 'build/client/search-index.json';

async function checkSearchIndex() {
    try {
        const stats = await fs.stat(SEARCH_INDEX_PATH);
        const ageHours = (Date.now() - stats.mtime.getTime()) / (1000 * 60 * 60);
        
        console.log(`📊 Search index found, age: ${ageHours.toFixed(1)} hours`);
        
        // Regenerate if older than 24 hours
        if (ageHours > 24) {
            console.log('🔄 Search index is stale, regenerating...');
            return false;
        }
        
        return true;
    } catch (error) {
        console.log('❌ Search index not found, generating...');
        return false;
    }
}

async function generateSearchIndex() {
    return new Promise((resolve, reject) => {
        console.log('🚀 Generating search index...');
        
        const child = spawn('node', ['scripts/build-search-index.js'], {
            stdio: 'inherit',
            cwd: process.cwd()
        });
        
        child.on('close', (code) => {
            if (code === 0) {
                console.log('✅ Search index generated successfully');
                
                // Copy to build directory
                fs.copyFile('static/search-index.json', SEARCH_INDEX_PATH)
                    .then(() => {
                        console.log('📋 Search index copied to build directory');
                        resolve();
                    })
                    .catch(reject);
            } else {
                reject(new Error(`Search index generation failed with code ${code}`));
            }
        });
        
        child.on('error', reject);
    });
}

async function startServer() {
    console.log('🚀 Starting production server...');
    
    const server = spawn('node', ['build/index.js'], {
        stdio: 'inherit',
        env: { ...process.env, NODE_ENV: 'production' }
    });
    
    server.on('error', (error) => {
        console.error('❌ Server failed to start:', error);
        process.exit(1);
    });
    
    // Handle graceful shutdown
    process.on('SIGTERM', () => {
        console.log('📴 Received SIGTERM, shutting down gracefully...');
        server.kill('SIGTERM');
    });
    
    process.on('SIGINT', () => {
        console.log('📴 Received SIGINT, shutting down gracefully...');
        server.kill('SIGINT');
    });
}

async function main() {
    try {
        const indexExists = await checkSearchIndex();
        
        if (!indexExists) {
            await generateSearchIndex();
        }
        
        await startServer();
    } catch (error) {
        console.error('❌ Startup failed:', error);
        process.exit(1);
    }
}

main();
