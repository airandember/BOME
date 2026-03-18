# Backend Braids - Implementation Documentation

## Relationship to _braids/

The **`backend/braids/`** directory contains **implementation-specific** braid documentation for flows that live within the main braids but warrant separate technical deep-dives.

| Backend Braid | Parent _braids/ Braid | Purpose |
|---------------|----------------------|---------|
| **subscription-checkout** | `_braids/subscription-billing/` | Dual-confirmation checkout flow, Stripe session handling |
| **video-analytics** | `_braids/video-streaming/` | Video view tracking, watch history, Bunny.net sync |

## When to Use This vs _braids/

- **`_braids/{braid}/`**: Full braid documentation (scope, layers, strands, file maps)
- **`backend/braids/{sub-braid}/`**: Implementation guides for specific workflows within a braid

For overall architecture and production file maps, see `_braids/`. For checkout flow details or video analytics implementation, use these backend braids.
