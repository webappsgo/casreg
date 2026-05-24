The user wants me to create an extremely detailed and comprehensive description of the casreg project that is so complete and verbose that any AI system (including other models like ollama, localai, etc.) could take this description and generate the entire working codebase without needing any additional instructions or clarifications.

They want:
1. Every rule, definition, setting, intention
2. All features described in detail
3. OS and distro rules
4. Complete file structure
5. All source code requirements
6. Build instructions
7. Everything needed to create a working application on first build

They want it in a single raw format that they can copy and paste.

Looking at the original specification they provided, I need to expand on it significantly to make it much more detailed and comprehensive. I need to include:

1. Detailed technical specifications for every component
2. Exact API endpoints and their behaviors
3. Database schema details
4. Configuration options and their effects
5. UI/UX specifications
6. CLI command specifications
7. Build and deployment instructions
8. Error handling requirements
9. Security requirements
10. Performance requirements
11. Testing requirements
12. Documentation requirements

I should structure this as a comprehensive specification document that leaves no room for interpretation or missing details.

```
PROJECT SPECIFICATION: casreg - Self-Hosted Docker Registry Platform

=============================================================================
CRITICAL IMPLEMENTATION DIRECTIVE
=============================================================================

This document contains the COMPLETE and EXHAUSTIVE specification for the casreg application. Every component, feature, behavior, configuration option, file structure, code pattern, API endpoint, database schema, UI element, CLI command, security rule, validation logic, and operational requirement is explicitly defined herein. 

NO EXTERNAL DECISIONS, ASSUMPTIONS, OR INTERPRETATIONS ARE PERMITTED. Every implementation detail must conform EXACTLY to the specifications provided. Any ambiguity should be resolved by following the most explicit path defined in this document.

The resulting application must compile, run, and function correctly on the first build attempt without any modifications, additional configuration, or external dependencies beyond those explicitly specified.

=============================================================================
PROJECT OVERVIEW AND ARCHITECTURE
=============================================================================

PROJECT NAME: casreg
PROJECT TYPE: Self-hosted, full-featured Docker Registry platform with integrated authentication, scanning, and support systems

CORE MISSION: Provide a complete, production-ready Docker Registry platform that combines OCI-compliant container storage with comprehensive user management, security scanning, signature verification, real-time notifications, and integrated support systems, all delivered as a single, self-contained binary with optional web UI and CLI interfaces.

ARCHITECTURAL PRINCIPLES:
- Single binary deployment with embedded UI and CLI
- Database-agnostic with intelligent migration capabilities
- Storage-backend agnostic with seamless switching
- API-first design with complete parity across UI, CLI, and API
- Security-first approach with built-in scanning and signature verification
- Real-time capabilities with graceful degradation
- Zero-external-dependency operation with optional enhanced features
- Comprehensive self-documentation and support systems

=============================================================================
TECHNOLOGY STACK AND PLATFORM REQUIREMENTS
=============================================================================

PRIMARY LANGUAGE: Go (Golang) version 1.22 or higher
COMPILATION TARGET: Single static binary with no external dependencies
EMBEDDED UI: Modern web frontend using Svelte 4.x (preferred) or Vue 3.x
CLI FRAMEWORK: Bubbletea with Lipgloss for themed terminal user interface
API FRAMEWORK: Chi router with integrated middleware stack
DATABASE ORM: GORM with comprehensive driver support
DOCUMENTATION: Embedded Swagger UI with auto-generated specifications

SUPPORTED OPERATING SYSTEMS:
- Linux (all distributions): Native binary execution
  - Debian/Ubuntu: systemd service integration
  - RHEL/CentOS/Fedora: systemd service integration  
  - Alpine Linux: OpenRC service integration
  - Arch Linux: systemd service integration
  - Any Linux with glibc 2.17+ or musl libc
- macOS: Native binary with Homebrew distribution support
- Windows: Native binary execution via PowerShell/Command Prompt or WSL2
- Docker: Multi-architecture containers (linux/amd64, linux/arm64, linux/arm/v7)

CONTAINER COMPLIANCE: Full OCI (Open Container Initiative) specification compliance
REGISTRY COMPATIBILITY: Docker Registry HTTP API V2 complete implementation

=============================================================================
DATABASE SUPPORT AND MIGRATION SYSTEM
=============================================================================

DEFAULT DATABASE: SQLite 3.x with WAL mode enabled
- File location: configurable via CASREG_DATABASE_PATH (default: ./casreg.db)
- Connection pooling: 25 max open connections, 25 max idle connections
- Busy timeout: 30 seconds
- Pragma settings: journal_mode=WAL, synchronous=NORMAL, cache_size=100000

SUPPORTED DATABASE SYSTEMS:
1. PostgreSQL 12+ 
   - Full ACID compliance required
   - JSON/JSONB column support required
   - Connection string format: postgresql://user:pass@host:port/dbname?sslmode=mode
   
2. MySQL 8.0+ / MariaDB 10.5+
   - InnoDB storage engine required
   - utf8mb4 charset required
   - Connection string format: mysql://user:pass@host:port/dbname?charset=utf8mb4&parseTime=true
   
3. Microsoft SQL Server 2019+
   - Connection string format: sqlserver://user:pass@host:port?database=dbname
   
4. MongoDB 5.0+ (preferred for document storage)
   - Connection string format: mongodb://user:pass@host:port/dbname
   - Document-based storage for configurations and metadata
   - GridFS for large file storage

DATABASE MIGRATION SYSTEM:
- Automatic schema detection and version comparison
- Live migration with zero-downtime capability
- Pre-migration validation with rollback support
- Data integrity verification post-migration
- Migration progress tracking with detailed logging
- Support for custom migration scripts
- Backup creation before migration execution

MIGRATION UI INTERFACE:
- Connection string input with validation
- Database type dropdown selection
- Test connection functionality with detailed error reporting
- Migration preview with affected tables and record counts
- Progress bar with real-time status updates
- Rollback capability for failed migrations
- Migration history with timestamps and success/failure status

=============================================================================
STORAGE BACKEND ARCHITECTURE
=============================================================================

STORAGE ABSTRACTION LAYER:
- Pluggable interface supporting multiple backend types
- Automatic failover and retry mechanisms
- Content-addressable storage with deduplication
- Streaming upload/download with progress tracking
- Integrity verification using SHA256 checksums
- Compression support (gzip, lz4) with configurable levels

SUPPORTED STORAGE BACKENDS:

1. LOCAL FILESYSTEM STORAGE (default)
   - Base path: CASREG_STORAGE_PATH (default: /var/lib/casreg/storage)
   - Directory structure: {base_path}/{registry}/{repository}/{layer_digest}
   - File permissions: 644 for files, 755 for directories
   - Symlink support for deduplication
   - Automatic cleanup of orphaned files
   - Disk usage monitoring and quota enforcement

2. S3-COMPATIBLE STORAGE (AWS S3, MinIO, etc.)
   - Endpoint configuration: CASREG_STORAGE_S3_ENDPOINT
   - Bucket configuration: CASREG_STORAGE_S3_BUCKET
   - Region configuration: CASREG_STORAGE_S3_REGION (default: us-east-1)
   - Access credentials: CASREG_STORAGE_S3_ACCESS_KEY, CASREG_STORAGE_S3_SECRET_KEY
   - SSL/TLS configuration: CASREG_STORAGE_S3_USE_SSL (default: true)
   - Path-style addressing support for MinIO compatibility
   - Multipart upload support for large files (>5MB)
   - Server-side encryption support (AES256, KMS)
   - Lifecycle management integration

3. NFS NETWORK STORAGE
   - Server configuration: CASREG_STORAGE_NFS_SERVER
   - Mount path configuration: CASREG_STORAGE_NFS_PATH
   - NFS version support: NFSv3, NFSv4, NFSv4.1
   - Authentication: Kerberos, UNIX authentication
   - Automatic mount/unmount with retry logic
   - Network failure handling with graceful degradation

STORAGE CONFIGURATION VALIDATION:
- Connection testing before activation
- Permissions verification (read/write/delete)
- Available space checking and monitoring
- Performance benchmarking during setup
- Automatic configuration suggestion based on environment

=============================================================================
AUTHENTICATION AND AUTHORIZATION SYSTEM
=============================================================================

USER ROLES AND PERMISSIONS:

ADMIN ROLE (first user, non-demotable):
- Complete system administration access
- User management: create, modify, delete, suspend users
- Organization management: create, modify, delete organizations
- Registry management: access all registries regardless of ownership
- Configuration management: modify all system settings
- Support system: access all tickets and user communications
- Audit log access: view all system activities and security events
- Database management: perform migrations and maintenance operations

USER ROLE (default for all subsequent users):
- Profile management: modify own profile, password, preferences
- Organization creation: create and own organizations
- Registry creation: create personal and organizational registries
- Token management: create, rotate, delete personal access tokens
- Support access: create and manage own support tickets
- Limited audit access: view own activities only

ORGANIZATION ROLES:
- Owner: Full organization control, member management, registry creation
- Admin: Member management, registry management, cannot delete organization
- Member: Registry access based on organization permissions, no management rights

AUTHENTICATION METHODS:

1. USERNAME/PASSWORD AUTHENTICATION:
   - Minimum password length: 8 characters
   - Password complexity: at least one uppercase, lowercase, number, special character
   - Password hashing: bcrypt with cost factor 12
   - Failed login protection: account lockout after 5 failed attempts for 15 minutes
   - Password expiration: optional, configurable (default: disabled)
   - Password history: prevent reuse of last 12 passwords

2. API TOKEN AUTHENTICATION:
   - Token generation: cryptographically secure random 64-character strings
   - Named tokens: user-defined names for identification and management
   - Scoped permissions: granular access control per token
   - Expiration options: never, 7 days, 1 month, 1 year, 2 years, custom date
   - Single-view security: tokens displayed only once upon creation
   - Token rotation: generate new token while maintaining same permissions
   - Usage tracking: last used timestamp and access frequency

3. PASSKEY/WEBAUTHN AUTHENTICATION (optional):
   - WebAuthn Level 2 compliance
   - Platform authenticator support (TouchID, FaceID, Windows Hello)
   - Cross-platform authenticator support (YubiKey, etc.)
   - Multiple passkeys per user account
   - Passkey naming and management interface
   - Backup authentication methods required

TOKEN SCOPE DEFINITIONS:
- global: Complete API access equivalent to user session
- registry:read: Read access to specific registry
- registry:write: Write access to specific registry (includes read)
- registry:admin: Administrative access to specific registry
- org:read: Read access to organization and its registries
- org:write: Write access to organization registries
- org:admin: Administrative access to organization
- user:profile: Access to modify user profile only
- api:readonly: Read-only access to all permitted resources

JWT CONFIGURATION:
- Signing algorithm: RS256 (RSA with SHA-256)
- Key rotation: automatic every 90 days with 7-day overlap
- Token expiration: 24 hours for web sessions, configurable for API tokens
- Refresh token support: 30-day expiration with sliding window
- Issuer claim: configurable (default: casreg.local)
- Audience claim: configurable (default: casreg-api)

=============================================================================
ACCESS CONTROL AND VISIBILITY SYSTEM
=============================================================================

PUBLIC ACCESS (GUEST USERS):
- Registry discovery: browse public registries and repositories
- Image pulling: download from public repositories
- Search functionality: search across public registries and repositories
- Documentation access: view public documentation and API specifications
- Rate limiting: 100 requests per hour per IP address
- No account creation required for read-only operations

REGISTRY VISIBILITY LEVELS:
- Public: Visible to all users including guests, pullable by anyone
- Organization: Visible to organization members only
- Private: Visible only to registry owner and explicitly granted users
- Internal: Visible to all authenticated users but not guests

REPOSITORY-LEVEL PERMISSIONS:
- Inherit from registry: use registry visibility settings
- Override to public: make individual repository public in private registry
- Override to private: make individual repository private in public registry
- Custom access lists: grant specific users or organizations access

ORGANIZATION VISIBILITY:
- Public: Organization and public registries visible to all
- Internal: Organization visible to authenticated users only
- Private: Organization visible to members only
- Hidden: Organization not discoverable through search or browsing

=============================================================================
REGISTRY AND REPOSITORY MANAGEMENT
=============================================================================

DOCKER REGISTRY V2 API COMPLIANCE:
- Complete implementation of Docker Registry HTTP API V2
- OCI Distribution Specification compliance
- Manifest schema version 2 support
- Multi-architecture manifest support
- Layer deduplication across repositories
- Streaming upload and download with resume capability
- Chunked upload support for large images
- Manifest validation and integrity checking

REGISTRY CONFIGURATION OPTIONS:
- Maximum repositories per registry: configurable (default: unlimited)
- Storage quota per registry: configurable (default: unlimited)
- Retention policies: time-based and count-based cleanup
- Vulnerability scanning: enable/disable per registry
- Signature verification: require/optional/disabled per registry
- Anonymous access: allow/deny anonymous pulls
- Push restrictions: limit push operations to specific users/organizations

REPOSITORY FEATURES:
- Automatic tag creation on push
- Tag immutability options: prevent tag overwriting
- Lifecycle management: automatic cleanup of old tags
- Usage statistics: pull/push counts, storage utilization
- Access logs: detailed audit trail of all operations
- Webhook integration: configurable event notifications
- README support: markdown documentation per repository

TAG MANAGEMENT:
- Semantic version sorting and display
- Tag protection: prevent deletion of specified tags
- Tag aliases: create alternative names for existing tags
- Batch operations: bulk tag deletion and management
- Tag metadata: creation date, last pulled, size information
- Vulnerability scan results integration
- Signature verification status display

REGISTRY CLEANUP OPERATIONS:
- Orphaned layer detection and removal
- Unused manifest cleanup
- Configurable retention policies by tag count or age
- Manual cleanup triggers via UI and API
- Cleanup operation logging and reporting
- Storage reclamation verification

=============================================================================
SECURITY SCANNING INTEGRATION
=============================================================================

TRIVY SCANNER INTEGRATION (EMBEDDED):
- Trivy binary embedded in main application binary
- No external dependencies or network requirements for basic operation
- Vulnerability database bundled with application
- Automatic database updates when available
- Support for disconnected/air-gapped environments

SCANNING CONFIGURATION:
- Automatic scanning: scan all pushed images by default
- Manual scanning: trigger scans via UI, API, or CLI
- Scheduled scanning: periodic rescans of existing images
- Scan on push: immediate scanning of new image uploads
- Scan exclusions: skip scanning for specific repositories or tags
- Vulnerability severity filtering: configure minimum severity for reporting

VULNERABILITY REPORTING:
- Severity levels: Critical, High, Medium, Low, Unknown
- CVE database integration: detailed vulnerability information
- CVSS score reporting: quantitative risk assessment
- Package-level vulnerability mapping: identify affected components
- Fix recommendations: suggested remediation actions when available
- Historical scanning: track vulnerability changes over time

SCAN RESULT STORAGE:
- Detailed scan results stored in database
- JSON format vulnerability reports
- Scan result comparison between image versions
- Export capabilities: JSON, CSV, PDF reporting formats
- API access to scan results for automation integration
- Webhook notifications for new vulnerabilities

SCANNING PERFORMANCE:
- Parallel scanning: multiple concurrent scan jobs
- Scan queue management: prioritize critical images
- Resource limiting: configurable CPU and memory usage
- Scan result caching: avoid duplicate scanning of identical layers
- Progress tracking: real-time scan status updates

=============================================================================
SIGNATURE VERIFICATION SYSTEM
=============================================================================

COSIGN INTEGRATION (EMBEDDED):
- Cosign binary embedded in main application binary
- Support for keyless signing with Fulcio/Rekor
- Traditional key-based signing support
- Signature verification on image pull operations
- Batch signature verification for existing images

SIGNATURE POLICIES:
- Require signatures: enforce signature verification for all images
- Optional signatures: verify when present but allow unsigned images
- Signature enforcement by registry/repository: granular policy control
- Trusted signer configuration: specify allowed signing keys/certificates
- Policy exceptions: allow specific images to bypass signature requirements

SIGNATURE VERIFICATION WORKFLOW:
- Real-time verification: check signatures during image pull
- Batch verification: verify signatures for existing images
- Signature validation: cryptographic verification of signature integrity
- Certificate chain validation: verify signer certificate authority
- Transparency log verification: check Rekor entries for keyless signatures

SIGNATURE STORAGE AND REPORTING:
- Signature metadata storage in database
- Verification status tracking per image tag
- Signature verification audit logs
- Failed verification alerting and reporting
- Integration with vulnerability scanning for comprehensive security assessment

=============================================================================
QUOTA AND RESOURCE MANAGEMENT
=============================================================================

QUOTA SYSTEM ARCHITECTURE:
- Multi-level quota enforcement: user, organization, registry, repository
- Real-time quota monitoring and enforcement
- Soft and hard quota limits with configurable behaviors
- Quota inheritance: organization quotas apply to member registries
- Quota pooling: shared quotas across organization resources

QUOTA CONFIGURATION OPTIONS:
- Storage quotas: size-based limits (bytes, KB, MB, GB, TB, PB)
- Count quotas: limits on repositories, tags, images per entity
- Bandwidth quotas: upload/download rate limiting
- Request quotas: API call rate limiting per user/token
- Time-based quotas: daily/weekly/monthly limits

QUOTA ENFORCEMENT BEHAVIORS:
- Hard limits: block operations that would exceed quota
- Soft limits: warn users approaching quota limits
- Grace periods: temporary quota overages with automatic cleanup
- Quota notifications: email/webhook alerts for quota usage
- Automatic cleanup: remove oldest content when quota exceeded

QUOTA MONITORING AND REPORTING:
- Real-time quota usage display in UI
- Quota usage history and trending
- Quota efficiency reporting: utilization statistics
- Automated quota adjustment recommendations
- Integration with alerting systems for proactive management

DEFAULT QUOTA SETTINGS:
- User quota: unlimited (configurable globally)
- Organization quota: unlimited (configurable globally)
- Registry quota: unlimited (configurable per registry)
- Repository quota: unlimited (configurable per repository)
- API rate limits: 1000 requests per hour per authenticated user
- Anonymous rate limits: 100 requests per hour per IP address

=============================================================================
REAL-TIME NOTIFICATION SYSTEM
=============================================================================

REDIS/VALKEY INTEGRATION (OPTIONAL):
- Redis 6.0+ or Valkey compatibility
- Pub/Sub messaging for real-time updates
- Connection pooling and failover support
- SSL/TLS encryption for Redis connections
- Redis Cluster support for high availability
- Graceful degradation when Redis unavailable

NOTIFICATION CHANNELS:
- WebSocket connections: real-time UI updates
- Email notifications: SMTP-based messaging
- Webhook notifications: HTTP callbacks for external integration
- In-application notifications: persistent message queue
- System notifications: OS-level notifications for desktop applications

NOTIFICATION TYPES:
- Security alerts: vulnerability discoveries, failed authentication attempts
- System events: registry creation, user registration, quota warnings
- Operational events: scanning completion, signature verification status
- Administrative events: configuration changes, user management actions
- Support events: new tickets, ticket updates, system announcements

NOTIFICATION CONFIGURATION:
- Per-user notification preferences: enable/disable notification types
- Delivery method selection: email, in-app, webhook combinations
- Notification frequency: immediate, hourly digest, daily digest
- Notification filtering: severity-based filtering and custom rules
- Notification templates: customizable message formatting

NOTIFICATION DEDUPLICATION:
- Identical notification merging within time windows
- Escalation rules: increase notification frequency for critical events
- Notification batching: group related notifications for efficiency
- Delivery confirmation: track successful notification delivery
- Retry mechanisms: automatic retry for failed notification delivery

=============================================================================
SMTP AND MESSAGING SYSTEM
=============================================================================

SMTP CONFIGURATION REQUIREMENTS:
- SMTP server host: CASREG_SMTP_HOST (required when SMTP enabled)
- SMTP server port: CASREG_SMTP_PORT (default: 587)
- Authentication: username/password (optional for open relays)
- Encryption: TLS, STARTTLS, or unencrypted (configurable)
- From address: CASREG_SMTP_FROM (required, used for all outgoing mail)
- Subject prefix: CASREG_SMTP_SUBJECT_PREFIX (default: "[casreg]")

SMTP VALIDATION AND TESTING:
- Connection testing: verify SMTP server connectivity before saving configuration
- Authentication testing: validate credentials with actual login attempt
- Test message sending: send test email to verify complete configuration
- Delivery confirmation: track successful email delivery when possible
- Error handling: detailed error messages for SMTP configuration issues

EMAIL MESSAGE TYPES:
- Welcome emails: new user registration confirmation
- Password reset: secure password reset links with expiration
- Security alerts: suspicious activity notifications
- Quota warnings: approaching and exceeded quota notifications
- Vulnerability alerts: new vulnerability discoveries in user images
- System announcements: maintenance windows and feature updates

FALLBACK MESSAGING SYSTEM:
- Built-in message queue: persistent storage for undelivered messages
- In-application inbox: web-based message viewing for all users
- Message categories: system, security, administrative, promotional
- Message marking: read/unread status tracking
- Message archiving: automatic cleanup of old messages
- Export capabilities: message export for compliance and auditing

EMAIL TEMPLATE SYSTEM:
- HTML and text templates: dual-format email support
- Template customization: admin-configurable email templates
- Variable substitution: dynamic content insertion (user names, URLs, etc.)
- Localization support: multi-language email templates
- Brand customization: configurable logos, colors, and styling

=============================================================================
ADMIN PANEL SYSTEM
=============================================================================

ADMIN PANEL ACCESS CONTROL:
- Restricted to users with admin role only
- Session management: automatic timeout after inactivity
- Audit logging: all administrative actions logged with timestamps
- Two-factor authentication: optional additional security layer
- IP restrictions: limit admin access to specific IP addresses/ranges

ADMIN PANEL SECTIONS:

1. USER MANAGEMENT:
   - User listing: paginated table with search and filtering
   - User creation: create new users with specified roles
   - User modification: edit profiles, roles, permissions
   - User suspension: temporary account disabling
   - Password reset: administrative password reset capability
   - User statistics: registration trends, activity metrics

2. ORGANIZATION MANAGEMENT:
   - Organization listing: all organizations with ownership details
   - Organization creation: admin-created organizations
   - Membership management: add/remove members, modify roles
   - Organization settings: visibility, permissions, quotas
   - Organization statistics: usage metrics, member activity

3. REGISTRY MANAGEMENT:
   - Global registry overview: all registries across all users/organizations
   - Registry statistics: storage usage, pull/push metrics
   - Registry configuration: modify settings across multiple registries
   - Cleanup operations: bulk cleanup and maintenance tasks
   - Registry monitoring: performance and health metrics

4. TOKEN MANAGEMENT:
   - Global token overview: all active tokens across all users
   - Token analytics: usage patterns, security metrics
   - Token revocation: emergency token disabling
   - Token policies: global token configuration and restrictions
   - Token audit: detailed token usage logging

5. STORAGE MANAGEMENT:
   - Storage backend configuration: modify active storage settings
   - Storage migration tools: move data between storage backends
   - Storage monitoring: usage statistics, performance metrics
   - Cleanup utilities: orphaned data removal, optimization tools
   - Backup management: storage backup and restore operations

6. SMTP CONFIGURATION:
   - SMTP settings management: modify email server configuration
   - Email template management: customize system email templates
   - Email testing tools: send test emails and verify delivery
   - Email queue management: view pending and failed email deliveries
   - Email statistics: delivery rates, bounce tracking

7. DATABASE MIGRATION:
   - Migration wizard: step-by-step database migration process
   - Connection testing: validate new database connections
   - Migration preview: show planned migration operations
   - Progress monitoring: real-time migration status tracking
   - Rollback capabilities: undo failed or problematic migrations

8. TRUSTED IP MANAGEMENT:
   - IP address configuration: add/remove trusted IP ranges
   - Automatic detection: identify common proxy and CDN IP ranges
   - Validation tools: verify IP address and CIDR notation formats
   - Access logs: monitor connections from trusted IP addresses
   - Security policies: configure behavior for untrusted IP addresses

9. REAL-TIME NOTIFICATIONS:
   - Global notification settings: configure system-wide notification behavior
   - Message broadcasting: send announcements to all users
   - Notification analytics: delivery rates, user engagement metrics
   - Template management: customize notification message templates
   - Integration testing: verify webhook and email notification delivery

10. SUPPORT AND TICKETING:
    - Ticket queue management: view and respond to all support tickets
    - User communication: direct messaging with users
    - Knowledge base management: create and update documentation
    - FAQ management: maintain frequently asked questions
    - Support analytics: ticket resolution times, user satisfaction metrics

11. SYSTEM LOGS:
    - Comprehensive log viewing: access all system logs from web interface
    - Log filtering: search and filter logs by level, component, user
    - Log export: download logs for external analysis
    - Log retention: configure automatic log cleanup policies
    - Real-time monitoring: live log streaming for troubleshooting

12. DOCUMENTATION AND SWAGGER:
    - Documentation management: update built-in documentation
    - API documentation: manage Swagger specifications
    - Help system: configure context-sensitive help
    - User guides: maintain user and administrator guides
    - Video tutorials: embed and manage instructional videos

=============================================================================
COMMAND-LINE INTERFACE SPECIFICATION
=============================================================================

CLI ARCHITECTURE:
- Bubbletea framework: interactive terminal user interface
- Lipgloss styling: consistent theming across all CLI components
- XDG compliance: configuration stored in ~/.config/casreg/
- Shell completion: bash, zsh, fish completion support
- Cross-platform compatibility: Windows, macOS, Linux support

CLI INSTALLATION AND CONFIGURATION:
- Single binary: same binary serves as server and CLI client
- Configuration detection: automatic server discovery on localhost
- Remote server support: connect to remote casreg instances
- Authentication: token-based authentication for remote connections
- Configuration wizard: interactive setup for first-time users

CLI COMMAND STRUCTURE:
casreg [global-options] <command> <subcommand> [arguments] [flags]

GLOBAL OPTIONS:
- --config, -c: specify configuration file path
- --server, -s: specify server URL for remote operations
- --token, -t: specify authentication token
- --format, -f: output format (json, yaml, table, raw)
- --quiet, -q: suppress non-essential output
- --verbose, -v: increase output verbosity
- --help, -h: display help information
- --version: display version information

CLI THEMES:
- dracula (default): dark theme with purple/pink accents
- dark: high-contrast dark theme
- light: clean light theme for bright environments
- Theme selection via: casreg config set theme <theme-name>

ADMIN COMMANDS (require sudo/root privileges):
casreg admin user create <username> <email> [--password] [--role]
casreg admin user list [--format table|json] [--filter active|inactive]
casreg admin user modify <username> [--role] [--active] [--quota]
casreg admin user delete <username> [--force]
casreg admin user suspend <username> [--reason] [--duration]

casreg admin org create <name> [--display-name] [--description] [--owner]
casreg admin org list [--format table|json] [--filter public|private]
casreg admin org modify <name> [--display-name] [--description] [--visibility]
casreg admin org delete <name> [--force]

casreg admin registry list [--format table|json] [--owner user|org]
casreg admin registry stats [--registry] [--detailed]
casreg admin registry cleanup [--registry] [--dry-run] [--force]

casreg admin config get [<key>]
casreg admin config set <key> <value>
casreg admin config list [--format table|json]
casreg admin config reset [<key>] [--force]

casreg admin database migrate [--from] [--to] [--dry-run]
casreg admin database backup [--output] [--compress]
casreg admin database restore <backup-file> [--force]

casreg admin logs view [--level] [--component] [--follow] [--lines]
casreg admin logs export [--output] [--format] [--compress]

USER COMMANDS:
casreg user profile [--format json|yaml]
casreg user profile update [--first-name] [--last-name] [--email]
casreg user password change [--current] [--new]
casreg user settings get [<key>]
casreg user settings set <key> <value>

casreg user token create <name> [--scope] [--expires]
casreg user token list [--format table|json] [--show-expired]
casreg user token rotate <name|id>
casreg user token delete <name|id> [--force]

ORGANIZATION COMMANDS:
casreg org create <name> [--display-name] [--description] [--public]
casreg org list [--format table|json] [--owned] [--member]
casreg org view <name> [--format json|yaml]
casreg org update <name> [--display-name] [--description] [--visibility]
casreg org delete <name> [--force]

casreg org member add <org-name> <username> [--role]
casreg org member list <org-name> [--format table|json]
casreg org member remove <org-name> <username> [--force]
casreg org member role <org-name> <username> <role>

casreg org invite <org-name> <email> [--role] [--message]
casreg org invite list <org-name> [--format table|json] [--pending]
casreg org invite accept <token>
casreg org invite revoke <org-name> <email>

REGISTRY COMMANDS:
casreg registry create <name> [--display-name] [--description] [--org] [--public]
casreg registry list [--format table|json] [--owned] [--accessible]
casreg registry view <name> [--format json|yaml] [--stats]
casreg registry update <name> [--display-name] [--description] [--visibility]
casreg registry delete <name> [--force]

casreg registry repo create <registry-name> <repo-name> [--description]
casreg registry repo list <registry-name> [--format table|json]
casreg registry repo view <registry-name> <repo-name> [--format json|yaml]
casreg registry repo delete <registry-name> <repo-name> [--force]

casreg registry tag list <registry-name> <repo-name> [--format table|json]
casreg registry tag view <registry-name> <repo-name> <tag> [--format json|yaml]
casreg registry tag delete <registry-name> <repo-name> <tag> [--force]

casreg registry scan <registry-name> [<repo-name>] [--force] [--wait]
casreg registry scan-results <registry-name> <repo-name> <tag> [--format json|yaml]

SUPPORT COMMANDS:
casreg support ticket create [--title] [--description] [--priority]
casreg support ticket list [--format table|json] [--status open|closed]
casreg support ticket view <ticket-id> [--format json|yaml]
casreg support ticket update <ticket-id> [--status] [--priority] [--add-comment]
casreg support ticket close <ticket-id> [--comment]

casreg support docs list [--format table|json] [--category]
casreg support docs view <doc-name> [--format raw|formatted]
casreg support docs search <query> [--limit]

CLI INTERACTIVE MODE:
casreg interactive: Launch full TUI interface with:
- Dashboard: system overview, statistics, notifications
- Registry browser: navigate registries and repositories
- User management: profile, tokens, organizations
- Support interface: tickets, documentation, help
- Admin panel: full administrative interface (admin users only)

CLI OUTPUT FORMATTING:
- Table format: human-readable tabular output with headers
- JSON format: machine-readable JSON output
- YAML format: human and machine-readable YAML output
- Raw format: unformatted output for scripting

CLI ERROR HANDLING:
- Detailed error messages with suggested resolutions
- Exit codes: 0 for success, non-zero for various error conditions
- Error logging: automatic error reporting to server logs
- Offline mode: graceful handling of server connectivity issues

=============================================================================
WEB USER INTERFACE SPECIFICATION
=============================================================================

UI FRAMEWORK AND ARCHITECTURE:
- Svelte 4.x (preferred) or Vue 3.x with Composition API
- Vite build system for fast development and optimized production builds
- TypeScript for type safety and improved developer experience
- Tailwind CSS for utility-first styling and consistent design
- Component-based architecture with reusable UI elements

UI LAYOUT AND NAVIGATION:
- Single-page application with client-side routing
- Responsive design: mobile-first approach with desktop enhancements
- Sidebar navigation: collapsible menu with hierarchical organization
- Top navigation bar: user profile, notifications, search, theme selector
- Breadcrumb navigation: contextual navigation path display
- Global search: registry, repository, and content search functionality

AUTHENTICATION UI:
- Login page: username/password and passkey authentication options
- Registration page: new user signup with email verification
- Password reset: secure password reset workflow with email confirmation
- Two-factor authentication: TOTP and passkey setup and management
- Session management: automatic session renewal and security warnings

DASHBOARD AND OVERVIEW:
- Personal dashboard: user statistics, recent activity, notifications
- System overview: registry statistics, storage usage, system health
- Quick actions: shortcuts to common operations
- Activity timeline: chronological view of user and system activities
- Notification center: in-app notifications with read/unread status

REGISTRY MANAGEMENT UI:
- Registry listing: searchable and filterable registry browser
- Registry creation wizard: step-by-step registry setup process
- Registry settings: comprehensive configuration management interface
- Repository browser: hierarchical view of repositories and tags
- Tag viewer: detailed tag information with vulnerability and signature status

ORGANIZATION MANAGEMENT UI:
- Organization dashboard: membership, registries, activity overview
- Member management: invite, remove, and role management interface
- Organization settings: visibility, permissions, and quota configuration
- Invitation management: pending invitations and acceptance workflow
- Organization statistics: usage metrics and activity reports

USER PROFILE AND SETTINGS:
- Profile management: personal information, avatar, preferences
- Security settings: password change, passkey management, session history
- Notification preferences: granular notification configuration
- API token management: create, view, rotate, and delete tokens
- Theme selection: visual theme customization

ADMIN INTERFACE:
- Admin dashboard: system-wide statistics and health monitoring
- User management: comprehensive user administration interface
- System configuration: global settings and policy management
- Database management: migration tools and maintenance operations
- Audit logs: searchable system activity and security logs

SUPPORT SYSTEM UI:
- Ticket creation: structured support request submission
- Ticket management: view, update, and track support requests
- Knowledge base: searchable documentation and FAQ system
- Live chat: real-time support communication (when available)
- Feedback system: user satisfaction and feature request submission

THEME SYSTEM:
DRACULA THEME (default):
- Primary colors: #282a36 (background), #44475a (selection), #f8f8f2 (foreground)
- Accent colors: #bd93f9 (purple), #ff79c6 (pink), #50fa7b (green)
- Syntax highlighting: #8be9fd (cyan), #ffb86c (orange), #ff5555 (red)
- Interactive elements: hover states, focus indicators, transitions

DARK THEME:
- High contrast design: #121212 (background), #1e1e1e (surface), #ffffff (foreground)
- Accent colors: #bb86fc (purple), #03dac6 (teal), #cf6679 (error)
- Emphasis: #ffffff (high), #b3b3b3 (medium), #808080 (disabled)

LIGHT THEME:
- Clean design: #ffffff (background), #f5f5f5 (surface), #212121 (foreground)
- Accent colors: #6200ea (purple), #018786 (teal), #b00020 (error)
- Material Design inspired: shadows, elevation, clean typography

UI ACCESSIBILITY:
- WCAG 2.1 AA compliance: color contrast, keyboard navigation, screen reader support
- Semantic HTML: proper heading hierarchy, landmark regions, form labels
- Keyboard shortcuts: comprehensive keyboard navigation support
- Focus management: visible focus indicators, logical tab order
- Alternative text: images and icons with descriptive alt text

UI PERFORMANCE:
- Code splitting: lazy loading of route components
- Image optimization: responsive images with WebP support
- Bundle optimization: tree shaking and minification
- Progressive enhancement: core functionality without JavaScript
- Offline support: service worker for cached resource access

=============================================================================
API SPECIFICATION AND ENDPOINTS
=============================================================================

API ARCHITECTURE:
- RESTful design principles with resource-based URLs
- JSON request and response payloads exclusively
- HTTP status codes for standardized response indication
- Comprehensive error handling with detailed error messages
- API versioning: all endpoints prefixed with /v1/
- Rate limiting: configurable per-user and per-IP rate limits
- CORS support: configurable cross-origin resource sharing

API AUTHENTICATION:
- Bearer token authentication: Authorization: Bearer <token>
- JWT token support: short-lived tokens with refresh capability
- API key authentication: long-lived tokens for automation
- Scope-based authorization: granular permission control
- Token introspection: validate and inspect token permissions

ERROR RESPONSE FORMAT:
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error description",
    "details": {
      "field": "specific field that caused the error",
      "value": "invalid value that was provided"
    },
    "documentation_url": "https://docs.casreg.io/errors/ERROR_CODE"
  }
}

AUTHENTICATION ENDPOINTS:

POST /v1/auth/login
- Request: {"username": "string", "password": "string"}
- Response: {"token": "jwt_token", "refresh_token": "refresh_token", "expires_in": 3600}
- Status: 200 (success), 401 (invalid credentials), 429 (rate limited)

POST /v1/auth/refresh
- Request: {"refresh_token": "refresh_token"}
- Response: {"token": "new_jwt_token", "expires_in": 3600}
- Status: 200 (success), 401 (invalid token), 403 (token expired)

POST /v1/auth/logout
- Request: {} (empty body, token in header)
- Response: {"message": "Logged out successfully"}
- Status: 200 (success), 401 (not authenticated)

POST /v1/auth/register
- Request: {"username": "string", "email": "string", "password": "string", "first_name": "string", "last_name": "string"}
- Response: {"user": {...}, "verification_required": true}
- Status: 201 (created), 400 (validation error), 409 (username/email exists)

USER MANAGEMENT ENDPOINTS:

GET /v1/users/me
- Response: Complete user profile with preferences and statistics
- Status: 200 (success), 401 (not authenticated)

PUT /v1/users/me
- Request: {"first_name": "string", "last_name": "string", "email": "string", "theme": "string"}
- Response: Updated user profile
- Status: 200 (success), 400 (validation error), 401 (not authenticated)

GET /v1/users/{user_id}
- Response: Public user profile (limited information)
- Status: 200 (success), 404 (user not found), 403 (access denied)

POST /v1/users/me/password
- Request: {"current_password": "string", "new_password": "string"}
- Response: {"message": "Password updated successfully"}
- Status: 200 (success), 400 (validation error), 401 (invalid current password)

TOKEN MANAGEMENT ENDPOINTS:

GET /v1/tokens
- Response: List of user's tokens with metadata (excluding token values)
- Status: 200 (success), 401 (not authenticated)

POST /v1/tokens
- Request: {"name": "string", "scopes": ["scope1", "scope2"], "expires_at": "ISO8601_date"}
- Response: {"token": "generated_token", "id": "token_id", "name": "string", "scopes": [...]}
- Status: 201 (created), 400 (validation error), 401 (not authenticated)

DELETE /v1/tokens/{token_id}
- Response: {"message": "Token deleted successfully"}
- Status: 200 (success), 404 (token not found), 401 (not authenticated)

POST /v1/tokens/{token_id}/rotate
- Response: {"token": "new_generated_token", "id": "token_id"}
- Status: 200 (success), 404 (token not found), 401 (not authenticated)

ORGANIZATION ENDPOINTS:

GET /v1/organizations
- Query parameters: ?limit=20&offset=0&visibility=public&member=true
- Response: Paginated list of organizations
- Status: 200 (success)

POST /v1/organizations
- Request: {"name": "string", "display_name": "string", "description": "string", "is_public": boolean}
- Response: Created organization object
- Status: 201 (created), 400 (validation error), 409 (name exists), 401 (not authenticated)

GET /v1/organizations/{org_name}
- Response: Complete organization details with member information
- Status: 200 (success), 404 (not found), 403 (access denied)

PUT /v1/organizations/{org_name}
- Request: {"display_name": "string", "description": "string", "is_public": boolean}
- Response: Updated organization object
- Status: 200 (success), 400 (validation error), 404 (not found), 403 (access denied)

DELETE /v1/organizations/{org_name}
- Response: {"message": "Organization deleted successfully"}
- Status: 200 (success), 404 (not found), 403 (access denied), 409 (has repositories)

POST /v1/organizations/{org_name}/members
- Request: {"username": "string", "role": "member|admin"}
- Response: {"message": "Member added successfully"}
- Status: 201 (created), 400 (validation error), 404 (org/user not found), 403 (access denied)

GET /v1/organizations/{org_name}/members
- Response: List of organization members with roles and join dates
- Status: 200 (success), 404 (not found), 403 (access denied)

DELETE /v1/organizations/{org_name}/members/{username}
- Response: {"message": "Member removed successfully"}
- Status: 200 (success), 404 (not found), 403 (access denied)

REGISTRY ENDPOINTS:

GET /v1/registries
- Query parameters: ?limit=20&offset=0&owner_type=user&owner_id=123&visibility=public
- Response: Paginated list of registries with metadata
- Status: 200 (success)

POST /v1/registries
- Request: {"name": "string", "display_name": "string", "description": "string", "owner_type": "user|organization", "owner_id": number, "is_public": boolean}
- Response: Created registry object
- Status: 201 (created), 400 (validation error), 409 (name exists), 401 (not authenticated)

GET /v1/registries/{registry_name}
- Response: Complete registry details with repository list
- Status: 200 (success), 404 (not found), 403 (access denied)

PUT /v1/registries/{registry_name}
- Request: {"display_name": "string", "description": "string", "is_public": boolean, "enable_scanning": boolean}
- Response: Updated registry object
- Status: 200 (success), 400 (validation error), 404 (not found), 403 (access denied)

DELETE /v1/registries/{registry_name}
- Response: {"message": "Registry deleted successfully"}
- Status: 200 (success), 404 (not found), 403 (access denied), 409 (has repositories)

REPOSITORY ENDPOINTS:

GET /v1/registries/{registry_name}/repositories
- Query parameters: ?limit=20&offset=0&sort=name|created_at|last_pushed
- Response: Paginated list of repositories with metadata
- Status: 200 (success), 404 (registry not found), 403 (access denied)

POST /v1/registries/{registry_name}/repositories
- Request: {"name": "string", "description": "string", "is_public": boolean}
- Response: Created repository object
- Status: 201 (created), 400 (validation error), 404 (registry not found), 403 (access denied)

GET /v1/registries/{registry_name}/repositories/{repo_name}
- Response: Complete repository details with tag list and statistics
- Status: 200 (success), 404 (not found), 403 (access denied)

PUT /v1/registries/{registry_name}/repositories/{repo_name}
- Request: {"description": "string", "is_public": boolean, "enable_scanning": boolean}
- Response: Updated repository object
- Status: 200 (success), 400 (validation error), 404 (not found), 403 (access denied)

DELETE /v1/registries/{registry_name}/repositories/{repo_name}
- Response: {"message": "Repository deleted successfully"}
- Status: 200 (success), 404 (not found), 403 (access denied)

TAG ENDPOINTS:

GET /v1/registries/{registry_name}/repositories/{repo_name}/tags
- Query parameters: ?limit=20&offset=0&sort=name|created_at|size
- Response: Paginated list of tags with metadata and scan results
- Status: 200 (success), 404 (not found), 403 (access denied)

GET /v1/registries/{registry_name}/repositories/{repo_name}/tags/{tag_name}
- Response: Complete tag details with layers, scan results, and signature status
- Status: 200 (success), 404 (not found), 403 (access denied)

DELETE /v1/registries/{registry_name}/repositories/{repo_name}/tags/{tag_name}
- Response: {"message": "Tag deleted successfully"}
- Status: 200 (success), 404 (not found), 403 (access denied)

POST /v1/registries/{registry_name}/repositories/{repo_name}/tags/{tag_name}/scan
- Response: {"message": "Scan initiated", "scan_id": "uuid"}
- Status: 202 (accepted), 404 (not found), 403 (access denied), 409 (scan in progress)

GET /v1/registries/{registry_name}/repositories/{repo_name}/tags/{tag_name}/scan-results
- Response: Detailed vulnerability scan results and signature verification status
- Status: 200 (success), 404 (not found), 403 (access denied)

ADMIN ENDPOINTS:

GET /v1/admin/users
- Query parameters: ?limit=20&offset=0&role=admin|user&active=true|false&sort=username|created_at
- Response: Paginated list of all users with detailed information
- Status: 200 (success), 403 (not admin)

POST /v1/admin/users
- Request: {"username": "string", "email": "string", "password": "string", "role": "admin|user", "is_active": boolean}
- Response: Created user object
- Status: 201 (created), 400 (validation error), 409 (username/email exists), 403 (not admin)

PUT /v1/admin/users/{user_id}
- Request: {"role": "admin|user", "is_active": boolean, "quota_limit": "string"}
- Response: Updated user object
- Status: 200 (success), 400 (validation error), 404 (not found), 403 (not admin)

DELETE /v1/admin/users/{user_id}
- Response: {"message": "User deleted successfully"}
- Status: 200 (success), 404 (not found), 403 (not admin), 409 (cannot delete last admin)

GET /v1/admin/organizations
- Query parameters: ?limit=20&offset=0&visibility=public|private&sort=name|created_at
- Response: Paginated list of all organizations with detailed information
- Status: 200 (success), 403 (not admin)

GET /v1/admin/registries
- Query parameters: ?limit=20&offset=0&owner_type=user|organization&sort=name|created_at|size
- Response: Paginated list of all registries with detailed information
- Status: 200 (success), 403 (not admin)

GET /v1/admin/system/stats
- Response: Comprehensive system statistics and health metrics
- Status: 200 (success), 403 (not admin)

POST /v1/admin/system/cleanup
- Request: {"type": "orphaned_layers|unused_manifests|expired_tokens", "dry_run": boolean}
- Response: {"message": "Cleanup initiated", "job_id": "uuid", "estimated_items": number}
- Status: 202 (accepted), 403 (not admin)

SUPPORT ENDPOINTS:

GET /v1/support/tickets
- Query parameters: ?limit=20&offset=0&status=open|closed&priority=low|medium|high|critical
- Response: Paginated list of user's support tickets
- Status: 200 (success), 401 (not authenticated)

POST /v1/support/tickets
- Request: {"title": "string", "description": "string", "priority": "low|medium|high|critical", "category": "string"}
- Response: Created ticket object
- Status: 201 (created), 400 (validation error), 401 (not authenticated)

GET /v1/support/tickets/{ticket_id}
- Response: Complete ticket details with comments and attachments
- Status: 200 (success), 404 (not found), 403 (access denied), 401 (not authenticated)

POST /v1/support/tickets/{ticket_id}/comments
- Request: {"comment": "string", "attachments": ["file_id1", "file_id2"]}
- Response: Created comment object
- Status: 201 (created), 400 (validation error), 404 (ticket not found), 403 (access denied)

GET /v1/support/docs
- Query parameters: ?category=string&search=string&limit=20&offset=0
- Response: Paginated list of documentation articles
- Status: 200 (success)

GET /v1/support/docs/{doc_id}
- Response: Complete documentation article with content and metadata
- Status: 200 (success), 404 (not found)

DOCKER REGISTRY V2 API ENDPOINTS:

GET /v2/
- Response: {"name": "casreg", "version": "1.0.0"}
- Status: 200 (success)

GET /v2/{registry_name}/{repository_name}/manifests/{reference}
- Response: Docker manifest (schema version 2)
- Headers: Content-Type, Docker-Content-Digest
- Status: 200 (success), 404 (not found), 403 (access denied)

PUT /v2/{registry_name}/{repository_name}/manifests/{reference}
- Request: Docker manifest in JSON format
- Response: Empty body
- Headers: Location, Docker-Content-Digest
- Status: 201 (created), 400 (invalid manifest), 403 (access denied)

GET /v2/{registry_name}/{repository_name}/blobs/{digest}
- Response: Binary blob data
- Headers: Content-Type, Content-Length, Docker-Content-Digest
- Status: 200 (success), 404 (not found), 403 (access denied)

POST /v2/{registry_name}/{repository_name}/blobs/uploads/
- Response: Empty body
- Headers: Location (upload URL), Range, Docker-Upload-UUID
- Status: 202 (accepted), 403 (access denied)

PUT /v2/{registry_name}/{repository_name}/blobs/uploads/{uuid}
- Request: Binary blob data
- Query parameters: ?digest=sha256:hash
- Response: Empty body
- Headers: Location, Docker-Content-Digest
- Status: 201 (created), 400 (invalid digest), 403 (access denied)

GET /v2/{registry_name}/{repository_name}/tags/list
- Query parameters: ?n=100&last=tag_name
- Response: {"name": "repository_name", "tags": ["tag1", "tag2", ...]}
- Status: 200 (success), 404 (not found), 403 (access denied)

API RESPONSE PAGINATION:
- Standard pagination using limit and offset query parameters
- Response headers: X-Total-Count, X-Page-Count, X-Current-Page
- Link header with next, prev, first, last URLs
- Maximum limit: 100 items per page
- Default limit: 20 items per page

API RATE LIMITING:
- Rate limit headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
- Different limits for authenticated vs anonymous users
- Per-endpoint rate limiting for expensive operations
- Burst allowances for normal usage patterns
- Rate limit bypass for admin users on administrative operations

=============================================================================
SUPPORT AND TICKETING SYSTEM
=============================================================================

TICKETING SYSTEM ARCHITECTURE:
- Integrated support ticket management within the application
- Multi-channel support: web interface, email integration, API access
- Ticket lifecycle management: creation, assignment, escalation, resolution
- Knowledge base integration: automatic documentation suggestions
- User satisfaction tracking: post-resolution feedback collection

TICKET CREATION AND MANAGEMENT:
- Web-based ticket creation: structured forms with category selection
- Email-to-ticket conversion: automatic ticket creation from support emails
- API ticket creation: programmatic ticket creation for automation
- Ticket templates: predefined templates for common issue types
- File attachments: support for screenshots, logs, and configuration files

TICKET CATEGORIES AND PRIORITIES:
CATEGORIES:
- technical-issue: bugs, errors, performance problems
- feature-request: new feature suggestions and enhancements
- account-management: user account, authentication, permissions
- billing-inquiry: quota, usage, billing questions (if applicable)
- security-concern: security vulnerabilities, suspicious activity
- general-question: usage questions, how-to inquiries

PRIORITIES:
- critical: system down, security breach, data loss
- high: major functionality broken, significant user impact
- medium: minor functionality issues, moderate user impact
- low: cosmetic issues, feature requests, general questions

TICKET ASSIGNMENT AND ROUTING:
- Automatic assignment: route tickets based on category and keywords
- Manual assignment: admin users can reassign tickets
- Escalation rules: automatic escalation based on age and priority
- Load balancing: distribute tickets evenly among support staff
- Expertise routing: assign tickets based on staff expertise areas

TICKET COMMUNICATION:
- Internal comments: staff-only communication on tickets
- Public comments: user-visible responses and updates
- Email notifications: automatic notifications for ticket updates
- Status updates: automated status change notifications
- Resolution notifications: ticket closure and satisfaction surveys

KNOWLEDGE BASE SYSTEM:
- Hierarchical documentation structure: categories, subcategories, articles
- Full-text search: searchable across all documentation content
- Article versioning: track changes and maintain article history
- User contributions: allow users to suggest documentation improvements
- Usage analytics: track popular articles and search queries

KNOWLEDGE BASE CONTENT:
- Installation guides: platform-specific installation instructions
- Configuration reference: comprehensive configuration documentation
- API documentation: complete API reference with examples
- Troubleshooting guides: common problems and solutions
- Best practices: recommended configurations and usage patterns
- Video tutorials: embedded video content for complex procedures

SUPPORT ANALYTICS AND REPORTING:
- Ticket volume metrics: track ticket creation and resolution rates
- Response time analytics: measure first response and resolution times
- User satisfaction scores: collect and analyze feedback ratings
- Knowledge base analytics: track article views and search effectiveness
- Support staff performance: measure individual and team performance

ADMIN SUPPORT INTERFACE:
- Ticket queue management: prioritized list of open tickets
- Bulk operations: assign, close, or update multiple tickets
- Response templates: predefined responses for common inquiries
- User communication: direct messaging with users outside of tickets
- Support statistics: comprehensive reporting and analytics dashboard

USER SUPPORT INTERFACE:
- Ticket creation wizard: guided ticket creation process
- Ticket history: view all submitted tickets and their status
- Knowledge base browser: categorized documentation access
- Search functionality: search tickets and documentation
- Feedback system: rate support interactions and suggest improvements

AUTOMATED SUPPORT FEATURES:
- Auto-responses: immediate acknowledgment of ticket submission
- Keyword detection: automatic tagging and routing based on content
- Duplicate detection: identify and merge duplicate tickets
- Solution suggestions: recommend knowledge base articles based on ticket content
- Follow-up automation: scheduled check-ins for resolved tickets

=============================================================================
ISSUE TRACKING SYSTEM (REPOSITORY-LEVEL)
=============================================================================

REPOSITORY ISSUE SYSTEM:
- GitHub-style issue tracking for individual repositories
- Public and private issue visibility based on repository settings
- Issue templates: standardized formats for bug reports and feature requests
- Label system: categorize and organize issues with colored labels
- Milestone tracking: group issues by release or project milestones

ISSUE CREATION AND MANAGEMENT:
- Markdown editor: rich text editing with preview functionality
- Template selection: choose from predefined issue templates
- Automatic labeling: suggest labels based on issue content
- Assignment system: assign issues to repository collaborators
- Due date tracking: set and monitor issue deadlines

ISSUE CATEGORIES AND LABELS:
DEFAULT LABELS:
- bug: software defects and unexpected behavior
- enhancement: feature requests and improvements
- documentation: documentation updates and corrections
- question: usage questions and clarifications
- help-wanted: issues suitable for community contribution
- good-first-issue: beginner-friendly issues for new contributors

CUSTOM LABELS:
- User-defined labels with custom names and colors
- Label descriptions: provide context for label usage
- Label management: create, edit, and delete repository labels
- Bulk labeling: apply labels to multiple issues simultaneously

ISSUE WORKFLOW:
- Issue states: open, in-progress, resolved, closed
- State transitions: automatic and manual state changes
- Issue linking: reference other issues, pull requests, or commits
- Cross-referencing: automatic detection of issue references
- Resolution tracking: link issues to specific commits or releases

ISSUE NOTIFICATIONS:
- Real-time notifications: immediate updates for issue changes
- Email notifications: configurable email alerts for issue activity
- Webhook integration: HTTP callbacks for external issue tracking
- Subscription management: subscribe to specific issues or repositories
- Notification preferences: granular control over notification types

ISSUE SEARCH AND FILTERING:
- Full-text search: search issue titles, descriptions, and comments
- Advanced filtering: filter by state, labels, assignee, milestone
- Saved searches: bookmark frequently used search queries
- Bulk operations: perform actions on multiple filtered issues
- Export functionality: export issue data in various formats

ISSUE ANALYTICS:
- Issue metrics: track open/closed ratios, resolution times
- Label analytics: most used labels and categorization trends
- Activity tracking: issue creation and resolution patterns
- Contributor analytics: track issue participation and contributions
- Repository health: measure project activity and responsiveness

=============================================================================
CONFIGURATION SYSTEM
=============================================================================

CONFIGURATION HIERARCHY:
1. Default values: hard-coded sensible defaults
2. Environment variables: runtime configuration via environment
3. Configuration files: YAML/JSON configuration files
4. Database configuration: persistent configuration stored in database
5. Admin UI overrides: real-time configuration changes via web interface

ENVIRONMENT VARIABLE NAMING:
- Prefix: All environment variables prefixed with CASREG_
- Naming convention: UPPERCASE with underscores for separation
- Nested configuration: Double underscores for nested properties
- Boolean values: Accept yes/true/1/enable/on for true, no/false/0/disable/off for false

CONFIGURATION VALIDATION:
- Type validation: ensure configuration values match expected types
- Range validation: validate numeric values within acceptable ranges
- Format validation: validate email addresses, URLs, file paths
- Dependency validation: ensure related configuration values are consistent
- Real-time validation: validate configuration changes before applying

CONFIGURATION CHANGE MANAGEMENT:
- Change tracking: log all configuration changes with timestamps and users
- Rollback capability: revert configuration changes to previous states
- Change approval: require admin approval for critical configuration changes
- Impact analysis: assess the impact of configuration changes before applying
- Staged deployment: apply configuration changes gradually across system components

CONFIGURATION SECTIONS:

SERVER CONFIGURATION:
- CASREG_PORT: Server listening port (default: 8080)
- CASREG_HOST: Server binding address (default: 0.0.0.0)
- CASREG_BASE_URL: Public base URL for the application
- CASREG_DEBUG: Enable debug mode (default: false)
- CASREG_LOG_LEVEL: Logging level (default: info)
- CASREG_MAX_REQUEST_SIZE: Maximum request body size (default: 100MB)
- CASREG_REQUEST_TIMEOUT: Request timeout duration (default: 30s)

DATABASE CONFIGURATION:
- CASREG_DATABASE_TYPE: Database type (default: sqlite)
- CASREG_DATABASE_URL: Database connection string
- CASREG_DATABASE_HOST: Database server hostname
- CASREG_DATABASE_PORT: Database server port
- CASREG_DATABASE_NAME: Database name
- CASREG_DATABASE_USER: Database username
- CASREG_DATABASE_PASSWORD: Database password
- CASREG_DATABASE_SSL_MODE: SSL connection mode (default: disable)
- CASREG_DATABASE_POOL_SIZE: Connection pool size (default: 25)
- CASREG_DATABASE_MIGRATIONS: Enable automatic migrations (default: true)

STORAGE CONFIGURATION:
- CASREG_STORAGE_BACKEND: Storage backend type (default: local)
- CASREG_STORAGE_PATH: Local storage path (default: /var/lib/casreg/storage)
- CASREG_STORAGE_S3_ENDPOINT: S3 endpoint URL
- CASREG_STORAGE_S3_BUCKET: S3 bucket name
- CASREG_STORAGE_S3_REGION: S3 region (default: us-east-1)
- CASREG_STORAGE_S3_ACCESS_KEY: S3 access key
- CASREG_STORAGE_S3_SECRET_KEY: S3 secret key
- CASREG_STORAGE_S3_USE_SSL: Use SSL for S3 connections (default: true)
- CASREG_STORAGE_NFS_SERVER: NFS server address
- CASREG_STORAGE_NFS_PATH: NFS mount path

SECURITY CONFIGURATION:
- CASREG_JWT_SECRET: JWT signing secret (minimum 32 characters)
- CASREG_JWT_EXPIRATION: JWT token expiration (default: 24h)
- CASREG_BCRYPT_COST: Bcrypt hashing cost (default: 12)
- CASREG_TRUSTED_IPS: Comma-separated list of trusted IP addresses/ranges
- CASREG_RATE_LIMIT_ENABLED: Enable rate limiting (default: true)
- CASREG_RATE_LIMIT_REQUESTS: Requests per time window (default: 1000)
- CASREG_RATE_LIMIT_WINDOW: Rate limit time window (default: 1h)
- CASREG_PASSKEYS_ENABLED: Enable passkey authentication (default: true)

FEATURE TOGGLES:
- CASREG_SEARCH_ENABLED: Enable search functionality (default: true)
- CASREG_QUOTAS_ENABLED: Enable quota system (default: true)
- CASREG_NOTIFICATIONS_ENABLED: Enable notifications (default: true)
- CASREG_SCANNING_ENABLED: Enable vulnerability scanning (default: true)
- CASREG_SIGNATURE_VERIFICATION: Enable signature verification (default: true)
- CASREG_ALLOW_PUBLIC_REGISTRIES: Allow public registry creation (default: true)
- CASREG_ALLOW_GUEST_ACCESS: Allow guest access to public resources (default: true)
- CASREG_ALLOW_USER_REGISTRATION: Allow new user registration (default: true)

SMTP CONFIGURATION:
- CASREG_SMTP_ENABLED: Enable SMTP email sending (default: false)
- CASREG_SMTP_HOST: SMTP server hostname
- CASREG_SMTP_PORT: SMTP server port (default: 587)
- CASREG_SMTP_USERNAME: SMTP authentication username
- CASREG_SMTP_PASSWORD: SMTP authentication password
- CASREG_SMTP_FROM: From email address for outgoing emails
- CASREG_SMTP_USE_TLS: Use TLS encryption (default: true)
- CASREG_SMTP_USE_STARTTLS: Use STARTTLS encryption (default: false)
- CASREG_SMTP_SUBJECT_PREFIX: Email subject prefix (default: [casreg])

REDIS/VALKEY CONFIGURATION:
- CASREG_REDIS_ENABLED: Enable Redis/Valkey for real-time features (default: false)
- CASREG_REDIS_HOST: Redis server hostname (default: localhost)
- CASREG_REDIS_PORT: Redis server port (default: 6379)
- CASREG_REDIS_PASSWORD: Redis authentication password
- CASREG_REDIS_DATABASE: Redis database number (default: 0)
- CASREG_REDIS_SSL: Use SSL for Redis connections (default: false)
- CASREG_REDIS_POOL_SIZE: Redis connection pool size (default: 10)

QUOTA CONFIGURATION:
- CASREG_DEFAULT_USER_QUOTA: Default user storage quota (default: unlimited)
- CASREG_DEFAULT_ORG_QUOTA: Default organization storage quota (default: unlimited)
- CASREG_DEFAULT_REPO_QUOTA: Default repository storage quota (default: unlimited)
- CASREG_QUOTA_GRACE_PERIOD: Grace period for quota overages (default: 24h)
- CASREG_QUOTA_CLEANUP_INTERVAL: Quota cleanup interval (default: 1h)

THEME CONFIGURATION:
- CASREG_DEFAULT_THEME: Default UI theme (default: dracula)
- CASREG_THEME_CUSTOMIZATION_ENABLED: Allow theme customization (default: true)
- CASREG_CUSTOM_CSS_URL: URL for custom CSS stylesheet
- CASREG_CUSTOM_LOGO_URL: URL for custom logo image

=============================================================================
SCHEDULER AND BACKGROUND TASKS
=============================================================================

SCHEDULER ARCHITECTURE:
- Cron-like scheduler running every 1 minute
- Task queue with priority levels and retry mechanisms
- Configurable task execution limits and timeouts
- Task logging and error handling with detailed reporting
- Graceful shutdown handling for running tasks

TASK TYPES AND SCHEDULES:

CLEANUP TASKS (every 5 minutes):
- Orphaned layer cleanup: remove unreferenced storage layers
- Unused manifest cleanup: remove manifests without tags
- Temporary file cleanup: remove abandoned upload files
- Log rotation: archive and compress old log files
- Session cleanup: remove expired user sessions

TOKEN MANAGEMENT (every 15 minutes):
- Expired token purge: remove expired authentication tokens
- Token usage analytics: update token usage statistics
- JWT key rotation: rotate signing keys according to schedule
- Refresh token cleanup: remove expired refresh tokens

QUOTA MONITORING (every 30 minutes):
- Quota usage calculation: update storage usage for all entities
- Quota violation detection: identify quota overages
- Quota warning notifications: send approaching quota alerts
- Automatic cleanup: remove old content when quotas exceeded

NOTIFICATION PROCESSING (every 1 minute):
- Email queue processing: send queued email notifications
- Webhook delivery: deliver pending webhook notifications
- Notification retry: retry failed notification deliveries
- Notification analytics: update delivery statistics

AUDIT LOG MANAGEMENT (every 1 hour):
- Audit log rotation: archive old audit log entries
- Audit log compression: compress archived audit logs
- Audit log cleanup: remove audit logs older than retention period
- Audit analytics: generate audit usage reports

VULNERABILITY SCANNING (continuous):
- Scan queue processing: process pending vulnerability scans
- Scan result updates: update scan status and results
- Scan retry logic: retry failed scan operations
- Scan cleanup: remove old scan results and temporary files

SIGNATURE VERIFICATION (continuous):
- Signature verification queue: process pending verification requests
- Certificate validation: verify signing certificate chains
- Transparency log checking: verify entries in transparency logs
- Signature status updates: update verification status in database

STORAGE OPTIMIZATION (every 6 hours):
- Deduplication analysis: identify duplicate storage content
- Compression optimization: optimize storage compression
- Storage health checks: verify storage backend connectivity
- Storage migration: move content between storage backends

DATABASE MAINTENANCE (every 12 hours):
- Database optimization: run database maintenance operations
- Index maintenance: rebuild and optimize database indexes
- Statistics updates: update database query statistics
- Connection pool health: monitor database connection health

SYSTEM HEALTH MONITORING (every 5 minutes):
- Service health checks: verify all system components
- Performance metrics: collect system performance data
- Resource usage monitoring: track CPU, memory, disk usage
- Error rate monitoring: track error rates across components

TASK CONFIGURATION:
- Task scheduling: configurable task execution schedules
- Task priorities: high, medium, low priority task queues
- Task timeouts: configurable maximum execution times
- Task retries: automatic retry with exponential backoff
- Task concurrency: configurable parallel task execution limits

TASK MONITORING AND LOGGING:
- Task execution logs: detailed logging of all task execution
- Task performance metrics: track task execution times and success rates
- Task failure analysis: detailed error reporting and analysis
- Task queue monitoring: real-time monitoring of task queue status
- Task alerting: notifications for task failures and performance issues

=============================================================================
REVERSE PROXY AND TRUSTED IP HANDLING
=============================================================================

REVERSE PROXY SUPPORT:
- Complete support for all standard reverse proxy headers
- Automatic client IP detection through proxy chains
- Support for multiple proxy layers (CDN, load balancer, reverse proxy)
- X-Forwarded-For header processing with validation
- X-Real-IP header support for single-proxy setups

SUPPORTED HEADERS:
- X-Forwarded-For: Client IP through proxy chain
- X-Real-IP: Real client IP address
- X-Forwarded-Proto: Original protocol (http/https)
- X-Forwarded-Host: Original host header
- X-Forwarded-Port: Original port number
- Forwarded: RFC 7239 standard forwarded header
- CF-Connecting-IP: Cloudflare specific client IP
- True-Client-IP: Akamai specific client IP
- X-Original-Forwarded-For: Original forwarded header

TRUSTED IP CONFIGURATION:
DEFAULT TRUSTED RANGES (automatically included):
- 127.0.0.1/32: IPv4 loopback
- ::1/128: IPv6 loopback
- 10.0.0.0/8: Private IPv4 class A
- 172.16.0.0/12: Private IPv4 class B
- 192.168.0.0/16: Private IPv4 class C
- fc00::/7: Unique local IPv6 addresses
- fe80::/10: Link-local IPv6 addresses

CUSTOM TRUSTED IPS:
- Admin-configurable additional trusted IP ranges
- Support for CIDR notation and individual IP addresses
- Automatic deduplication of overlapping ranges
- Real-time updates without service restart
- Validation of IP address formats and ranges

IP VALIDATION AND SECURITY:
- Header validation: verify proxy headers against trusted IPs
- Spoofing protection: ignore headers from untrusted sources
- IP address normalization: consistent IPv4/IPv6 handling
- Geographic IP blocking: optional country-based access control
- Rate limiting by real client IP: accurate rate limiting through proxies

PROXY CHAIN HANDLING:
- Multi-hop proxy support: handle complex proxy configurations
- Proxy order validation: verify expected proxy chain order
- Header priority: configurable header precedence for IP detection
- Proxy authentication: support for proxy authentication headers
- Health check exclusions: exclude health check IPs from logging

=============================================================================
FILE STRUCTURE AND CODE ORGANIZATION
=============================================================================

PROJECT ROOT STRUCTURE:
```
casreg/
├── main.go                     # Application entry point with server initialization
├── go.mod                      # Go module definition with all dependencies
├── go.sum                      # Go module checksums for dependency verification
├── Makefile                    # Build automation with targets for all platforms
├── README.md                   # Comprehensive documentation and quick start guide
├── LICENSE                     # MIT license text
├── .env.example               # Complete environment variable examples with descriptions
├── .gitignore                 # Git ignore rules for build artifacts and sensitive files
├── .dockerignore              # Docker ignore rules for efficient image building
├── swagger.yaml               # Auto-generated API documentation in OpenAPI format
├── CHANGELOG.md               # Version history and release notes
├── CONTRIBUTING.md            # Contribution guidelines and development setup
├── SECURITY.md                # Security policy and vulnerability reporting
```

CONFIGURATION MODULE:
```
config/
├── config.go                  # Main configuration structure and loading logic
├── validation.go              # Comprehensive configuration validation rules
├── defaults.go                # Default configuration values and constants
├── database.go                # Database-specific configuration handling
├── storage.go                 # Storage backend configuration management
├── security.go                # Security-related configuration validation
└── migrations.go              # Configuration migration between versions
```

DATA MODELS:
```
models/
├── user.go                    # User model with authentication and profile management
├── organization.go            # Organization model with membership and permissions
├── registry.go                # Registry and repository models with OCI compliance
├── token.go                   # Authentication token model with scoping
├── ticket.go                  # Support ticket model with workflow management
├── notification.go            # Notification model with multi-channel delivery
├── issue.go                   # Repository issue tracking model
├── quota.go                   # Quota management and enforcement model
├── audit.go                   # Audit logging model for security compliance
├── migrations.go              # Database migration scripts and version management
└── base.go                    # Common model structures and interfaces
```

HTTP HANDLERS:
```
handlers/
├── auth.go                    # Authentication endpoints (login, logout, register, tokens)
├── users.go                   # User management endpoints (profile, settings, preferences)
├── organizations.go           # Organization management endpoints (CRUD, membership)
├── registries.go              # Registry management endpoints (CRUD, configuration)
├── repositories.go            # Repository endpoints (Docker Registry V2 API implementation)
├── tags.go                    # Tag management endpoints (listing, deletion, metadata)
├── tokens.go                  # Token management endpoints (create, rotate, delete)
├── admin.go                   # Administrative endpoints (system management, configuration)
├── support.go                 # Support system endpoints (tickets, documentation, FAQ)
├── issues.go                  # Repository issue tracking endpoints
├── search.go                  # Search endpoints (registries, repositories, documentation)
├── webhooks.go                # Webhook management and delivery endpoints
├── swagger.go                 # Swagger UI handler with theme integration
├── static.go                  # Static file serving for UI assets
└── middleware.go              # Custom middleware functions
```

STORAGE ABSTRACTION:
```
storage/
├── interface.go               # Storage interface definition and contracts
├── local.go                   # Local filesystem storage implementation
├── s3.go                      # S3-compatible storage implementation (AWS S3, MinIO)
├── nfs.go                     # Network File System storage implementation
├── memory.go                  # In-memory storage for testing and development
├── utils.go                   # Storage utility functions (validation, path handling)
├── migration.go               # Storage backend migration utilities
└── compression.go             # Compression utilities for storage optimization
```

BACKGROUND SCHEDULER:
```
scheduler/
├── scheduler.go               # Main scheduler with task queue management
├── cleanup.go                 # Cleanup tasks (orphaned files, logs, sessions)
├── notifications.go           # Notification processing and delivery
├── scanning.go                # Vulnerability scanning task coordination
├── audit.go                   # Audit log rotation and maintenance
├── quotas.go                  # Quota monitoring and enforcement
├── health.go                  # System health checks and monitoring
└── tasks.go                   # Task interface and common task utilities
```

SECURITY COMPONENTS:
```
security/
├── trivy.go                   # Embedded Trivy scanner integration
├── cosign.go                  # Embedded Cosign signature verification
├── scanning.go                # Scan orchestration and result processing
├── signatures.go              # Signature verification workflow
├── policies.go                # Security policy enforcement
├── crypto.go                  # Cryptographic utilities and key management
└── vault.go                   # Secret management and encryption
```

WEB USER INTERFACE:
```
ui/
├── public/                    # Static assets (favicon, manifest, robots.txt)
│   ├── favicon.ico
│   ├── manifest.json
│   └── robots.txt
├── src/                       # Source code for web application
│   ├── App.svelte             # Main application component
│   ├── main.js                # Application entry point
│   ├── routes/                # Route definitions and page components
│   │   ├── login.svelte
│   │   ├── dashboard.svelte
│   │   ├── registries.svelte
│   │   ├── organizations.svelte
│   │   ├── profile.svelte
│   │   ├── admin.svelte
│   │   └── support.svelte
│   ├── components/            # Reusable UI components
│   │   ├── header.svelte
│   │   ├── sidebar.svelte
│   │   ├── modal.svelte
│   │   ├── table.svelte
│   │   ├── form.svelte
│   │   └── notifications.svelte
│   ├── stores/                # State management stores
│   │   ├── auth.js
│   │   ├── user.js
│   │   ├── organizations.js
│   │   ├── registries.js
│   │   └── notifications.js
│   ├── api/                   # API client functions
│   │   ├── auth.js
│   │   ├── users.js
│   │   ├── organizations.js
│   │   ├── registries.js
│   │   └── admin.js
│   ├── utils/                 # Utility functions
│   │   ├── formatting.js
│   │   ├── validation.js
│   │   └── helpers.js
│   └── themes/                # Theme definitions
│       ├── dracula.css
│       ├── dark.css
│       └── light.css
├── package.json               # Frontend dependencies and build scripts
├── vite.config.js             # Vite build configuration
├── tailwind.config.js         # Tailwind CSS configuration
└── postcss.config.js          # PostCSS configuration for CSS processing
```

COMMAND-LINE INTERFACE:
```
cli/
├── main.go                    # CLI entry point and command parsing
├── commands/                  # Command implementations
│   ├── admin.go               # Administrative commands
│   ├── user.go                # User management commands
│   ├── organization.go        # Organization commands
│   ├── registry.go            # Registry management commands
│   ├── token.go               # Token management commands
│   ├── support.go             # Support system commands
│   └── config.go              # Configuration commands
├── tui/                       # Terminal UI components
│   ├── dashboard.go           # Interactive dashboard
│   ├── browser.go             # Registry/repository browser
│   ├── forms.go               # Interactive forms
│   └── tables.go              # Data tables and lists
├── themes/                    # CLI theme definitions
│   ├── dracula.go
│   ├── dark.go
│   └── light.go
└── utils/                     # CLI utility functions
    ├── formatting.go
    ├── input.go
    └── output.go
