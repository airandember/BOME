# Subscription Checkout BRAID

## 📚 Documentation Index

Welcome to the Subscription Checkout BRAID documentation. This BRAID handles the complete user subscription journey from plan selection through payment to video access provisioning.

### Quick Links

- **[BRAID.md](BRAID.md)** - Main architecture overview
- **[FLOW_DIAGRAM.md](FLOW_DIAGRAM.md)** - Visual flow from frontend to database
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Troubleshooting & common operations
- **[WORKFLOW_GUIDE.md](WORKFLOW_GUIDE.md)** - Role-based workflows & step-by-step guides

### Related Documentation (Root Level)

- **[EMAIL_PASSWORD_BUG_FIX.md](../../EMAIL_PASSWORD_BUG_FIX.md)** - Password setup missing auto-linking (FIXED 2025-11-18)
- **[RETROACTIVE_ACCESS_FIX.md](../../RETROACTIVE_ACCESS_FIX.md)** - Subscribe-before-register handling
- **[INCOMPLETE_SUBSCRIPTION_FIX.md](../../INCOMPLETE_SUBSCRIPTION_FIX.md)** - Incomplete status edge case
- **[DUAL_CONFIRMATION_SUBSCRIPTION_FLOW.md](../../DUAL_CONFIRMATION_SUBSCRIPTION_FLOW.md)** - Architecture proposal

---

## 🎯 Use This BRAID When...

