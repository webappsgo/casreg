# casreg — project-specific rule overrides

This file wins over `AI.md` and over global `CLAUDE.md`.
Add an entry only when a rule here must actively differ from the template or global conventions.

## Overrides

None currently. All global and AI.md rules apply as-is.

---

## Support System Specification

## 📖 Table of Contents
1. [Purpose & Scope](#purpose--scope)
2. [Core Design Principles](#core-design-principles)
3. [User Roles & Permissions](#user-roles--permissions)
4. [Support Agent System](#support-agent-system)
5. [Bot Automation System](#bot-automation-system)
6. [Ticket Lifecycle](#ticket-lifecycle)
7. [Live Chat System](#live-chat-system)
8. [Knowledge Base](#knowledge-base)
9. [Notifications](#notifications)
10. [Security & Access Control](#security--access-control)
11. [Configuration Management](#configuration-management)
12. [Mobile & Accessibility](#mobile--accessibility)
13. [Audit & Compliance](#audit--compliance)
14. [Integration Points](#integration-points)
15. [Implementation Guidelines](#implementation-guidelines)
16. [Appendix](#appendix)

---

## 🎯 Purpose & Scope

### Purpose
This specification defines the support system for casreg — a self-hosted OCI and Incus image registry. It covers ticket management, bot pre-screening, live chat, knowledge base, and admin oversight as they apply specifically to casreg.

### Scope
casreg is a self-hosted single binary written in Go. This specification is grounded in casreg's actual architecture: chi router, GORM, SQLite/PostgreSQL/MySQL, and the existing `Ticket`, `TicketComment`, `User`, and `Notification` models.

### What This Specification Defines
- Behavioral requirements and user flows
- Data structures and fields extending casreg's existing models
- Security and access control rules
- User interface requirements and interactions
- Integration with casreg's existing `/v1/support/` routes
- Configuration options for `casreg.toml` and the web UI
- casreg-specific ticket categories (`verification-request` for Verified Publisher and Official Image requests)

---

## 🏗️ Core Design Principles

### 1. User Agency Over Automation
**Principle**: While automation enhances efficiency, users maintain ultimate control over all decisions.
- Bot suggestions can always be overridden
- No automated action is final without user confirmation
- Users can modify any auto-populated data
- Clear indication of what is automated vs. user-entered

### 2. Privacy by Design
**Principle**: User data and support interactions remain private and secure.
- Tickets visible only to creator and support agents
- Support agents' role is never exposed to users (they appear as named agents)
- All data access is logged and auditable
- Minimal data collection policy

### 3. Mobile-First, Accessibility-Always
**Principle**: The system must be fully functional on all devices and for all users.
- Responsive design that works on screens from 320px to 4K
- Full keyboard navigation support
- Screen reader compatibility (WCAG 2.1 AA compliant)
- Touch-optimized interfaces for mobile devices
- Reduced motion options for users with vestibular disorders

### 4. Configuration by Layer
**Principle**: Configuration belongs at the right level for its audience.

Server-level settings (listen address, database URL, storage backend, rate limits) are configured via `casreg.toml` or environment variables. User-facing preferences (notification settings, display theme, agent display name) are managed through the web UI. No configuration requires editing source code or restarting the server for user preferences; server config changes require a restart.

### 5. Fail Gracefully
**Principle**: When components fail, the system degrades gracefully.
- If bot fails → direct ticket creation remains available
- If chat fails → ticket system remains functional
- If email fails → in-app notifications continue
- Clear error messages for users and agents
- Automatic recovery when services restore

---

## 👤 User Roles & Permissions

casreg has three stored roles: `user`, `support`, and `admin`. Guest (unauthenticated) is a session state, not a stored role.

### Role Hierarchy

#### 1. Guest (Unauthenticated)
**Capabilities**:
- View public knowledge base articles
- Submit support tickets (with email verification)
- Access bot for initial troubleshooting
- View status of their tickets via email link

**Restrictions**:
- Cannot view other tickets
- Cannot access live chat
- May require CAPTCHA for ticket submission
- Rate-limited ticket creation (configurable in `casreg.toml`)

#### 2. User (`user` role)
**Capabilities**:
- All Guest capabilities
- Create tickets without email verification
- Access live chat when agents available
- View all their historical tickets
- Update their profile and notification preferences
- Rate and provide feedback on resolved tickets

**Restrictions**:
- Can only view their own tickets
- Cannot see agent-internal notes (`is_internal = true` comments)
- Cannot access support dashboard
- Subject to configured rate limits

#### 3. Support Agent (`support` role)
**Capabilities**:
- Toggle between user mode and support mode
- View and respond to all tickets (when in support mode)
- Access internal knowledge base and notes
- Use canned responses and templates
- Set availability status for live chat
- Add internal notes to tickets (`is_internal = true`)
- Assign and reassign tickets
- View support metrics and queue status

**Restrictions**:
- Cannot create tickets while in support mode
- Cannot modify system configuration
- Cannot delete tickets (only archive)
- Cannot view certain admin-only logs

#### 4. Administrator (`admin` role)
**Capabilities**:
- All Support Agent capabilities
- Configure system settings via web UI and `casreg.toml`
- Manage support agent accounts and grant the `support` role
- Customize email templates and notifications
- Configure bot settings (enable/disable, custom patterns)
- Access system logs and audit trails
- Manage knowledge base structure
- Set rate limits and security policies
- Grant Verified Publisher and Official Image badges (from `verification-request` ticket view)

**Restrictions**:
- When acting as support, appears as support agent (not admin)
- Cannot bypass audit logging
- Cannot access database directly through support interface

### Role Assignment

#### Method 1: Internal Assignment
Admin manually assigns roles through the web interface. Effect is immediate. Roles are stored in the `users.role` column.

#### Method 2: External Authentication Mapping
When OIDC is configured, claims from the identity provider can be mapped to casreg roles via admin configuration in the web UI. Mapping rules are evaluated on each login.

### Support Group Membership
- Users with the `support` role must switch to Support Mode to act as agents.
- Administrators have full support capabilities without a separate toggle, but adopt the same support-mode UI when acting in a support capacity.
- Neither role is ever exposed to end users — agents appear by their configured display name only.

---

## 👥 Support Agent System

### Support Mode Toggle

#### Entering Support Mode
**Trigger Methods**:
1. Manual toggle switch in navigation bar
2. Accessing the support dashboard URL directly
3. Clicking "Enter Support Mode" from the user dashboard

**What Happens**:
1. System verifies the user has the `support` or `admin` role
2. UI transitions to support interface
3. Banner appears indicating support mode is active
4. Available actions change to agent actions
5. Session flag `support_mode` set to `true`

#### Support Mode Banner
**Location**: Top of every page (static, not sticky — scrolls with content)

**Desktop Display**:
```
┌─────────────────────────────────────────────────────────────────┐
│ 🎧 SUPPORT AGENT MODE ACTIVE - Viewing as: [Agent Display Name]  │
│ You cannot create tickets while in support mode                  │
│ Active tickets in queue: 12 | Your assigned: 3                  │
│                                         [Exit Support Mode]      │
└─────────────────────────────────────────────────────────────────┘
```

**Mobile Display**:
```
┌────────────────────────────┐
│ 🎧 SUPPORT MODE            │
│ As: [Name] | Queue: 12     │
│         [Exit Mode]        │
└────────────────────────────┘
```

**Banner Properties**:
- Background color: Configurable (default: blue/purple)
- Text color: High contrast based on background
- Font size: System default with 1.1× scaling
- Z-index: Below modals but above content
- Animation: Slide down on enter, slide up on exit
- Persistence: Remains across page navigations while in mode
- Mobile-friendly: Does not stick to viewport on scroll

#### In Support Mode

**Enabled Features**:
- Support dashboard access
- All tickets visible in queue
- Ticket filtering and search across all users
- Ability to reply to any ticket
- Internal notes on tickets (`is_internal = true`)
- Ticket assignment capabilities
- Canned response library (system and personal)
- Support metrics dashboard
- Live chat agent console
- Bulk ticket operations

**Disabled Features**:
- "Create New Ticket" button (disabled with tooltip explanation)
- Bot interaction flow
- Personal ticket view (only see all tickets)
- User preferences editing
- Normal user dashboard

**UI Differences**:
- Navigation bar shows support-specific menu
- Different color scheme (configurable)
- Browser tab title prefixed with "[Support]"
- Breadcrumbs show "Support > [Current Page]"
- Footer shows "Support Mode Active"

#### Exiting Support Mode

**Trigger Methods**:
1. Click "Exit Support Mode" in banner
2. Toggle switch in navigation
3. Logout (always returns to user mode on next login)

**What Happens**:
1. Confirmation if agent has unsaved work
2. UI transitions back to user interface
3. Banner disappears
4. Available actions return to normal
5. Session flag `support_mode` set to `false`
6. Any draft responses are saved for later

### Agent Display Names

#### Configuration
**Where**: Web UI → Support Settings → My Agent Profile

**Fields**:
- **Display Name** (required): What users see
  - Examples: "Sarah", "Tech Support — John", "Alex from Registry Team"
  - Character limit: 50 characters
  - Allowed characters: Letters, numbers, spaces, hyphens
- **Fallback**: If not set, shows "Support Agent"
- **Avatar** (optional): Image or initials
- **Specialty Tags** (optional): "Registry", "Billing", "Account", "Incus"

#### Display Rules
- In tickets: "Sarah replied to your ticket"
- In chat: "You're chatting with Sarah"
- In email: "From: Sarah (Support Team)"
- Never shows role or admin status to users

### Agent Availability

#### Status Options
1. **Available** (Green) — Can receive chats; visible in chat queue; new tickets can be auto-assigned
2. **Busy** (Yellow) — No new chats; finishing current conversations; still working tickets
3. **Away** (Gray) — No chats or auto-assignments; manual ticket work only; shows "Back at [time]" if set
4. **Offline** (Red) — Not logged in or support mode off; not counted in available agents

#### Automatic Status Management
- Set to "Away" after 15 minutes of inactivity (configurable in `casreg.toml`)
- Set to "Offline" on logout or session timeout
- Optional: Set to "Busy" when reaching chat limit
- Returns to previous status on activity

---

## 🤖 Bot Automation System (Logic-Based, No ML)

### Core Principle: Deterministic Logic Only

The production bot uses pure pattern matching — regex, exact string comparison, keyword density analysis. There is no machine learning inference at runtime. The bot works completely offline and produces 100% predictable responses.

### Architecture

#### Components

**1. Knowledge Base Scanner**
- Indexes all documentation on startup
- Refreshes index when documents change
- Creates searchable patterns from content
- Maintains relevance scoring

**Security Note**: The bot's pattern database, error codes, and solution mappings are never exposed through any endpoint, API, or interface.

**2. Pattern Database** (Internal — completely isolated)

**Access Restrictions**:
- Not exposed via any API endpoint
- Not accessible through web UI
- Not visible to support agents
- Only the bot process reads it

**Storage**: Compiled into the casreg binary as Go constants or embedded files. The format is an implementation detail; behavioral requirements only are specified here.

**Security Requirements**:
- No network exposure
- Read-only from bot process
- Updated only through redeployment
- No logging of actual patterns

**3. Pattern Matching Engine**
- Regex-based pattern matching
- Exact string matching for error codes
- Keyword density analysis
- Confidence scoring (0–100)
- Responds only at 100% confidence
- Patterns loaded from compiled storage; no external dependencies

**4. Conversation Manager**
- Maintains conversation context
- Tracks attempted solutions
- Generates ticket payload
- Maximum 3 attempts before creating ticket

### Built-in Pattern Recognition

**Hybrid Approach: Universal + casreg-Specific Patterns**

The bot uses two layers:
1. **Universal Patterns** (built-in) — work for any system
2. **casreg Patterns** (compiled in) — specific to registry operations

**Pre-defined Universal Patterns**:

**Authentication Issues**:
- Patterns: "can't log in", "cannot login", "authentication failed", "invalid password", "account locked"
- Solutions: Clear cache, reset password, check account status

**Performance Issues**:
- Patterns: "slow", "loading forever", "takes too long", "not responding", "frozen"
- Solutions: Check connection, clear cache, try different browser, check system status

**Access/Permission Issues**:
- Patterns: "access denied", "not authorized", "permission denied", "can't access", "forbidden"
- Solutions: Verify account permissions, check subscription status, contact admin for access

**Error Messages**:
- Patterns: Any message containing "error", "exception", "failed to", specific error codes
- Solutions: Searches knowledge base for error code, provides standard troubleshooting

**Data/Sync Issues**:
- Patterns: "not syncing", "data missing", "lost my", "disappeared", "not saving"
- Solutions: Check connection, refresh, verify auto-save settings

**How-to Questions**:
- Patterns: "how do I", "how to", "where is", "can't find"
- Solutions: Searches documentation, provides navigation help

**Account Management**:
- Patterns: "delete account", "change email", "update profile"
- Solutions: Links to account settings, step-by-step guide

**casreg-Specific Patterns**:

**OCI Push/Pull Issues**:
- Patterns: "docker push", "docker pull", "manifest unknown", "blob upload", "digest mismatch"
- Solutions: Verify registry URL format, check auth token, confirm repository exists

**Incus Image Issues**:
- Patterns: "incus image", "lxc image", "simplestreams", "alias not found", "image import"
- Solutions: Verify alias format (`namespace/os/release/variant`), check registry URL

**Robot Account Issues**:
- Patterns: "robot account", "service token", "CI token", "pipeline push"
- Solutions: Verify token scope, check expiry, confirm repository access

**Namespace/Verification Issues**:
- Patterns: "verified badge", "official image", "publisher verification", "namespace reserved"
- Solutions: Links to verification process, explains `verification-request` ticket category

**Image Scanning**:
- Patterns: "scan results", "CVE", "vulnerability", "security scan"
- Solutions: Explains scan status flow, links to scan results view

**Storage Quota**:
- Patterns: "quota exceeded", "storage full", "upload failed", "disk space"
- Solutions: Check current quota usage, contact admin for increase

### Smart Pattern Matching Logic

**Built-in Intelligence** (no configuration required):
1. **Error Code Detection**: Automatically extracts patterns like "ERR_*", "ERROR *", "[0-9]{3,5}"
2. **URL Detection**: Identifies problematic URLs or registry paths mentioned
3. **Timestamp Detection**: Recognizes when issues occurred
4. **Frequency Detection**: Identifies "always", "sometimes", "randomly"
5. **Urgency Detection**: Recognizes "urgent", "ASAP", "production down", "critical"
6. **Sentiment Analysis**: Simple negative keyword detection for escalation

**Category Auto-Detection**:
- `technical-issue` (errors, bugs, performance, push/pull failures)
- `account-management` (login, profile, permissions)
- `billing-inquiry` (quota, storage, subscription)
- `general-question` (how-to, features, documentation)
- `verification-request` (verified badge, official image, publisher verification)
- `security-concern` (CVE, vulnerability, scan results, compromise)

### Bot Interaction Flow

#### Phase 1: Initial Contact (Mandatory Before Ticket Creation)
```
User: Clicks "Get Support"
     ↓
Bot: "Hello! I'm the casreg support bot. I'll try to help you resolve
      your issue quickly. Please describe what you're experiencing."
     ↓
User: Describes issue
     ↓
Bot: Analyzes input
```

#### Phase 2: Solution Attempts (Maximum 3)

**Attempt Structure**:
```
Bot: "I found something that might help with [issue summary]:"
     
     📚 Relevant Article: [Title with link]
     
     Solution Steps:
     1. [Step one with specific action]
     2. [Step two with expected result]
     3. [Step three if needed]
     
     ⚡ Quick Actions:
     [Button: Clear Cache] [Button: Reset Password] [Button: Check Status]
     
     Did this resolve your issue?
     [Yes, resolved] [No, still having issues]
```

**If User Says "No"**:
- Bot attempts up to 2 more different solutions
- Each attempt must be substantially different
- Bot tracks what has been tried

**If User Says "Yes"**:
- Bot confirms resolution
- Asks for optional feedback
- Logs successful resolution
- Ends conversation

#### Phase 3: Ticket Creation (After 3 Failed Attempts)

**Bot Response**:
```
Bot: "I understand this issue needs human attention. I'll help you create a
      support ticket with the information we've gathered."
      
      Here's what I've prepared:
      • Issue Summary: [Generated from conversation]
      • Solutions Attempted: [List of tried solutions]
      • Category: [Auto-detected category]
      • Priority: [Based on urgency keywords]
      
      [Open Ticket Form]
```

**Generated Ticket Payload** (pre-fills the form; user must confirm):
- Title (auto-generated from issue keywords)
- Description (user's original description + context)
- Category (auto-detected, user can change)
- Priority (based on keywords: `critical` for production/urgent, `medium` default)
- Bot conversation history stored in `bot_metadata` JSON field
- Solutions already attempted
- Metadata (timestamp, user agent, referrer)

**User Control**:
- User can modify ALL fields before submission
- User must click "Submit Ticket" — bot never saves automatically

### Optional Admin Overrides

Admins can optionally configure via the web UI:
- **Disable** specific built-in patterns (not modify them)
- **Add** casreg-instance-specific patterns
- **Set** category priorities
- **Configure** escalation keywords
- **Toggle** bot on/off entirely

**Admin UI**:
```
Bot Configuration
━━━━━━━━━━━━━━━━
Status: [● Enabled ○ Disabled]

Built-in Patterns: ✓ Active (Recommended)

Custom Patterns (Optional):
[+ Add Instance-Specific Pattern]

Escalation Words (Optional):
[Default list + Add custom words]

[Save Settings]
```

### Bot Limitations

**Cannot Do**:
- Make changes to user accounts or registry data
- Access private user data
- Close or resolve tickets
- Save tickets (user must confirm)
- Override user decisions
- Respond below 100% confidence
- Access external services

**Must Do**:
- Clearly identify as bot
- Log all interactions (stored in `bot_metadata` on the resulting ticket)
- Respect user choices
- Provide path to human support
- Maintain conversation context
- Generate accurate ticket summaries

### Bot Analytics

**Tracked Metrics**:
- Resolution rate (solved without ticket)
- Average attempts before resolution
- Most common patterns matched
- Failed pattern matches
- User satisfaction scores
- Time to resolution
- Category distribution
- User override frequency (when they change bot's categorization)

---

## 📋 Ticket Lifecycle

### States & Transitions

#### State Definitions

**`draft`** (user-side only)
- Ticket being composed; auto-saved every 30 seconds
- Not visible to support; can be abandoned

**`open`**
- Newly submitted; awaiting initial agent response
- Visible in support queue; SLA timer starts

**`in-progress`**
- Agent has claimed the ticket and is actively working
- Maps to casreg's existing `in-progress` constant

**`awaiting-user`**
- Agent has responded and needs user input
- SLA timer paused; reminder sent after configurable days

**`awaiting-agent`**
- User has responded; back in agent queue
- SLA timer resumes; priority may increase

**`resolved`**
- Solution provided; awaiting user confirmation
- Can be reopened; triggers satisfaction survey
- Maps to casreg's existing `resolved` constant

**`closed`**
- Confirmed resolved or auto-closed after 7 days in `resolved` state
- Read-only; archived after configurable retention period
- Can be reopened
- Maps to casreg's existing `closed` constant

**`reopened`**
- User reactivated a closed ticket
- Returns to `open`; maintains history; higher priority flag

#### State Transition Rules

```
draft         → open:            User submits
open          → in-progress:     Agent claims or starts work
in-progress   → awaiting-user:   Agent responds, needs user input
awaiting-user → awaiting-agent:  User responds
awaiting-agent → awaiting-user:  Agent responds
*             → resolved:        Agent marks resolved
resolved      → closed:          User confirms or auto-close after 7 days
closed        → reopened:        User requests (system transitions to open)
*             → closed:          Admin force-close
```

### Ticket Data Model

#### Existing Fields (casreg model)
- `UserID` — creator
- `Title` — brief description (max 200 chars)
- `Description` — full details (max 10,000 chars)
- `Status` — current state (see above)
- `Priority` — `low` | `medium` | `high` | `critical`
- `Category` — from predefined list (see below)
- `AssignedTo` — agent user ID (nullable)
- `ResolvedAt` — timestamp (nullable)
- `ClosedAt` — timestamp (nullable)

#### Fields to Add
- `Tags` — flexible text array for labeling
- `BotMetadata` — JSON blob of the bot conversation that led to this ticket
- `Resolution` — text summary of how the ticket was solved (filled by agent)
- `TimeSpent` — agent work time in minutes
- `ReopenedAt` — timestamp of most recent reopen (nullable)

#### Categories (casreg-specific)
- `technical-issue` — push/pull errors, registry errors, performance
- `feature-request` — requests for new functionality
- `account-management` — login, profile, robot accounts, tokens
- `billing-inquiry` — quota, storage limits
- `security-concern` — CVEs, vulnerability disclosures, account compromise
- `general-question` — documentation, how-to
- `verification-request` — Verified Publisher or Official Image badge requests (casreg-specific; admin grants badge directly from ticket view)

#### TicketComment: Existing Fields
- `Comment` — text content
- `IsInternal` — agent-only note (hidden from ticket creator)
- `IsResolution` — marks the comment that resolved the ticket

#### TicketComment: Fields to Add
- `EditedAt` — timestamp of last edit (nullable; edit window: 5 minutes)
- `Attachments` — references to uploaded files

#### Thread Structure
Each ticket maintains a threaded conversation with: user messages, agent messages, system events (state changes), internal notes (agent-only), timestamps, edit history, and attachments.

### Ticket Operations

#### User Operations
- View own tickets; add replies; upload attachments (size/type limits configurable)
- Request reopening; rate resolved tickets
- Modify ticket details before submission
- Cannot delete tickets, view other users' tickets, see internal notes, or change assignment

#### Agent Operations
- View all tickets; assign/reassign; change priority
- Add internal notes; merge related tickets
- Add tags; set reminders; use canned responses; link tickets
- Require admin permission to: delete tickets, bypass SLA, access deleted tickets

### SLA & Escalation (Optional — configurable in `casreg.toml`)

#### SLA Levels
```
critical:  First response: 1 hour,  Resolution: 4 hours
high:      First response: 4 hours, Resolution: 1 day
medium:    First response: 1 day,   Resolution: 3 days
low:       First response: 3 days,  Resolution: 7 days
```

#### Automatic Escalation Triggers
- SLA breach imminent (80% of time elapsed)
- SLA breached
- User mentions escalation keywords ("urgent", "production down")
- Multiple reopens
- `verification-request` tickets: auto-flag for admin attention

#### Escalation Actions
- Increase priority; notify senior agents; add to priority queue
- Send admin alert; auto-assign to available senior agent

---

## 💬 Live Chat System

### Availability Logic

```
Chat Available = (
  At least one agent status == "Available" AND
  Total active chats < max_concurrent_chats AND
  Current time within business hours (if configured) AND
  Chat feature enabled in casreg.toml
)
```

#### User Interface Changes

**When Available**:
```
[💬 Live Chat — Agent Available]
Click to start a conversation
```

**When Unavailable**:
```
[📧 Leave a Message]
No agents available — Create a ticket
Expected response time: ~2 hours
```

### Chat Initialization

#### Pre-chat Form (Optional)
```
┌─────────────────────────────┐
│ Start Live Chat             │
├─────────────────────────────┤
│ Name: [_______________]     │
│ Email: [______________]     │
│ Topic: [Dropdown      ▼]    │
│                             │
│ [Start Chat] [Cancel]       │
└─────────────────────────────┘
```

#### Chat Assignment
1. System finds available agents
2. Applies routing rules: least busy, round-robin, specialty matching, or previous agent
3. Assigns to selected agent and notifies immediately

### Chat Interface

#### User View
```
┌────────────────────────────────────┐
│ Support Chat — Sarah               │
├────────────────────────────────────┤
│                                    │
│ Sarah: Hi! How can I help today?  │
│                           2:34 PM  │
│                                    │
│ You: I can't push to my registry  │
│                           2:35 PM  │
│                                    │
│ Sarah is typing...                 │
│                                    │
├────────────────────────────────────┤
│ [Type message... (Enter to send,   │
│  Shift+Enter for new line)]        │
│ [📎] [Send]                        │
└────────────────────────────────────┘
```

#### Agent View
```
┌────────────────────────────────────┐
│ Chat with: John Doe (#USR-0923)    │
│ Email: john@example.com            │
│ Previous tickets: 3 [View]         │
├────────────────────────────────────┤
│ [Same chat content]                │
├────────────────────────────────────┤
│ Quick Actions:                     │
│ [Create Ticket] [Send Article]     │
│ [Transfer Chat] [End Chat]         │
├────────────────────────────────────┤
│ Canned Responses: [Select    ▼]    │
│ [Type message...]                  │
│ [📎] [Send]                        │
└────────────────────────────────────┘
```

### Chat Features

#### Message Features
- Real-time delivery
- Read receipts (optional)
- Typing indicators
- File sharing (size limits apply)
- Emoji support
- Link detection and preview
- Code formatting with triple backtick blocks
- Message editing within 5 minutes
- Enter to send; Shift+Enter for new line

#### Agent Tools
- **Canned Responses**: Quick access to system and personal templates; search/filter by category; auto-suggest based on content; one-click insertion with variable replacement
- **Article Insertion**: Share KB articles inline
- **Transfer**: Hand off to another agent
- **Convert to Ticket**: Create ticket from chat
- **User Info Panel**: History, tickets, notes

### Chat Conclusion

#### Ending a Chat

**Agent Ends Chat**: Confirmation prompt; optional wrap-up notes; optional conversion to ticket; satisfaction survey triggered.

**User Ends Chat**: Confirmation prompt; option to download transcript; option to create ticket; saved to history.

**System Timeout**: Warning after 10 minutes of inactivity; auto-end after 15 minutes; transcript saved; both parties notified.

#### Post-Chat
- Transcript saved to user's history
- Optionally converted to ticket
- Available for download
- Searchable in support dashboard

---

## 📚 Knowledge Base

### Structure

#### Public Knowledge Base
**Access**: Everyone (including unauthenticated users)
**Content Types**: Getting started guides, FAQs, troubleshooting articles, feature documentation

Served via `/v1/support/docs` and `/v1/support/docs/{id}`.

#### Internal Knowledge Base
**Access**: Support agents (`support` and `admin` roles) only
**Content Types**: Internal procedures, escalation guides, known issues, workarounds, customer notes

### Article Format

#### Metadata Header
```
title: "How to Push Images to casreg"
category: "Getting Started"
tags: ["docker", "push", "registry", "OCI"]
author: "Support Team"
created: 2024-01-15
updated: 2025-06-21
visibility: "public" or "internal"
priority: 1
related: ["article-id-1", "article-id-2"]
```

#### Content
Articles written in Markdown with: clear headings, step-by-step instructions, screenshots (where applicable), related articles, troubleshooting sections.

### Search & Discovery

#### Search Algorithm
1. Title match (highest weight)
2. Tag match (high weight)
3. Content match (medium weight)
4. Related articles (low weight)

#### Bot Integration
- Bot searches KB before attempting solutions
- Articles ranked by relevance and success rate
- Bot tracks which articles resolve issues
- Failed articles receive lower ranking

#### User Features
- Search as you type
- Category filtering
- Most helpful articles
- Recently updated
- Related articles sidebar

### Management

#### Article Lifecycle
1. **Draft** — being written
2. **Review** — awaiting approval
3. **Published** — live and searchable
4. **Archived** — outdated but preserved

#### Analytics Tracked
- View count
- Helpful/not helpful votes
- Bot usage frequency
- Resolution success rate
- Search queries leading to article
- Time on page

---

## 📬 Notifications

Support system notifications integrate with casreg's existing `Notification` model and in-app notification framework. Email delivery uses casreg's configured SMTP/email provider settings in `casreg.toml`.

### Notification Channels

#### 1. Email Notifications
**Sending Triggers**: Real-time or batched per user preferences

**Required Notification Types**:
- Ticket created confirmation
- Agent response on ticket
- Ticket status changed
- Ticket resolved
- Chat transcript delivered
- SLA warning (agent-facing)

**Template Variables** (illustrative — resolved at send time):
- `[recipient name]` — user's display name
- `[ticket ID]` — ticket identifier
- `[ticket title]` — ticket subject
- `[agent name]` — responding agent's display name
- `[response content]` — latest message
- `[ticket link]` — direct URL to ticket
- `[unsubscribe link]` — opt-out link

#### 2. In-App Notifications
**Display**: Badge counter + dropdown (casreg's existing notification UI)
**Persistence**: Until marked read
**Types**:
- New response on ticket
- Ticket status change
- Chat request (agent-facing)
- System announcements

#### 3. Browser Push (Optional)
- Requires user permission
- Immediate for urgent items
- Brief content with action link

### Notification Rules

#### User Preferences
Configurable per user in the web UI:
- Email for ticket updates
- Email for status changes
- In-app notifications
- Push notifications
- Quiet hours
- Batching preferences

#### Smart Batching
- Combine multiple updates within 5 minutes
- Never batch critical SLA notifications
- Respect user's time zone
- Configurable quiet hours (default: do not send between 22:00–08:00 local time)

---

## 🔐 Security & Access Control

### Authentication Integration

casreg's existing JWT-based session system applies to the support system. Support mode uses the same session token with an added `support_mode` claim.

**Supported Methods**:
1. Session-based: casreg's existing JWT sessions
2. OIDC: Claims mapped to roles via admin configuration
3. API tokens: Robot accounts for programmatic ticket creation (scoped to `support:write`)

#### Session Management
- Support sessions inherit casreg's main session timeout
- Force re-auth for sensitive operations (e.g., bulk ticket delete)
- Concurrent session limits configurable in `casreg.toml`

### Data Access Controls

#### Bot System Isolation
**Complete Separation**:
- Bot pattern database has no API endpoints
- No URL routes to bot internals
- No database queries expose bot patterns
- Support agents cannot view bot logic
- Admins configure bot via UI only; never see compiled patterns

**Why This Matters**:
- Prevents pattern exploitation
- Avoids social engineering (users cannot reverse-engineer trigger logic)
- Reduces attack surface

#### Ticket Visibility Matrix
```
              View Own | View All | Modify Own | Modify All
Guest         Limited  |    No    |     No     |     No
User            Yes    |    No    |    Yes     |     No
Support (Off)   Yes    |    No    |    Yes     |     No
Support (On)    N/A    |   Yes    |     No     |    Yes
Admin           N/A    |   Yes    |     No     |    Yes
```

#### Sensitive Data Handling
- PII masked in logs
- Passwords never stored in tickets or comments
- Attachments scanned before storage
- Data retention policies enforced per configuration

### Rate Limiting

Configurable in `casreg.toml`:

```toml
[support.rate_limit]
guest_ticket_create  = "3/hour"
guest_kb_search      = "30/min"
guest_file_upload_mb = 10

user_ticket_create   = "10/hour"
user_chat_messages   = "60/min"
user_file_upload_mb  = 25

# support and admin roles have no rate limits in support mode
```

### Audit Logging

#### Events Logged (support-specific additions to casreg's existing audit log)
- All ticket state changes
- Agent actions (assign, reply, close, force-close)
- Configuration changes
- Failed authentication attempts
- Permission changes
- Data exports
- Bulk operations
- Bot interactions

#### Log Format
Structured logging with: timestamp, event type, actor (user ID + display name), target (ticket ID), details, IP address, user agent.

---

## ⚙️ Configuration Management

### Server-Level Settings (`casreg.toml`)

```toml
[support]
enabled              = true
bot_enabled          = true
chat_enabled         = true
guest_tickets        = false
file_attachments     = true
satisfaction_surveys = true
max_ticket_chars     = 10000
max_attachment_mb    = 25
ticket_rate_limit    = 10        # per hour, for user role
chat_timeout_min     = 15
sla_auto_close_days  = 7

[support.business_hours]
timezone      = "UTC"
weekday_start = "09:00"
weekday_end   = "17:00"
weekend       = false
```

### Web UI Configuration

#### System Settings (Admin Only)
Configurable without restart:
- Support email sender name and address
- Bot on/off toggle and maximum attempts (1–5)
- Chat routing rules (least-busy, round-robin, specialty)
- Email template content
- Custom bot patterns
- Escalation keyword list
- Knowledge base categories

#### Canned Responses Management (Admin Only)

**Hierarchy**:
1. System responses (admin-created, available to all agents)
2. Department responses (admin-created, scoped to specific teams)
3. Personal responses (agent-created, individual use only)

**Admin Controls**:
```
System Canned Responses
━━━━━━━━━━━━━━━━━━━━━
[+ Add New Response]

Category: [Greeting ▼]
Title: Welcome Message
Content: [Rich text editor]
Tags: [greeting, initial]
Available to: [All Agents ▼]
Status: [Active ▼]

[Save] [Preview] [Delete]
```

**Variables** available in templates: `[user name]`, `[ticket ID]`, `[agent name]`, `[registry name]`

#### Agent Personal Settings

```
My Support Profile
━━━━━━━━━━━━━━━━

Display Name: [Sarah from Registry Support]
Avatar: [Upload Image]
Specialties: [Registry, Incus, Billing]
Signature: [Best regards,\nSarah]

Preferences
├── ☑ Sound for new tickets
├── ☑ Desktop notifications
├── ☐ Auto-accept chats
└── ☑ Show keyboard shortcuts

Canned Responses
├── System Responses (View Only)
│   ├── "Welcome to casreg support..." [View]
│   └── "Escalating to admin..." [View]
├── Personal Responses
│   ├── [+ Add Personal Response]
│   ├── "Thanks for contacting..." [Edit]
│   └── "For registry push issues..." [Edit]
```

---

## 📱 Mobile & Accessibility

### Responsive Design

#### Breakpoints
- Mobile: 320px–768px
- Tablet: 769px–1024px
- Desktop: 1025px+

#### Mobile Optimizations
- Touch-friendly buttons (minimum 44×44px)
- Swipe gestures for ticket actions
- Bottom sheet for quick actions
- Simplified navigation
- Non-sticky support mode banner (scrolls with content)
- Reduced data mode option

### Accessibility Standards

#### WCAG 2.1 AA Compliance
- Color contrast ratios (4.5:1 minimum)
- Focus indicators visible
- Skip navigation links
- Semantic HTML structure
- ARIA labels and roles
- Error messages associated with inputs

#### Keyboard Navigation
```
Tab         — Next element
Shift+Tab   — Previous element
Enter       — Activate button/link
Space       — Checkbox, button
Arrow keys  — Radio buttons, menus
Escape      — Close modal/dropdown
/ key       — Focus search
? key       — Show keyboard shortcuts
```

#### Screen Reader Support
- Announce status changes
- Describe form requirements
- Read error messages
- Indicate required fields
- Provide context for links
- Alternative text for images

---

## 📊 Audit & Compliance

### Data Retention

Configurable in `casreg.toml` or via admin UI:

| Data type          | Default retention     |
|--------------------|-----------------------|
| Active tickets     | No limit              |
| Closed tickets     | 2 years               |
| Deleted tickets    | 30 days (soft delete) |
| Chat transcripts   | 1 year                |
| Audit logs         | 7 years               |
| System logs        | 90 days               |
| Active attachments | No limit              |
| Closed attachments | 1 year                |
| Orphaned files     | 30 days               |

### Privacy Compliance

#### User Data Rights
- Right to access: export all user ticket data as JSON
- Right to deletion: purge user PII from tickets (anonymise assignee/author references)
- Data portability: JSON export includes all tickets, comments, and chat transcripts
- Privacy policy acceptance tracked at account creation

#### Data Export Format
Standardized JSON export including: user profile, all tickets with comments, chat transcripts, attachment references, activity history.

### Compliance Reports

Available via Admin Dashboard:
- Agent activity summaries
- SLA compliance percentages
- Response time averages
- Resolution rates
- User satisfaction scores
- Audit log extracts
- Data retention compliance

---

## 🔌 Integration Points

Support functionality is exposed under casreg's `/v1/support/` route prefix. All routes below are required by this specification; the implementation must provide them.

### Required Routes

#### Tickets

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/support/tickets` | user/support/admin | List tickets — users see own; agents see all |
| `POST` | `/v1/support/tickets` | user/support/admin | Create ticket |
| `GET` | `/v1/support/tickets/{id}` | owner/support/admin | Get ticket with comment thread |
| `PUT` | `/v1/support/tickets/{id}` | owner (before assignment) / admin | Update title, description, category |
| `DELETE` | `/v1/support/tickets/{id}` | admin | Delete ticket |
| `POST` | `/v1/support/tickets/{id}/comments` | owner/support/admin | Add comment or internal note |
| `PUT` | `/v1/support/tickets/{id}/comments/{cid}` | author (5-min window) | Edit comment |
| `DELETE` | `/v1/support/tickets/{id}/comments/{cid}` | admin | Delete comment |
| `POST` | `/v1/support/tickets/{id}/assign` | support/admin | Assign ticket to agent |
| `POST` | `/v1/support/tickets/{id}/resolve` | support/admin | Mark resolved with resolution comment |
| `POST` | `/v1/support/tickets/{id}/close` | support/admin | Close ticket |
| `POST` | `/v1/support/tickets/{id}/reopen` | owner/admin | Reopen closed ticket |
| `POST` | `/v1/support/tickets/{id}/attachments` | owner/support/admin | Upload attachment |
| `GET` | `/v1/support/tickets/{id}/attachments/{aid}` | owner/support/admin | Download attachment |

#### Knowledge Base

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/support/docs` | public (public articles); support/admin (internal) | List KB articles |
| `GET` | `/v1/support/docs/{id}` | public / support/admin | Get KB article |
| `POST` | `/v1/support/docs` | support/admin | Create KB article |
| `PUT` | `/v1/support/docs/{id}` | support/admin | Update KB article |
| `DELETE` | `/v1/support/docs/{id}` | admin | Archive KB article |

#### Live Chat

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/support/chat/status` | public | Chat availability |
| `POST` | `/v1/support/chat/sessions` | user/support/admin | Start chat session |
| `WS` | `/v1/support/chat/sessions/{id}` | session owner / agent | WebSocket connection |
| `POST` | `/v1/support/chat/sessions/{id}/end` | session owner / agent | End chat |
| `POST` | `/v1/support/chat/sessions/{id}/transfer` | support/admin | Transfer to another agent |

#### Canned Responses

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/support/canned` | support/admin | List canned responses (system + personal) |
| `POST` | `/v1/support/canned` | support/admin | Create canned response |
| `PUT` | `/v1/support/canned/{id}` | admin (system) / author (personal) | Update |
| `DELETE` | `/v1/support/canned/{id}` | admin (system) / author (personal) | Delete |

#### Agent Management

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/support/agents` | support/admin | List agents and availability status |
| `PUT` | `/v1/support/agents/me/status` | support/admin | Update own availability |
| `GET` | `/v1/support/metrics` | admin | Support metrics dashboard data |

### Incoming Integrations

The support API supports programmatic ticket creation via robot accounts with the `support:write` token scope. Authentication follows casreg's existing token model.

### Outgoing Integrations (Webhooks)

Configurable events:
- `ticket.created`
- `ticket.status_changed`
- `ticket.commented`
- `chat.started`
- `chat.ended`
- `sla.breached`

**Webhook Payload**:
```json
{
  "event": "ticket.created",
  "timestamp": "2025-06-21T12:00:00Z",
  "data": {
    "ticket_id": 1234,
    "title": "Cannot push to registry",
    "user_id": 42,
    "category": "technical-issue"
  },
  "signature": "HMAC-SHA256"
}
```

---

## 🚀 Implementation Guidelines

### Database Additions

Fields to add to existing GORM models (see [Ticket Data Model](#ticket-data-model) above for full list):

**`tickets` table**:
- `tags TEXT` — JSON array
- `bot_metadata TEXT` — JSON blob of bot conversation
- `resolution TEXT` — agent-written resolution summary
- `time_spent INTEGER` — agent minutes
- `reopened_at DATETIME` — nullable timestamp
- Additional statuses: `draft`, `awaiting-user`, `awaiting-agent`, `reopened`

**`ticket_comments` table**:
- `edited_at DATETIME` — nullable; set on edit within 5-minute window
- `attachments TEXT` — JSON array of attachment references

**New tables** (not in current schema):
- `chat_sessions` — live chat conversations
- `chat_messages` — individual chat messages
- `kb_articles` — knowledge base articles (public and internal)
- `canned_responses` — system and personal response templates
- `agent_availability` — current status per agent
- `support_tickets_tags` — tag index (or use JSON column above)

The `support` role must be added as a valid `users.role` value alongside existing `user` and `admin`.

### Performance Requirements

| Target | Limit |
|--------|-------|
| Page load | < 2 seconds |
| Search results | < 500ms |
| Chat message delivery | < 100ms |
| Auto-save interval | 30 seconds |

### Scalability

casreg is a self-contained single binary. It runs on a single node. The embedded scheduler handles async work (SLA checks, auto-close, notification batching). Horizontal scaling and external caching are non-goals; optimize for low-memory, low-dependency operation.

### Testing Requirements

#### Functional Testing
- All user flows (ticket creation, bot flow, chat, KB)
- State transitions (all 9 states, all transitions)
- Permission matrix enforcement
- Bot conversation paths (all 3 attempt cycles)
- Chat scenarios (available, unavailable, timeout)
- Support mode toggle (banner, disabled actions)
- Canned responses (system, personal, variable substitution)

#### Non-Functional Testing
- Load testing (concurrent ticket submissions)
- Security testing (OWASP Top 10, insecure direct object references on ticket IDs)
- Accessibility testing (screen reader, keyboard-only navigation)
- Mobile device testing (320px minimum width)

### Deployment Checklist

#### Pre-Launch
- [ ] New ticket statuses and fields added to GORM models and migrations
- [ ] `support` role added to role constants and validation
- [ ] Bot pattern database compiled in
- [ ] `casreg.toml` support section documented and defaults set
- [ ] Email templates configured
- [ ] Bot patterns tested with sample inputs
- [ ] Knowledge base seeded with casreg-specific articles
- [ ] Support agents assigned `support` role

#### Post-Launch
- [ ] Monitor bot resolution rate
- [ ] Review SLA compliance report after first week
- [ ] Adjust rate limits based on observed traffic
- [ ] Update KB articles based on common tickets

---

## 📋 Appendix

### Ticket Status Reference

| Status | Description | SLA Timer |
|--------|-------------|-----------|
| `draft` | Being composed | Not started |
| `open` | Submitted, unassigned | Running |
| `in-progress` | Agent claimed | Running |
| `awaiting-user` | Agent needs input | Paused |
| `awaiting-agent` | User replied | Running |
| `resolved` | Solution provided | Stopped |
| `closed` | Confirmed or auto-closed | Stopped |
| `reopened` | Reactivated by user | Running |

### Ticket Priority Reference

| Priority | First Response | Resolution | Use Case |
|----------|---------------|------------|----------|
| `low` | 3 days | 7 days | Minor questions, documentation |
| `medium` | 1 day | 3 days | General issues (default) |
| `high` | 4 hours | 1 day | Service degraded, blocking work |
| `critical` | 1 hour | 4 hours | Production down, data loss risk |

### Ticket Category Reference

| Category | Description | Auto-routed to |
|----------|-------------|----------------|
| `technical-issue` | Push/pull errors, registry errors | Any agent |
| `feature-request` | New functionality requests | Any agent |
| `account-management` | Login, profile, tokens | Any agent |
| `billing-inquiry` | Quota, storage limits | Any agent |
| `security-concern` | CVEs, vulnerability disclosures | Senior agent |
| `general-question` | Documentation, how-to | Any agent |
| `verification-request` | Verified Publisher/Official Image badge | Admin |

### Keyboard Shortcuts

```
In ticket list:
  J / K         — Next / previous ticket
  O             — Open selected ticket
  A             — Assign to me
  R             — Reply

In ticket view:
  R             — Focus reply box
  I             — Toggle internal note
  Escape        — Cancel compose

In support mode:
  /             — Focus search
  N             — Next unread ticket
  ?             — Show keyboard shortcut help
```

### Glossary

- **Agent** — Support team member with `support` or `admin` role, acting in support mode
- **Bot** — Automated response system (deterministic logic, no ML)
- **Canned Response** — Pre-written reply (system, department, or personal)
- **Escalation** — Increasing priority or urgency of a ticket
- **KB** — Knowledge Base
- **Pattern** — Regex or exact-string match rule for bot
- **SLA** — Service Level Agreement (response and resolution time targets)
- **Support Mode** — Special UI mode for agents; disables ticket creation
- **Thread** — Conversation within a ticket (user messages, agent replies, system events, internal notes)
- **Ticket** — A support request
- **Verification Request** — casreg-specific ticket category for requesting Verified Publisher or Official Image badge

### Version History

- v1.0 — Initial casreg-specific specification (adapted from generic support spec)