```

DOCUMENTATION AND SUPPORT:
```
support/
├── docs/                      # Built-in documentation system
│   ├── installation/          # Installation guides
│   │   ├── linux.md
│   │   ├── macos.md
│   │   ├── windows.md
│   │   └── docker.md
│   ├── configuration/         # Configuration documentation
│   │   ├── basic-setup.md
│   │   ├── database.md
│   │   ├── storage.md
│   │   └── security.md
│   ├── usage/                 # Usage guides
│   │   ├── web-interface.md
│   │   ├── cli-usage.md
│   │   ├── api-reference.md
│   │   └── docker-client.md
│   ├── administration/        # Administrative guides
│   │   ├── user-management.md
│   │   ├── organization-management.md
│   │   ├── backup-restore.md
│   │   └── troubleshooting.md
│   └── development/           # Developer documentation
│       ├── building.md
│       ├── contributing.md
│       ├── api-development.md
│       └── plugin-development.md
├── templates/                 # Email and notification templates
│   ├── welcome.html
│   ├── password-reset.html
│   ├── notification.html
│   └── security-alert.html
└── tickets/                   # Support ticket system implementation
    ├── handler.go
    ├── workflow.go
    └── analytics.go
```

MIDDLEWARE COMPONENTS:
```
middleware/
├── auth.go                    # Authentication middleware with token validation
├── cors.go                    # CORS handling with configurable origins
├── logging.go                 # Request logging with structured output
├── ratelimit.go               # Rate limiting with Redis backend support
├── recovery.go                # Panic recovery with error reporting
├── proxy.go                   # Reverse proxy header handling
├── compression.go             # Response compression (gzip, deflate)
└── security.go                # Security headers and CSRF protection
```

UTILITY FUNCTIONS:
```
utils/
├── crypto.go                  # Cryptographic utilities (hashing, encryption, signing)
├── validation.go              # Input validation and sanitization
├── helpers.go                 # General helper functions
├── formatting.go              # Data formatting and conversion utilities
├── email.go                   # Email sending and template processing
├── compression.go             # File compression and decompression utilities
├── monitoring.go              # System monitoring and health check utilities
└── testing.go                 # Testing utilities and mock implementations
```

DOCKER DEPLOYMENT:
```
docker/
├── Dockerfile                 # Multi-stage Dockerfile for production builds
├── Dockerfile.dev             # Development Dockerfile with hot reload
├── docker-compose.yml         # Development environment with dependencies
├── docker-compose.prod.yml    # Production environment configuration
├── nginx.conf                 # Nginx reverse proxy configuration
└── scripts/                   # Docker utility scripts
    ├── build.sh
    ├── deploy.sh
    └── backup.sh
