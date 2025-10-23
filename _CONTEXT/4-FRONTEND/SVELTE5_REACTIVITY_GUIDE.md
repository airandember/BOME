# ⚡ Svelte 5 Reactivity Guide

**Version:** 1.0  
**Last Updated:** October 22, 2025  
**Critical for:** Frontend development  

---

## 🚨 THE 5 CORE RULES

### Rule 1: Use `$state()` for ALL UI-Bound Variables
```typescript
// ✅ CORRECT
let count = $state(0);
let user = $state<User | null>(null);
let items = $state<Item[]>([]);

// ❌ WRONG
let count = 0;  // Won't trigger reactivity!
let user: User | null = null;
```

### Rule 2: Use `$derived()` for Computed Values
```typescript
// ✅ CORRECT
let count = $state(0);
let doubled = $derived(count * 2);

// ❌ WRONG
let doubled = count * 2;  // Won't update!
```

### Rule 3: Immutable Updates for Arrays/Objects
```typescript
// ✅ CORRECT - Arrays
items = [...items, newItem];  // Creates new array
items = items.filter(i => i.id !== deleteId);

// ✅ CORRECT - Objects
user = { ...user, name: 'New Name' };

// ❌ WRONG
items.push(newItem);  // Mutates, won't trigger!
user.name = 'New Name';  // Mutates!
```

### Rule 4: Force Reactivity for Set/Map Mutations
```typescript
// ✅ CORRECT
let mySet = $state(new Set<number>());
mySet.add(1);
mySet = mySet;  // Force reactivity!

// ❌ WRONG
mySet.add(1);  // Won't trigger UI update!
```

### Rule 5: Use `$effect()` for Side Effects
```typescript
// ✅ CORRECT
$effect(() => {
    console.log('Count changed:', count);
});

// ❌ WRONG (Svelte 4 style)
$: console.log('Count:', count);  // Old syntax!
```

---

## COMMON PATTERNS

### Loading States
```typescript
let loading = $state(false);
let data = $state<Data | null>(null);
let error = $state<string | null>(null);

async function fetchData() {
    loading = true;
    error = null;
    
    try {
        const response = await api.getData();
        data = response.data;
    } catch (err) {
        error = err.message;
    } finally {
        loading = false;
    }
}
```

### Array Operations
```typescript
let items = $state<Item[]>([]);

// Add
items = [...items, newItem];

// Remove
items = items.filter(item => item.id !== deleteId);

// Update
items = items.map(item => 
    item.id === updateId ? { ...item, ...updates } : item
);

// Replace all
items = newItems;
```

### Object Updates
```typescript
let user = $state<User>({ id: 1, name: 'John' });

// Update field
user = { ...user, name: 'Jane' };

// Update nested
user = { ...user, profile: { ...user.profile, age: 30 } };

// Replace
user = newUser;
```

### Conditional Rendering
```typescript
{#if loading}
    <LoadingSpinner />
{:else if error}
    <ErrorMessage {error} />
{:else if data}
    <DataDisplay {data} />
{/if}
```

---

## LIFECYCLE

### Component Mount
```typescript
import { onMount } from 'svelte';

onMount(() => {
    fetchData();
    
    return () => {
        // Cleanup
    };
});
```

### Effect (Watch for Changes)
```typescript
$effect(() => {
    // Runs when dependencies change
    console.log('User changed:', user);
});
```

---

## STORES (Global State)

### Writable Store
```typescript
import { writable } from 'svelte/store';

export const userStore = writable<User | null>(null);

// Usage in component
import { userStore } from '$lib/stores/user';

let $userStore = $state(null);
userStore.subscribe(value => {
    $userStore = value;
});
```

---

## COMMON MISTAKES

### ❌ Mistake 1: Forgetting `$state()`
```typescript
// BAD
let count = 0;

// GOOD
let count = $state(0);
```

### ❌ Mistake 2: Mutating Arrays
```typescript
// BAD
items.push(newItem);

// GOOD
items = [...items, newItem];
```

### ❌ Mistake 3: Mutating Objects
```typescript
// BAD
user.name = 'New';

// GOOD
user = { ...user, name: 'New' };
```

### ❌ Mistake 4: Not Forcing Set/Map Reactivity
```typescript
// BAD
mySet.add(value);

// GOOD
mySet.add(value);
mySet = mySet;
```

### ❌ Mistake 5: Using Old Svelte 4 Syntax
```typescript
// BAD (Svelte 4)
$: doubled = count * 2;

// GOOD (Svelte 5)
let doubled = $derived(count * 2);
```

---

## DEBUGGING TIPS

### Check Reactivity
```typescript
$effect(() => {
    console.log('Value changed:', value);
});
```

### Verify State Updates
```typescript
function updateValue() {
    console.log('Before:', value);
    value = newValue;
    console.log('After:', value);
}
```

---

## PERFORMANCE

### Avoid Unnecessary Reactivity
```typescript
// ✅ GOOD - Only reactive when needed
let result = $derived(expensiveCalculation(input));

// ❌ BAD - Recalculates every render
let result = expensiveCalculation(input);
```

### Batch Updates
```typescript
// Updates happen together
loading = true;
data = null;
error = null;
```

---

## QUICK REFERENCE

| Pattern | Svelte 5 |
|---------|----------|
| State | `$state()` |
| Computed | `$derived()` |
| Effect | `$effect()` |
| Mount | `onMount()` |
| Array Add | `arr = [...arr, item]` |
| Object Update | `obj = { ...obj, key: val }` |
| Set Mutation | `set.add(x); set = set` |

---

## EXAMPLES

### Complete Component
```typescript
<script lang="ts">
import { onMount } from 'svelte';
import type { User } from '$lib/types/user';

let users = $state<User[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);
let search = $state('');

let filteredUsers = $derived(
    users.filter(u => 
        u.name.toLowerCase().includes(search.toLowerCase())
    )
);

onMount(async () => {
    await loadUsers();
});

async function loadUsers() {
    loading = true;
    error = null;
    
    try {
        const response = await fetch('/api/users');
        const data = await response.json();
        users = data.users;
    } catch (err) {
        error = err.message;
    } finally {
        loading = false;
    }
}

function addUser(user: User) {
    users = [...users, user];
}

function removeUser(id: number) {
    users = users.filter(u => u.id !== id);
}
</script>

{#if loading}
    <p>Loading...</p>
{:else if error}
    <p>Error: {error}</p>
{:else}
    <input bind:value={search} placeholder="Search..." />
    
    {#each filteredUsers as user (user.id)}
        <div>{user.name}</div>
    {/each}
{/if}
```

---

## MIGRATION FROM SVELTE 4

| Svelte 4 | Svelte 5 |
|----------|----------|
| `let x = 0` | `let x = $state(0)` |
| `$: doubled = x * 2` | `let doubled = $derived(x * 2)` |
| `$: { ... }` | `$effect(() => { ... })` |
| `$: console.log(x)` | `$effect(() => console.log(x))` |

---

## CONCLUSION

Svelte 5 reactivity is powerful but requires following the rules. Always use `$state()` for UI-bound variables, immutable updates for arrays/objects, and force reactivity for Set/Map.

**Remember the 5 Core Rules!**

---

*Last Updated: October 22, 2025*  
*Version: 1.0*  
*Status: Production Standard*
