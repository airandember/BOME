# 🎨 **FRONTEND DOCUMENTATION**

**Purpose:** Frontend patterns, Svelte 5 reactivity, and UI standards

---

## 📚 **DOCUMENTS IN THIS FOLDER**

### **⭐ REACTIVITY GUIDE**
- **`SVELTE5_REACTIVITY_GUIDE.md`** - Complete Svelte 5 patterns
  - 5 Core Rules (CRITICAL)
  - Common patterns
  - Debugging tips
  - Performance optimization
  - Examples and anti-patterns

---

## ⚡ **THE 5 CORE RULES**

### **Rule 1: Use `$state()` for ALL UI-bound variables**
```typescript
✅ let count = $state(0);
❌ let count = 0;  // Not reactive!
```

### **Rule 2: Use `$derived()` for computed values**
```typescript
✅ let doubled = $derived(count * 2);
❌ let doubled = count * 2;  // Stale!
```

### **Rule 3: Immutable updates for arrays/objects**
```typescript
✅ items = [...items, newItem];
❌ items.push(newItem);  // Doesn't trigger!
```

### **Rule 4: Force reactivity for Set/Map**
```typescript
✅ tags.add('new'); tags = tags;
❌ tags.add('new');  // Won't update UI!
```

### **Rule 5: `$effect()` for side effects, `onMount()` for init**
```typescript
✅ $effect(() => { console.log('Count:', count); });
✅ onMount(() => { loadData(); });
```

---

## 🎯 **WHEN TO USE**

### **Read `SVELTE5_REACTIVITY_GUIDE.md` when:**
- Building Svelte components
- Debugging reactivity issues
- UI not updating as expected
- Learning Svelte 5 runes
- Reviewing frontend code

### **Common Issues Solved:**
- "Why isn't my UI updating?"
- "Array changes not showing"
- "Set/Map mutations not reactive"
- "Derived values are stale"
- "When to use $effect vs onMount?"

---

## 📊 **FRONTEND ARCHITECTURE**

### **Directory Structure:**
```
frontend/src/
├── lib/
│   ├── api/                    # API client services
│   ├── services/               # Business logic services
│   ├── types/                  # TypeScript interfaces
│   ├── components/             # Reusable components
│   ├── stores/                 # Svelte stores
│   ├── websocket/              # WebSocket clients
│   └── utils/                  # Utility functions
└── routes/
    ├── admin/                  # Admin pages
    │   └── streaming/          # Streaming subsite
    │       ├── +layout.svelte  # Layout with auth
    │       ├── +page.svelte    # Dashboard
    │       ├── videos/
    │       ├── subscribers/
    │       ├── youtube/
    │       ├── creator-payouts/
    │       └── tags-categories/
    └── (public)/               # Public routes
```

### **Import Path Standards:**
```typescript
// ✅ CORRECT: Use $lib alias
import { service } from '$lib/services/service';
import type { Type } from '$lib/types/types';
import { apiClient } from '$lib/api/client';
import Component from '$lib/components/Component.svelte';

// ❌ WRONG: Relative paths
import { service } from './service';
import type { Type } from '../types/types';
```

---

## 🛠️ **FRONTEND TOOLS**

### **Development:**
```bash
npm run dev           # Start dev server
npm run build         # Production build
npm run preview       # Preview build
npm run type-check    # TypeScript check
npm run lint          # ESLint
```

### **Debugging:**
- Browser DevTools → Console
- Browser DevTools → Network (API calls)
- Browser DevTools → Vue/React DevTools (Svelte)
- `console.log()` in `$effect()` to trace reactivity

---

## 📝 **COMPONENT TEMPLATE**

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import type { MyType } from '$lib/types';
    
    // Props
    interface Props {
        id: number;
        onUpdate?: (data: MyType) => void;
    }
    let { id, onUpdate }: Props = $props();
    
    // State
    let data = $state<MyType | null>(null);
    let isLoading = $state(true);
    let error = $state<string | null>(null);
    
    // Derived
    let displayText = $derived(
        data ? data.name : 'Loading...'
    );
    
    // Functions
    async function loadData() {
        isLoading = true;
        try {
            const response = await fetch(`/api/data/${id}`);
            data = await response.json();
            error = null;
        } catch (e) {
            error = e.message;
        } finally {
            isLoading = false;
        }
    }
    
    // Effects
    $effect(() => {
        if (id) loadData();
    });
    
    // Lifecycle
    onMount(() => {
        console.log('Component mounted');
    });
</script>

{#if isLoading}
    <div>Loading...</div>
{:else if error}
    <div class="error">{error}</div>
{:else if data}
    <div>{displayText}</div>
{/if}

<style>
    .error {
        color: red;
    }
</style>
```

---

## 🔗 **RELATED DOCUMENTATION**

- **Architecture:** `../1-ARCHITECTURE/BOME_CONTEXT_STANDARD.md` (Frontend Standards section)
- **Testing:** `../3-TESTING/BRAID_COMBING_STANDARD.md` (Layers 1-3)
- **API Integration:** See services in `frontend/src/lib/services/`

---

## 🎓 **LEARNING RESOURCES**

### **Official Docs:**
- [Svelte 5 Documentation](https://svelte.dev/docs)
- [SvelteKit Documentation](https://kit.svelte.dev/docs)

### **BOME-Specific:**
- `SVELTE5_REACTIVITY_GUIDE.md` - Our patterns
- `BOME_CONTEXT_STANDARD.md` - Frontend section
- Example components in `frontend/src/routes/`

---

## 📊 **FRONTEND STATS**

- **Framework:** SvelteKit 2.0+
- **Language:** TypeScript 5.0+
- **Reactivity:** Svelte 5 Runes
- **Styling:** Tailwind CSS 3.0+
- **Components:** 50+
- **Pages:** 20+
- **Services:** 15+
- **Type Definitions:** 20+ interfaces

---

**Location:** `CONTEXT/4-FRONTEND/`  
**Files:** 1 comprehensive guide (more to come)  
**Status:** Complete ✅  
**Coverage:** Svelte 5 reactivity patterns