```

TEST STRUCTURE:
```
tests/
├── unit/                      # Unit tests for individual components
├── integration/               # Integration tests for component interaction
├── api/                       # API endpoint tests
├── ui/                        # User interface tests
├── performance/               # Performance and load tests
└── fixtures/                  # Test data and fixtures
```

=============================================================================
BUILD SYSTEM AND DEPLOYMENT
=============================================================================

MAKEFILE TARGETS:
```makefile
# Default target
all: build

# Build targets
build: build-server build-cli build-ui
build-server: clean-server go-mod-download go-generate go-build-server
build-cli: clean-cli go-mod-download go-build-cli
build-ui: clean-ui npm-install npm-build-ui

# Cross-compilation targets
build-linux: build-linux-amd64 build-linux-arm64 build-linux-arm
build-darwin: build-darwin-amd64 build-darwin-arm64
build-windows: build-windows-amd64

# Development targets
dev: build-dev run-dev
test: test-unit test-integration test-api
lint: go-lint ui-lint
format: go-format ui-format

# Docker targets
docker-build: docker-build-app docker-build-dev
docker-push: docker-push-app docker-push-dev
docker-run: docker-compose-up
docker-stop: docker-compose-down

# Release targets
release: clean test lint build-all package-all sign-all
package: package-linux package-darwin package-windows
sign: sign-linux sign-darwin sign-windows

