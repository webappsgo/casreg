## Project description

casreg is a self-hosted, public-first OCI container registry intended as a complete drop-in replacement for Docker Hub, GitHub Container Registry (GHCR), Quay.io, and GoHarbor. No account is required to browse, search, or pull public images — registration is only needed to push or manage private content. The platform delivers OCI-compliant image storage, supply-chain security (Trivy scanning, Cosign signing, SBOM generation, SLSA provenance), pull-through proxy caching, registry replication, robot/service accounts, SSO federation, and a comprehensive support system — all as a single self-contained binary with a server-side-rendered web UI and an interactive TUI CLI.

## Project variables

project_name: casreg
project_org: webappsgo
internal_name: casreg
app_name: casreg
official_site: https://github.com/webappsgo/casreg

## Business logic

### Product scope & non-goals

**In scope:**
- Full Docker Registry HTTP API V2 (distribution spec) compliance at `/v2/`
- OCI Distribution Specification compliance including the OCI Referrers API
- Public-first anonymous browsing, searching, and pulling of public registries — no account required
- Image push/pull with content-addressable, SHA256-verified blob storage
- Supply-chain security: Trivy CVE scanning, Cosign signature verification, SBOM generation (Syft), SLSA provenance attestation
- Pull-through proxy cache for upstream registries (Docker Hub, GHCR, Quay, gcr.io)
- Registry-to-registry replication (push/pull sync)
- Multi-database support: SQLite (default), PostgreSQL, MySQL/MariaDB, MSSQL, MongoDB
- Multi-storage backend: local filesystem (default), S3-compatible, NFS
- Organizational hierarchy: users → organizations → registries → repositories
- Robot/service accounts for CI/CD automation
- SSO federation: OIDC/OAuth2 (GitHub, GitLab, Google, Entra ID), LDAP/Active Directory
- Webhook delivery with HMAC-SHA256 payload signing
- Immutable append-only audit log with JSON/CSV export
- Integrated support ticket system and knowledge base
- Per-repository issue tracking
- Interactive TUI CLI companion binary
- Management REST API (versioned, rate-limited, paginated)

**Non-goals (explicit):**
- Docker Registry V1 protocol — deprecated since Docker 1.6 (2015), removed from Docker Hub, no modern client uses it; implementing it would be a security liability with zero real users
- Client-side rendering — web UI is server-rendered; no JavaScript framework required
- Native mobile apps
- Built-in CI/CD pipeline execution (use webhooks to trigger external CI)
- Image build service (casreg stores and serves images; building them is out of scope)
- Helm chart repository (OCI-based Helm is in scope as it uses the same V2 API)

### Roles & permissions

**System roles (global):**
- `admin` — full system control: user management, global config, audit logs, system-wide scan policies
- `user` — default for all accounts after the first; manages their own registries and orgs

**Organization roles:**
- `owner` — full org control including billing, member management, and deletion
- `admin` — manage members, registries, and org settings; cannot delete the org or remove owners
- `member` — contribute to repositories within the org per per-registry grants

**Registry/repository grants (additive):**
- `read` — pull images and view metadata
- `write` — push images (implies read)
- `admin` — manage registry settings, tokens, and access rules (implies write)

**Special identities:**
- Robot/service accounts — non-human, scoped to specific registries or orgs; never have system-level roles; credentials auto-expire and must be rotated
- Guests (unauthenticated) — can browse/search/pull from any `public` registry; rate-limited per IP

**Visibility model:**
- Registry levels: `public` (world-readable), `organization` (org members only), `private` (explicit grants only), `internal` (authenticated users on this casreg instance)
- Repository visibility can tighten (never loosen) its parent registry's visibility
- Organization visibility: `public`, `internal`, `private`, `hidden`

**Invariants:**
- First registered user becomes the sole system admin; cannot be demoted while they are the only admin
- Org owner cannot remove themselves if they are the only owner
- Deleting a registry requires explicit confirmation and cascades to all contained repositories, tags, and blobs
- Robot accounts cannot be promoted to system roles under any circumstance

### Data model & sensitivity

**Sensitivity tiers:**

| Tier | Data | Handling |
|------|------|----------|
| Critical | Passwords (bcrypt), JWT signing secret, API token values, OIDC client secrets, LDAP bind credentials, robot account tokens, SMTP password, S3 secret key | Never logged, never returned in API responses after creation, stored hashed or encrypted at rest |
| High | User email addresses, IP addresses from audit logs, webhook secrets, passkey credentials | Accessible only to the owning user and system admins; not included in public API responses |
| Medium | Image metadata, tag names, manifest digests, SBOM content, scan results, CVE data | Readable by anyone with read access to the repository |
| Low | Registry/org names, public image layers, documentation, knowledge base articles | Publicly accessible for public registries |

