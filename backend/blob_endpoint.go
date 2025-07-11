// Add blob URL endpoint for direct video data access
v1.GET("/blob/:videoId", middleware.AuthRequired(), func(c *gin.Context) {
    videoID := c.Param("videoId")
    
    // Get user info from context
    userID := c.GetInt("user_id")
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
        return
    }

    fmt.Printf("[Blob] Request for video: %s by user: %d\n", videoID, userID)

    // Get the direct video URL from Bunny
    directURL := fmt.Sprintf("https://vz-%s-%s.b-cdn.net/%s/play_720p.mp4",
        bunnyService.GetStreamLibrary(),
        bunnyService.GetRegion(),
        videoID)

    fmt.Printf("[Blob] Fetching from: %s\n", directURL)

    // Create the request
    req, err := http.NewRequest("GET", directURL, nil)
    if err != nil {
        fmt.Printf("[Blob] Failed to create request: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
        return
    }

    // Add headers
    req.Header.Set("Accept", "video/mp4,*/*")
    req.Header.Set("User-Agent", "BOME-Backend/1.0")
    
    // Try without authentication first
    client := &http.Client{Timeout: 60 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("[Blob] Request failed: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch video"})
        return
    }
    defer resp.Body.Close()

    fmt.Printf("[Blob] Response status: %d\n", resp.StatusCode)

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        fmt.Printf("[Blob] Error response: %s\n", string(body))
        c.JSON(resp.StatusCode, gin.H{"error": "Video not accessible"})
        return
    }

    // Set response headers for blob creation
    c.Header("Content-Type", "video/mp4")
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
    c.Header("Cache-Control", "public, max-age=3600")
    
    // Copy content length if available
    if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
        c.Header("Content-Length", contentLength)
    }

    // Stream the video data
    c.Status(http.StatusOK)
    written, err := io.Copy(c.Writer, resp.Body)
    if err != nil {
        fmt.Printf("[Blob] Error streaming: %v\n", err)
    } else {
        fmt.Printf("[Blob] Successfully streamed %d bytes\n", written)
    }
})