# Maintenance targets
clean: clean-build clean-dist clean-docker
update-deps: go-mod-update npm-update
security-scan: go-security-scan npm-security-scan

# Database targets
migrate-up: database-migrate-up
migrate-down: database-migrate-down
seed-data: database-seed

# Documentation targets
docs-build: swagger-generate docs-generate
docs-serve: docs-serve-local
```

GO BUILD CONFIGURATION:
- CGO_ENABLED=0: static binary compilation
- -ldflags="-s -w": strip debug information for smaller binaries
- -ldflags="-X main.version=": embed version information
- -ldflags="-X main.buildTime=": embed build timestamp
- -ldflags="-X main.gitCommit=": embed git commit hash
- -trimpath: remove local path information from binaries

CROSS-COMPILATION TARGETS:
- linux/amd64: Standard Linux 64-bit
- linux/arm64: ARM 64-bit (Apple Silicon, AWS Graviton)
- linux/arm: ARM 32-bit (Raspberry Pi)
- darwin/amd64: Intel Mac
- darwin/arm64: Apple Silicon Mac
- windows/amd64: Windows 64-bit

UI BUILD CONFIGURATION:
- Vite production build with optimization
- Asset minification and compression
- CSS purging for unused styles
- Bundle splitting for optimal loading
- Service worker generation for offline support

DOCKER IMAGE STRUCTURE:
```dockerfile
# Multi-stage build for optimal image size
FROM golang:1.22-alpine AS go-builder
FROM node:18-alpine AS ui-builder  
FROM alpine:latest AS runtime