**Core entities:**
- `User` — id, username, email (unique), bcrypt_password_hash, role, locked_until, passkey_credentials, 2fa_method, created_at
- `Organization` — id, name (unique), visibility, owner_user_id, quota_bytes, created_at
- `Registry` — id, org_id (nullable for personal), name, visibility, immutable_tags, retention_policy, scan_policy, created_at
- `Repository` — id, registry_id, name, visibility (≤ parent), description, created_at
- `Tag` — id, repository_id, name, manifest_digest, pushed_by, pushed_at, immutable
- `Manifest` — id, repository_id, digest (sha256:…), media_type, size_bytes, config_digest, created_at
- `Blob` — digest (PK), size_bytes, stored_at, ref_count (for deduplication GC)
- `Token` — id, user_id, name, hashed_value, scopes, expires_at, last_used_at
- `RobotAccount` — id, owner_type (user/org), owner_id, name, hashed_token, scopes, expires_at
- `ScanResult` — id, manifest_digest, scanner_version, db_version, scanned_at, vulnerabilities (JSONB), sbom_spdx (JSONB), sbom_cyclonedx (JSONB), secrets_found (bool)
- `Signature` — id, manifest_digest, payload_type (cosign/notation), keyless (bool), cert_subject, rekor_log_id, verified_at
- `AuditLog` — id, actor_type (user/robot/guest), actor_id, action, resource_type, resource_id, ip_addr, user_agent, created_at; append-only, no UPDATE/DELETE
- `WebhookConfig` — id, owner_type, owner_id, url, hmac_secret_hash, events, active
- `SupportTicket` — id, reporter_user_id, assignee_user_id, status, priority, title, body, created_at
- `ProxyCache` — id, name, upstream_url, upstream_auth_secret_ref, ttl_hours, last_synced_at
- `ReplicationRule` — id, name, source (local or remote url), dest (local or remote url), filter_pattern, trigger (push/schedule), enabled

### Trust boundaries & external services

**Trusted (inside the trust boundary):**
- Local database (SQLite/PostgreSQL/MySQL/MSSQL/MongoDB) — connection string from config only; never from user input
- Local blob storage or administrator-configured S3/NFS — paths validated and jail-checked against storage root
- Configured upstream OIDC/LDAP providers — certificates verified; JWTs validated against provider's JWKS endpoint
- Internal loopback and RFC 1918 ranges for X-Forwarded-For when `trusted_proxies` is set by an admin

**Untrusted (outside the trust boundary):**
- All HTTP request bodies from anonymous users and authenticated users alike — fully validated and size-limited
- Image layer content — treated as untrusted binary data; never executed; stored by digest only
- OCI manifest JSON — parsed and schema-validated; untrusted field values never interpolated into queries or shell commands
- Upstream pull-through registry responses — content verified by digest before serving to clients; TLS certs required
- Webhook endpoint URLs — SSRF-checked against a blocklist of internal IP ranges before delivery
- OIDC/OAuth2 authorization codes and access tokens from external providers — validated but treated as untrusted input until signature is verified

**External service failure modes:**

| Service | Failure mode | Behavior |
|---------|-------------|----------|
| SMTP | Unavailable | Notification falls back to in-app inbox; no error surfaced to end user |
| Redis/Valkey | Unavailable | Real-time WebSocket features degrade gracefully; rate limiting falls back to in-memory |
| OIDC provider | Unavailable | SSO login fails with a user-visible error; password login still available |
| LDAP/AD | Unavailable | LDAP users cannot log in; local accounts unaffected |
| Upstream proxy registry | Unavailable | Pull-through returns 502; local cache hit still serves successfully |
| Trivy DB update | Unavailable | Scanning uses cached DB (air-gapped mode); scan is not blocked |
| Rekor/Fulcio (Sigstore) | Unavailable | Keyless verification fails closed (image treated as unverified); error surfaced via scan result |

### Threat model & abuse cases

**Primary assets being protected:**
- Private image layers and manifests (confidentiality)
- Image integrity — layers must match their declared SHA256 digest (integrity)
- Auth credentials: passwords, API tokens, robot tokens, OIDC secrets (confidentiality)
- Audit log — must remain append-only and tamper-evident (integrity)
- The registry's blob storage — must not be used as arbitrary file storage (availability/abuse)

**Trusted vs untrusted inputs:**

