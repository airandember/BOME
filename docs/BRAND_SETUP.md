# 🎨 Brand Setup Guide

This guide explains how to create a new branded deployment of the platform.

---

## 🎯 **Overview**

The platform supports **multiple branded deployments** from a single codebase. Each brand has:
- Own domain and branding
- Own database
- Own third-party integrations (Stripe, Bunny.net)
- Own feature set
- Completely separate from other brands

---

## 🚀 **Quick Start (15 minutes)**

### **Step 1: Create Brand Environment File**

```bash
# Copy template
cp .env.template .env.yourbrand

# Edit with your brand details
code .env.yourbrand
```

### **Step 2: Configure Brand Identity**

Edit `.env.yourbrand`:

```env
PUBLIC_BRAND_NAME="Your Brand"
PUBLIC_BRAND_TAGLINE="Your Tagline Here"
PUBLIC_BRAND_DOMAIN="yourbrand.com"

# Colors (use hex format)
PUBLIC_COLOR_PRIMARY="#yourcolor"
PUBLIC_COLOR_SECONDARY="#yourcolor"
PUBLIC_COLOR_ACCENT="#yourcolor"

# Contact
PUBLIC_CONTACT_SUPPORT="support@yourbrand.com"
PUBLIC_CONTACT_BUSINESS="business@yourbrand.com"

# Storage prefix (lowercase, no spaces)
PUBLIC_STORAGE_PREFIX="yourbrand"
```

### **Step 3: Add Brand Assets**

```bash
# Create brand assets directory
mkdir -p frontend/static/brands/yourbrand

# Add your assets
frontend/static/brands/yourbrand/
├── logo-light.svg    # Logo for light mode
├── logo-dark.svg     # Logo for dark mode
├── favicon.ico       # Browser favicon
└── og-image.png      # Social media preview image
```

**Asset Requirements**:
- **Logos**: SVG format, transparent background
  - Light: For dark backgrounds
  - Dark: For light backgrounds
- **Favicon**: 32x32 or 64x64 ICO format
- **OG Image**: 1200x630 PNG for social sharing

### **Step 4: Create Separate Database**

```sql
CREATE DATABASE yourbrand_production;
CREATE USER yourbrand_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE yourbrand_production TO yourbrand_user;
```

Update in `.env.yourbrand`:
```env
DB_NAME=yourbrand_production
DB_USER=yourbrand_user
DB_PASSWORD=secure_password
```

### **Step 5: Setup Integrations**

#### **Stripe**:
1. Create Stripe account for brand
2. Get API keys
3. Add to `.env.yourbrand`:
```env
STRIPE_PUBLIC_KEY="pk_live_..."
STRIPE_SECRET_KEY="sk_live_..."
```

#### **Bunny.net**:
1. Create Stream Library
2. Get Library ID and API key
3. Add to `.env.yourbrand`:
```env
BUNNY_API_KEY="..."
BUNNY_LIBRARY_ID="12345"
```

### **Step 6: Deploy**

```bash
# Use brand environment
cp .env.yourbrand .env

# Build
npm run build        # Frontend
go build            # Backend

# Run migrations
./backend migrate

# Deploy to your domain
# (deployment process varies by hosting provider)
```

---

## 📋 **Configuration Reference**

### **Brand Identity**
| Variable | Description | Example |
|----------|-------------|---------|
| `PUBLIC_BRAND_NAME` | Short brand name | "BOME" |
| `PUBLIC_BRAND_TAGLINE` | SEO-friendly tagline | "Your Streaming Platform" |
| `PUBLIC_BRAND_LEGAL_NAME` | Legal entity name | "Book of Mormon Evidence" |
| `PUBLIC_BRAND_DOMAIN` | Primary domain | "yourbrand.com" |

### **Visual Branding**
| Variable | Description | Format |
|----------|-------------|--------|
| `PUBLIC_COLOR_PRIMARY` | Main brand color | Hex: "#6366f1" |
| `PUBLIC_COLOR_SECONDARY` | Secondary color | Hex: "#8b5cf6" |
| `PUBLIC_COLOR_ACCENT` | Accent color | Hex: "#ec4899" |
| `PUBLIC_GRADIENT_PRIMARY` | Primary gradient | CSS gradient string |

