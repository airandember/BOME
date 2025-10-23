# Admin Sidebar Style Guide

## 📐 User-Approved Styling Standards

This document captures the user's custom styling preferences for the admin sidebar that should be maintained across all future updates.

---

## 🎨 Icon Sizing

### Standard View (Expanded)
```css
.nav-item svg {
    width: 25px;
    height: 25px;
    flex-shrink: 0;
    margin: 0.5rem;
}
```

### Collapsed View
```css
.admin-sidebar.collapsed .nav-item svg {
    color: var(--text-secondary);
    stroke: var(--text-muted);
    width: 40px !important;
    height: 40px !important;
    padding: 0.15rem !important;
}
```

**Key Principle**: Icons should be **significantly larger** in collapsed view (40px) for better visual prominence and touch targets, while standard view uses 25px for balance with text.

---

## 🎯 Active State Styling

### Active Navigation Item
```css
.nav-item.active {
    /* background: var(--primary); */ /* REMOVED - no solid background */
    color: var(--primary-gold-dark); /* Gold accent instead of white */
}
```

**Key Principle**: Active items use **color accent only** (gold), NOT a solid background. This creates a cleaner, more refined look.

---

## 🏠 Subsite Icon Styling

### Subsite Icons (Streaming, Articles, Expo, etc.)
```css
.subsite-icon {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    
    /* 🎨 NEUMORPHIC GLASSMORPHIC STYLING */
    background: rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    box-shadow: 
        0 8px 32px 0 rgba(31, 38, 135, 0.15),
        inset 0 1px 3px rgba(255, 255, 255, 0.05);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.subsite-icon:hover {
    background: rgba(255, 255, 255, 0.08);
    transform: translateY(-2px);
    box-shadow: 
        0 12px 40px 0 rgba(31, 38, 135, 0.25),
        inset 0 1px 3px rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.15);
}
```

**Key Principle**: Subsite icons ALWAYS have **neumorphic glassmorphic styling** with:
- Semi-transparent backgrounds
- Blur effects
- Subtle inset shadows
- Smooth lift animation on hover
- 40px x 40px size for prominence

---

## 🔄 Collapse Behavior

### Detection Logic
```javascript
$effect(() => {
    const isSubsite = $page.url.pathname.includes('/admin/streaming') ||
                      $page.url.pathname.includes('/admin/#') ||  // Under construction
                      $page.url.pathname.includes('/admin/analytics') ||
                      $page.url.pathname.includes('/admin/ads');
    
    sidebarCollapsed.set(isSubsite);
});
```

**Key Principle**: Sidebar **auto-collapses** when entering any subsite, expanding when returning to main dashboard.

### Placeholder Routes
- Under construction subsites use `/admin/#` as placeholder path
- This prevents broken links while maintaining visual structure

---

## 🎭 Collapsed View Behavior

### Text Hiding
```css
/* Hide all text elements when collapsed */
.admin-sidebar.collapsed .brand-text,
.admin-sidebar.collapsed .nav-item span,
.admin-sidebar.collapsed .role-badge,
.admin-sidebar.collapsed .subsite-info,
.admin-sidebar.collapsed .subsite-name,
.admin-sidebar.collapsed .subsite-status,
.admin-sidebar.collapsed .section-header {
    display: none;
}

/* Hide text nodes using font-size trick */
.admin-sidebar.collapsed .nav-item {
    font-size: 0; /* This hides the text nodes! */
}
```

**Key Principle**: When collapsed, **ONLY icons are visible**. All text is completely hidden for a clean, icon-only interface.

---

## 🌊 Animation Standards

### Transitions
```css
.admin-sidebar {
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.admin-main {
    transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
```

**Key Principle**: All sidebar state changes use **0.3s cubic-bezier** for smooth, professional animations. No jarring or instant changes.

---

## 🎪 Icon Rendering

### Subsite Icon Rendering
```svelte
<div class="subsite-icon">
    {@html getSubsiteIcon(subsite.icon)}
</div>
```

**Key Principle**: Always use `{@html}` directive to render SVG icons. **NEVER** use `style="innerHTML={...}"` or other incorrect patterns.

---

## 📏 Spacing and Layout

### Collapsed Width
```css
.admin-sidebar.collapsed {
    width: 80px;
    overflow: hidden;
}
```

### Expanded Width
```css
.admin-sidebar {
    width: 320px;
}
```

**Key Principle**: 
- **Expanded**: 320px (comfortable for text + icons)
- **Collapsed**: 80px (perfect for centered 40px icons with padding)

---

## 🎨 Color Palette

### Primary Colors
- **Active Accent**: `var(--primary-gold-dark)` - For active state text
- **Icon Base**: `var(--text-secondary)` - For collapsed icon color
- **Icon Stroke**: `var(--text-muted)` - For collapsed icon strokes

### Glassmorphic Effects
- **Base Glass**: `rgba(255, 255, 255, 0.05)`
- **Hover Glass**: `rgba(255, 255, 255, 0.08)`
- **Border**: `rgba(255, 255, 255, 0.1)`
- **Border Hover**: `rgba(255, 255, 255, 0.15)`

---

## ✅ Implementation Checklist

When implementing or updating the admin sidebar:

- [ ] Icon sizes: 25px expanded, 40px collapsed
- [ ] Active state uses gold color, NO solid background
- [ ] All subsite icons have neumorphic glassmorphic styling
- [ ] Collapsed view shows ONLY icons (all text hidden)
- [ ] Auto-collapse on subsite entry
- [ ] Smooth 0.3s cubic-bezier transitions
- [ ] Use `{@html}` for SVG icon rendering
- [ ] 320px expanded, 80px collapsed
- [ ] Overflow hidden when collapsed

---

## 🚀 Example Implementation

```svelte
<nav class="admin-sidebar" class:collapsed={$sidebarCollapsed}>
    <!-- Navigation items with proper icon sizing and active states -->
    <a class="nav-item" class:active={isActive}>
        <svg><!-- Icon --></svg>
        <span>Label</span>
    </a>
    
    <!-- Subsite items with glassmorphic icons -->
    <a class="nav-item subsite-item">
        <div class="subsite-icon">
            {@html getSubsiteIcon('play')}
        </div>
        <div class="subsite-info">
            <div class="subsite-name">Streaming</div>
        </div>
    </a>
</nav>

<main class="admin-main" class:sidebar-collapsed={$sidebarCollapsed}>
    <slot />
</main>

<style>
    /* All styling rules from above sections */
</style>
```

---

**Last Updated**: October 22, 2025  
**Status**: ✅ User-Approved Production Standard  
**Applies To**: All admin dashboard interfaces