| Input | Trust level |
|-------|-------------|
| Admin-set config file / environment variables | Trusted |
| Database content written by casreg itself | Trusted |
| Authenticated user API requests (after token validation) | Partially trusted — still validated |
| Anonymous HTTP requests | Untrusted |
| Image layer bytes | Untrusted — verified by digest only |
| OCI manifest JSON | Untrusted — parsed and schema-validated |
| Upstream pull-through responses | Untrusted — digest-verified before serving |
| OIDC/OAuth2 callbacks | Untrusted until signature verified |
| Webhook delivery responses | Irrelevant — casreg is the sender, not the receiver |

**Attacker/abuser goals:**
- Exfiltrate private image content without authorization
- Push malicious images to public registries for supply-chain poisoning
- Use the registry as a free blob CDN by pushing arbitrary binary data
- Credential stuffing against user accounts
- Abuse pull-through cache to pivot to internal network (SSRF)
- Forge or strip image signatures to bypass signature enforcement policy
- Exhaust storage quota of other users/orgs
- Inject malicious content via image metadata (manifest labels, tag names)
- Scrape all public images for secrets embedded in layers

**Abuse cases and required defenses:**

| Abuse case | Defense |
|------------|---------|
| Credential stuffing | bcrypt with cost ≥ 12; account lockout after 5 failures for 15 min; constant-time comparison for all token checks |
| Scraping of public images | Per-IP rate limiting for anonymous requests (100 req/hour default); exponential backoff on 429 |
| Blob storage abuse (arbitrary file upload) | Blobs only accepted via the standardized Docker V2 upload endpoint; media type validated; digest verified against declared SHA256 |
| SSRF via pull-through or webhook URLs | Webhook and proxy upstream URLs blocked against RFC 1918, loopback, link-local, and metadata ranges before any outbound connection |
| Supply-chain poisoning via push | Scan gates block push or pull above configured CVE severity; signature enforcement per-registry; tag immutability prevents overwrite |
| Path traversal in storage | Blob paths are derived solely from the hex SHA256 digest; never from user-supplied filenames |
| Manifest injection (YAML/JSON fields) | Manifests parsed with strict schema validation; all string fields are stored as opaque values, never interpolated |
| Secret leakage in image layers | Trufflehog-based secret scanning on push; result surfaced in scan report; policy gate can block pull |
| Quota exhaustion by one org/user | Per-user, per-org, per-registry, and per-repo hard quota limits with configurable grace periods |
| Audit log tampering | AuditLog table has no UPDATE or DELETE grants in the ORM layer; application code only calls INSERT |
| Forged HMAC on webhooks | Outgoing webhooks signed with HMAC-SHA256 using a per-hook secret; receivers advised to verify |
| JWT secret compromise | JWT secret is auto-generated on first run if not set; minimum 256 bits enforced; never logged |
| Replay of API tokens | Tokens are stored hashed (SHA256); raw value shown only once at creation; no plaintext ever persisted |

### Security decisions & exceptions

- **Public anonymous access is intentional** — unauthenticated users can browse, search, and pull from `public` registries. This is the product's core value proposition (replacing Docker Hub's anonymous pull model). Rate limiting and digest-verified blob delivery are the controls.
- **Tag immutability is opt-in per registry** — some workflows (e.g., rolling `latest`) require mutable tags. Immutability is a per-registry setting, not a global default, to avoid breaking standard CI workflows out of the box.
- **SQLite in WAL mode is the default database** — chosen for zero-config first-run. Operators moving to production scale are expected to migrate to PostgreSQL. The in-app migration wizard enforces this.
- **Docker Registry V1 protocol is explicitly excluded** — V1 uses MD5 checksums for layer verification and has no content-addressable model. Implementing it would require deliberately weakening the integrity model. All Docker clients since version 1.6 (April 2015) use V2.
- **Single static binary distribution** — eliminates entire classes of glibc/musl version mismatch vulnerabilities and enables minimal container images. All embedded security tools (scanner, signature verifier, SBOM generator, secret scanner) must be available without external runtime dependencies.
- **OIDC/OAuth2 redirect URIs are admin-configured** — the allowed redirect URI list is never derived from user input. Authorization code flow only; implicit flow is disabled.
- **SSRF protection applies to all outbound connections casreg initiates** — pull-through upstream URLs, webhook delivery targets, and OIDC discovery endpoints are all checked against an internal-IP blocklist before connection. The blocklist includes RFC 1918, loopback, link-local (169.254/16), and the AWS metadata endpoint (169.254.169.254).
- **Passkey/WebAuthn is optional and admin-toggleable** — organizations that require FIDO2 can enforce it; those with hardware constraints can disable it. This is a UX tradeoff, not a security downgrade, because password + token auth is still available.
- **Replication and pull-through always verify TLS** — no option to skip certificate verification; operators who need self-signed certs on upstream registries must add the CA to the system trust store.