# Final image with minimal dependencies
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/casreg .
COPY --from=ui-builder /app/dist ./ui/dist
EXPOSE 8080
CMD ["./casreg"]
```

RELEASE PACKAGING:
- Tar.gz archives for Linux and macOS
- ZIP archives for Windows
- Debian packages (.deb) for Ubuntu/Debian
- RPM packages (.rpm) for RHEL/CentOS/Fedora
- Homebrew formula for macOS
- Docker images for containerized deployment

BINARY SIGNING:
- Code signing for macOS binaries (Apple Developer Certificate)
- Authenticode signing for Windows binaries
- GPG signatures for Linux packages
- Checksums (SHA256) for all release artifacts

=============================================================================
DOCUMENTATION REQUIREMENTS
=============================================================================

README.md STRUCTURE:
1. Project overview and key features
2. Quick start guide with Docker one-liner
3. Installation instructions for all platforms
4. Basic configuration examples
5. Usage examples for CLI, API, and UI
6. Docker deployment instructions
7. Development setup guide
8. Contributing guidelines
9. License information
10. Support and community links

INSTALLATION DOCUMENTATION:
- Platform-specific installation guides
- Prerequisites and system requirements
- Step-by-step installation procedures
- Configuration examples and best practices
- Troubleshooting common installation issues
- Migration guides from other registry solutions

API DOCUMENTATION:
- Complete OpenAPI/Swagger specification
- Interactive API explorer with authentication
- Request/response examples for all endpoints
- Error codes and troubleshooting guide
- SDK examples in multiple programming languages
- Rate limiting and usage guidelines

USER GUIDES:
- Web interface walkthrough with screenshots
- CLI usage guide with examples
- Registry management best practices
- Organization and user management guides
- Security configuration recommendations
- Backup and disaster recovery procedures

ADMINISTRATOR DOCUMENTATION:
- System administration guide
- Configuration reference manual
- Database management procedures
- Monitoring and alerting setup
- Performance tuning recommendations
- Security hardening checklist

DEVELOPER DOCUMENTATION:
- API development guide
- Plugin development framework
- Contributing guidelines and code standards
- Testing procedures and requirements
- Build and deployment instructions
- Architecture overview and design decisions

=============================================================================
TESTING REQUIREMENTS
=============================================================================

UNIT TESTING:
- 90% code coverage minimum requirement
- Test all public functions and methods
- Mock external dependencies (database, storage, APIs)
- Test error conditions and edge cases
- Benchmark performance-critical functions

INTEGRATION TESTING:
- Database integration tests with real database connections
- Storage backend integration tests
- Authentication flow testing
- API endpoint integration tests
- Cross-component interaction testing

API TESTING:
- Comprehensive API endpoint testing
- Authentication and authorization testing
- Input validation and error handling testing
- Rate limiting and security testing
- Docker Registry V2 API compliance testing

UI TESTING:
- Component unit testing with testing library
- End-to-end testing with Playwright or Cypress
- Cross-browser compatibility testing
- Responsive design testing
- Accessibility testing (WCAG compliance)

PERFORMANCE TESTING:
- Load testing for API endpoints
- Stress testing for concurrent operations
- Database performance testing
- Storage backend performance testing
- Memory usage and leak testing

SECURITY TESTING:
- Authentication bypass testing
- Authorization escalation testing
- Input validation and injection testing
- CSRF and XSS vulnerability testing
- Security header validation testing

=============================================================================
OPERATIONAL REQUIREMENTS
=============================================================================

LOGGING SYSTEM:
- Structured logging in JSON format
- Configurable log levels (debug, info, warn, error, fatal)
- Log rotation with compression and retention policies
- Audit logging for security and compliance
- Performance logging for monitoring and optimization

MONITORING AND HEALTH CHECKS:
- Health check endpoint (/health) with comprehensive status
- Metrics endpoint (/metrics) for Prometheus integration
- Application performance monitoring (APM) integration
- Database connection health monitoring
- Storage backend health monitoring

BACKUP AND RECOVERY:
- Database backup automation with configurable schedules
- Storage backend backup integration
- Configuration backup and versioning
- Point-in-time recovery capabilities
- Disaster recovery procedures and testing

SECURITY OPERATIONS:
- Security event logging and alerting
- Vulnerability scanning integration
- Incident response procedures
- Security patch management
- Compliance reporting (SOC 2, ISO 27001)

PERFORMANCE OPTIMIZATION:
- Database query optimization and indexing
- Caching strategies for frequently accessed data
- Content delivery network (CDN) integration
- Compression and optimization for large file transfers
- Resource usage monitoring and alerting

=============================================================================
DEPLOYMENT SCENARIOS
=============================================================================

SINGLE-NODE DEPLOYMENT:
- All components running on single server
- SQLite database for simplicity
- Local filesystem storage
- Built-in SMTP for notifications
- Suitable for small teams and testing

MULTI-NODE DEPLOYMENT:
- Separate database server (PostgreSQL/MongoDB)
- Distributed storage (S3/MinIO cluster)
- External Redis for real-time features
- Load balancer for high availability
- Suitable for production environments

KUBERNETES DEPLOYMENT:
- Helm chart for easy deployment
- ConfigMaps for configuration management
- Secrets for sensitive configuration
- Persistent volumes for storage
- Horizontal pod autoscaling

CLOUD DEPLOYMENT:
- AWS deployment with RDS and S3
- Azure deployment with Cosmos DB and Blob Storage
- Google Cloud deployment with Cloud SQL and Cloud Storage
- Terraform modules for infrastructure as code

EDGE DEPLOYMENT:
- Lightweight deployment for edge locations
- Local caching with remote synchronization
- Bandwidth-optimized operation
- Offline capability with sync when connected

=============================================================================
COMPLIANCE AND SECURITY
=============================================================================

SECURITY STANDARDS:
- OWASP Top 10 compliance
- NIST Cybersecurity Framework alignment
- ISO 27001 security controls implementation
- SOC 2 Type II audit readiness
- GDPR privacy compliance

DATA PROTECTION:
- Encryption at rest for sensitive data
- Encryption in transit with TLS 1.3
- Key management and rotation procedures
- Data anonymization and pseudonymization
- Right to be forgotten implementation

ACCESS CONTROL:
- Role-based access control (RBAC)
- Principle of least privilege
- Multi-factor authentication support
- Session management and timeout
- Audit trail for all access events

VULNERABILITY MANAGEMENT:
- Regular security assessments
- Automated vulnerability scanning
- Security patch management
- Penetration testing procedures
- Bug bounty program consideration

=============================================================================
FINAL IMPLEMENTATION NOTES
=============================================================================

CRITICAL SUCCESS FACTORS:
1. All components must work seamlessly together without external configuration
2. The application must compile and run successfully on first attempt
3. All features must be fully functional as specified
4. Documentation must be complete and accurate
5. Security must be implemented correctly throughout

QUALITY ASSURANCE:
- Code must follow Go best practices and idioms
- UI must be responsive and accessible
- API must be fully REST-compliant
- All error conditions must be handled gracefully
- Performance must meet specified requirements

DEPLOYMENT VERIFICATION:
- Application starts successfully with default configuration
- Database migrations run automatically
- Web interface loads and functions correctly
- CLI commands execute successfully
- API endpoints respond correctly
- Docker images build and run successfully

This specification is COMPLETE and FINAL. Implementation must follow every detail exactly as specified. No deviations, assumptions, or interpretations are permitted. The resulting application must be production-ready and fully functional upon first compilation.
```