### **Feature Flags**
| Variable | Description | Default |
|----------|-------------|---------|
| `PUBLIC_ENABLE_ADVERTISEMENTS` | Show ads | true |
| `PUBLIC_ENABLE_SUBSCRIPTIONS` | Subscription paywall | true |
| `PUBLIC_ENABLE_COMMENTS` | User comments | true |
| `PUBLIC_ENABLE_ARTICLES` | Article system | true |
| `PUBLIC_ENABLE_EVENTS` | Events/exhibitions | true |
| `PUBLIC_ENABLE_LIVE_STREAMING` | Live streams | false |

---

## 🎨 **Brand Assets Guide**

### **Logo Design Guidelines**:
- **Format**: SVG (vector) for scalability
- **Sizes**: Responsive (works at any size)
- **Variants**:
  - Light version: For dark backgrounds
  - Dark version: For light backgrounds
- **Background**: Transparent
- **Safe area**: Leave padding around edges

### **Color Selection**:
- **Primary**: Main brand color (buttons, links, headers)
- **Secondary**: Supporting color (accents, highlights)
- **Accent**: Call-to-action elements
- **Contrast**: Ensure WCAG AA compliance (4.5:1 ratio)

### **Testing Colors**:
```html
<!-- Test in browser console -->
<div style="background: #yourcolor; padding: 20px; color: white;">
  Test Text Contrast
</div>
```

---

## 🔧 **Development & Testing**

### **Local Development**:
```bash
# Use brand environment
cp .env.yourbrand .env

# Start development servers
npm run dev          # Frontend
go run main.go       # Backend
```

### **Testing Brand Configuration**:
```javascript
// Open browser console
console.log(window.__BRAND_CONFIG__);

// Should show your brand config
```

### **Verify Assets**:
1. Check logo displays correctly
2. Verify colors apply site-wide
3. Test feature flags work
4. Confirm contact emails correct

---

## 📦 **Multiple Brand Deployment**

### **Running Multiple Brands Simultaneously**:

**Docker Compose**:
```yaml
# docker-compose.yml
services:
  brand1:
    build: .
    env_file: .env.brand1
    ports:
      - "3000:3000"
  
  brand2:
    build: .
    env_file: .env.brand2
    ports:
      - "3001:3000"
```

**Nginx Routing** (by domain):
```nginx
# Route brand1.com to brand1 instance
server {
    server_name brand1.com;
    location / {
        proxy_pass http://localhost:3000;
    }
}

# Route brand2.com to brand2 instance
server {
    server_name brand2.com;
    location / {
        proxy_pass http://localhost:3001;
    }
}
```

---

## ✅ **Brand Setup Checklist**

### **Pre-Launch**:
- [ ] Brand environment file created
- [ ] All brand identity fields filled
- [ ] Brand assets uploaded
- [ ] Colors tested for accessibility
- [ ] Feature flags configured
- [ ] Contact emails set up
- [ ] Database created
- [ ] Stripe account configured
- [ ] Bunny.net library created
- [ ] Domain DNS configured

### **Testing**:
- [ ] Homepage loads with correct branding
- [ ] Logo displays on all pages
- [ ] Colors apply consistently
- [ ] Feature flags work as expected
- [ ] Contact forms use correct emails
- [ ] localStorage keys are unique
- [ ] No console errors
- [ ] Mobile responsive

### **Go Live**:
- [ ] Production environment configured
- [ ] SSL certificate installed
- [ ] Database backups configured
- [ ] Monitoring set up
- [ ] Analytics installed
- [ ] Email sending verified

---

## 🆘 **Troubleshooting**

### **Brand assets not showing**:
- Check file paths in `.env`
- Verify files exist in `/static/brands/yourbrand/`
- Clear browser cache
- Check console for 404 errors

### **Colors not applying**:
- Verify hex format: `#RRGGBB`
- Check PUBLIC_ prefix on variables
- Rebuild frontend: `npm run build`
- Clear browser cache

### **Feature flags not working**:
- Verify values are `true` or `false` (lowercase)
- Rebuild application
- Check console: `window.__BRAND_CONFIG__.features`

### **LocalStorage conflicts**:
- Ensure unique `PUBLIC_STORAGE_PREFIX`
- Clear browser storage
- Use incognito mode for testing

---

## 📞 **Support**

Having issues? Check:
1. This documentation
2. `.env.template` for all available options
3. `frontend/src/lib/config/brand.ts` for config structure

---

**Last Updated**: October 14, 2025  
**Version**: 1.0.0