- ❓ A user paid but doesn't have video access → [WORKFLOW_GUIDE.md](WORKFLOW_GUIDE.md#workflow-1-trace-a-specific-users-subscription)
- ❓ OAuth2 works but email/password doesn't → [QUICK_REFERENCE.md](QUICK_REFERENCE.md#oauth2-works-but-emailpassword-doesnt)
- ❓ You need to trace a subscription flow from frontend to backend → [FLOW_DIAGRAM.md](FLOW_DIAGRAM.md)
- ❓ Webhooks aren't arriving or are delayed → [WORKFLOW_GUIDE.md](WORKFLOW_GUIDE.md#workflow-3-debug-a-production-issue)
- ❓ You're implementing a new feature that touches subscriptions → [WORKFLOW_GUIDE.md](WORKFLOW_GUIDE.md#workflow-2-implement-a-new-feature)
- ❓ You're onboarding a new developer → [WORKFLOW_GUIDE.md](WORKFLOW_GUIDE.md#workflow-4-onboard-a-junior-developer)
- ❓ You need to understand the dual-confirmation pattern → [BRAID.md](BRAID.md#architecture-pattern-dual-confirmation)

---

## 🏗️ Architecture at a Glance

```
Frontend (Svelte)
    ↓
Backend Routes (Go)
    ↓
Services (Business Logic)
    ├─ Stripe Public Service (Session verification)
    ├─ Customer Linking Service (Auto-linking)
    ├─ Subscription Manager (Access grants)
    └─ OAuth2 Service (Alternative flow)
    ↓
Database (PostgreSQL)
    ├─ users (access status)
    ├─ stripe_customers_v2 (customer records)
    ├─ user_stripe_customers_v2 (linking table)
    └─ stripe_subscriptions_v2 (subscriptions)
```

---

## 🔄 The Dual-Confirmation Pattern

This system uses **two independent confirmation mechanisms** that work in parallel:

### Primary: Session Verification (Immediate)
- User redirected from Stripe → Frontend verifies session → Backend grants access
- **Speed:** < 500ms
- **User Experience:** Instant access

### Secondary: Webhook Confirmation (Backup)
- Stripe sends webhooks → Backend processes → Confirms/grants access
- **Speed:** 1-30 seconds
- **Reliability:** Ensures no missed grants

### Why Both?

- **User Experience:** Primary gives instant feedback
- **Reliability:** Secondary catches edge cases and delays
- **Idempotent:** Both can run safely without duplicate grants

---

## 🐛 Recent Bug Fixes

### 2025-11-18: Email/Password Registration Bug
**Problem:** OAuth2 subscriptions worked flawlessly, but email/password users didn't get access.

**Root Cause:** Password setup handler was missing auto-linking logic.

**Solution:** Added auto-linking to `SetupPasswordHandler` with retroactive access grant.

**Status:** ✅ Fixed - Both flows now have parity

**Details:** See [EMAIL_PASSWORD_BUG_FIX.md](../../EMAIL_PASSWORD_BUG_FIX.md)

---

## 📊 System Health

### Key Metrics to Monitor

| Metric | Target | Check With |
|--------|--------|------------|
| Instant access rate | > 95% | Logs: "SESSION-GRANT.*instant access" |
| Dual-confirmation rate | > 99% | DB: `video_access_source` contains both |
| Webhook arrival time | < 30s | Logs: Time between events |
| Failed access grants | 0 | Logs: "Failed to grant access" |
| Unlinked subscribers | 0 | See QUICK_REFERENCE.md SQL |

### Health Check Queries

```sql
-- Overall access source distribution
SELECT video_access_source, COUNT(*)
FROM users
WHERE has_video_access = true
GROUP BY video_access_source;

-- Recent problematic cases
SELECT u.id, u.email, u.has_video_access, ss.status
FROM users u
JOIN user_stripe_customers_v2 usc ON u.id = usc.user_id
JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
WHERE ss.status IN ('active', 'trialing')
AND u.has_video_access = false;
```

---

## 🚀 Getting Started

### For Developers

1. **Read BRAID.md** - Understand the overall architecture
2. **Review FLOW_DIAGRAM.md** - See the complete user journey
3. **Bookmark QUICK_REFERENCE.md** - For troubleshooting
4. **Run health checks** - Ensure system is working

### For Debugging

1. **Start with QUICK_REFERENCE.md** - Common issues & solutions
2. **Check logs** - Search for user email or customer ID
3. **Query database** - Use provided SQL queries
4. **Trace flow** - Follow FLOW_DIAGRAM.md

### For New Features

1. **Understand dual-confirmation** - Maintain pattern for new flows
2. **Add idempotency** - All access grants must be idempotent
3. **Multiple safety nets** - Link at every possible point
4. **Update this BRAID** - Document new flows and edge cases

---

## 📁 Directory Structure

```
subscription-checkout/
├── README.md                    ← You are here
├── BRAID.md                     ← Architecture overview
├── FLOW_DIAGRAM.md              ← Visual flow diagram
├── QUICK_REFERENCE.md           ← Troubleshooting guide
└── strands/                     ← Individual flow documentation
    ├── checkout-flow/
    ├── session-verification/
    ├── customer-linking/
    ├── webhook-confirmation/
    └── access-management/
```

---

## 🎓 Key Concepts

### Auto-Linking
Automatic association of Stripe customers with BOME users based on email matching. Happens at:
- User registration
- Email verification
- Password setup
- OAuth2 login
- Session verification

### Retroactive Access Grant
When a customer is linked to a user after a subscription already exists, the system automatically grants video access.

### Idempotent Access Grants
Calling `GrantVideoAccess()` multiple times is safe - it checks existing access and only updates the source tracking.

### Source Tracking
Each access grant records its source (`session_verification`, `webhook`, `retroactive_linking`, etc.) for audit trails.

### Dual-Confirmation
Two independent systems (session verification + webhooks) confirm payment and grant access for reliability.

---

## 🤝 Contributing

When making changes to the subscription flow:

1. ✅ Maintain idempotency
2. ✅ Update source tracking
3. ✅ Add logging at key points
4. ✅ Test both OAuth2 and email/password flows
5. ✅ Update this BRAID documentation
6. ✅ Add SQL migration if needed
7. ✅ Update QUICK_REFERENCE.md with new patterns

---

## 📞 Support

**For Bugs:** Check QUICK_REFERENCE.md first, then trace through FLOW_DIAGRAM.md

**For Features:** Read BRAID.md, understand dual-confirmation pattern

**For Onboarding:** Start with README.md (this file), then BRAID.md, then FLOW_DIAGRAM.md

---

## 📜 Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2025-11-18 | Fixed email/password parity bug |
| 1.5 | 2025-11-18 | Added retroactive access grant |
| 1.0 | 2025-11-17 | Initial dual-confirmation implementation |

---

**Maintained by:** BOME Development Team  
**Last Updated:** 2025-11-18  
**Status:** ✅ Production Ready

