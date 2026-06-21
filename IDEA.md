## Project description

casreg is a self-hosted, public-first image registry intended as a complete drop-in replacement for Docker Hub, GitHub Container Registry (GHCR), Quay.io, and GoHarbor — and as a self-hosted alternative to images.linuxcontainers.org for Incus images. No account is required to browse, search, or pull public images — registration is only needed to push or manage private content. The platform hosts both OCI/Docker images (via the Docker Registry V2 API) and Incus container/VM images (via the Incus REST API and simplestreams protocol), with supply-chain security (Trivy scanning, Cosign signing, SBOM generation, SLSA provenance), pull-through proxy caching, registry replication, robot/service accounts, SSO federation, and a comprehensive support system — all as a single self-contained binary with a server-side-rendered web UI and an interactive TUI CLI.

## Project variables

project_name: casreg
project_org: webappsgo
internal_name: casreg
app_name: casreg
official_site: https://github.com/webappsgo/casreg

## Business logic

### Product scope & non-goals

**In scope:**
- Full Docker Registry HTTP API V2 (distribution spec) compliance
- OCI Distribution Specification compliance including the OCI Referrers API
- Incus REST API compliance (`/1.0/images`, `/1.0/operations`, `/1.0/` server info) for authenticated push and pull
- Simplestreams protocol for anonymous public Incus image discovery and pull
- Public-first anonymous browsing, searching, and pulling of public registries — no account required
- Docker Hub-style public landing page with three mutually exclusive sections — each public image appears in at most one, assigned by priority: (1) **Spotlight**: images with a non-empty description, no critical CVEs, and pull count above the configured threshold, recalculated daily; an admin `is_pinned` flag forces an image into Spotlight regardless of metrics; (2) **Newest**: images pushed within the last 7 days that did not qualify for Spotlight; (3) **Most Pulled**: images in the top-N by rolling 30-day pull count that aged out of Newest and did not qualify for Spotlight — plus a global search bar with filters (image type, OS, architecture); all browsable without login
- Image detail pages showing markdown description, available tags with digest/size/platform metadata, a copy-ready pull command, vulnerability scan summary, and signature verification status
- Organization and user profile pages listing public registries, member count, and featured images
- Global registry: a system-owned namespace (default name `library` for OCI images; top-level simplestreams catalog for Incus images) managed exclusively by server admins; unqualified image names resolve to it (`casreg pull alpine` → `library/alpine`; `incus image copy casreg:oracle/7/default` → global Incus catalog); admins can push directly or promote any public image from a user namespace into the global registry; if a requested unqualified OCI name is not found locally, casreg transparently proxies to Docker Hub and caches the result into the global registry on first fetch; if a requested global Incus alias is not found locally, casreg transparently proxies to `images.linuxcontainers.org` and caches the result; this global-registry pull-through is independent of user-configured pull-through caches
- OCI image push/pull with content-addressable, SHA256-verified blob storage
- Incus image push: `lxd.tar.xz` metadata and rootfs file pair received, metadata.yaml extracted automatically to populate os/release/variant/arch/type properties; images identified by combined SHA256 fingerprint
- Both Incus image types: container (rootfs tarball or squashfs) and VM (disk image)
- Incus image alias management: multiple aliases per image, scoped by namespace prefix (`{namespace}/os/release/variant`); aliases are globally unique within the casreg instance; the global Incus registry responds to unqualified aliases (no namespace prefix)
- Async operation model for Incus push: operation ID returned immediately, client polls for completion
- Lightweight in-registry image operations requiring no container runtime:
  - **Cross-namespace copy**: copy any tag to another namespace or registry (`casreg copy {src}/image:tag {dst}/image:tag`); creates new Tag and Manifest records pointing to the same blob digests and increments each blob's ref_count — zero additional storage due to content-addressable deduplication; requires read permission on source and write permission on destination
  - **Multi-arch manifest assembly**: combine existing single-arch manifests already present in the registry into a multi-arch manifest index — pure metadata, no layer data moved
  - **Incus cross-namespace copy**: copy an Incus image to another namespace by fingerprint; creates a new IncusImage row and IncusAlias records referencing the same on-disk rootfs and metadata files — no file copy, new database rows only
  - All operations available via the management API, web UI image detail page, and the casreg CLI
- Native CLI companion (`casreg`) that speaks both OCI V2 and Incus REST directly — not a wrapper around the docker or incus CLIs; `casreg login`, `casreg push`, `casreg pull`, `casreg copy`, `casreg tag`, `casreg rm`, `casreg search` work for both image types; type is detected from the image reference format or overridden with `--type`
- Supply-chain security: Trivy CVE scanning (Incus container images supported; VM disk images excluded — see non-goals), Cosign signature verification, SBOM generation (Syft), SLSA provenance attestation
- Pull-through proxy cache for upstream OCI registries (Docker Hub, GHCR, Quay, gcr.io)
- Registry-to-registry replication (push/pull sync)
- Multi-database support: SQLite (default), PostgreSQL, MySQL/MariaDB, MSSQL, MongoDB
- Multi-storage backend: local filesystem (default), S3-compatible, NFS
- Organizational hierarchy: users → organizations → registries → repositories; a registry is typed as either OCI or Incus — the same org/namespace tree, same visibility and access control, same quota system
- Robot/service accounts for CI/CD automation
- SSO federation: OIDC/OAuth2 (GitHub, GitLab, Google, Entra ID), LDAP/Active Directory
- Webhook delivery with HMAC-SHA256 payload signing
- Immutable append-only audit log with JSON/CSV export
- Integrated support ticket system and knowledge base
- Per-repository issue tracking
- Management REST API (versioned, rate-limited, paginated)

**Non-goals (explicit):**
- Docker Registry V1 protocol — deprecated since Docker 1.6 (2015), removed from Docker Hub, no modern client uses it; implementing it would be a security liability with zero real users
- LXD protocol support — Incus is the actively maintained fork; LXD reached end-of-life for community support
- Incus VM disk image scanning — scanning a VM disk image (`.qcow2`, `.img`) requires mounting the filesystem, which needs kernel privileges incompatible with a static binary deployment; container rootfs tarballs are scanned normally
- Client-side rendering — web UI is server-rendered; no JavaScript framework required
- Native mobile apps
- Built-in CI/CD pipeline execution (use webhooks to trigger external CI)
- Image build service involving Dockerfile execution or arbitrary RUN commands — requires a privileged container runtime on the server, conflicts with the single static binary model, and is an unacceptable attack surface; use webhooks to trigger external CI instead
- Helm chart repository (OCI-based Helm is in scope as it uses the same V2 API)
- Incus cluster federation or instance-to-instance live migration

### Roles & permissions

**System roles (global):**
- `admin` — full system control: user management, global config, audit logs, system-wide scan policies, push/promote to the global `library` registry, set the `is_pinned` editorial override on any public image (OCI or Incus)
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
- Tags are mutable pointers scoped to their own namespace and repository; overwriting `{userB}/alpine:latest` (including after a cross-namespace copy) only updates that tag record and never affects any tag in another namespace — a user can only modify tags they have write permission on
- Blobs are immutable and content-addressed by SHA256 digest; ref_count tracks all references across all namespaces; a blob is only eligible for GC deletion when ref_count reaches zero — overwriting or deleting a tag in one namespace cannot cause blob deletion while any other namespace still holds a reference to that digest

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
- `Registry` — id, org_id (nullable for personal), name, image_type (`oci` or `incus`), is_global (bool; at most one global OCI registry and one global Incus registry per casreg instance), visibility, immutable_tags, retention_policy, scan_policy, created_at
- `Repository` — id, registry_id, name, visibility (≤ parent), description, pull_count, is_pinned (admin override for spotlight), created_at
- `Tag` — id, repository_id, name, manifest_digest, pushed_by, pushed_at, immutable
- `Manifest` — id, repository_id, digest (sha256:…), media_type, size_bytes, config_digest, created_at
- `Blob` — digest (PK), size_bytes, stored_at, ref_count (for deduplication GC)
- `IncusImage` — id, registry_id, fingerprint (SHA256 of combined metadata+rootfs files, unique per registry), os, release, variant, arch, image_type (`container` or `vm`), metadata_size_bytes, rootfs_size_bytes, pull_count, is_pinned (admin override for spotlight), pushed_by, pushed_at
- `IncusAlias` — id, incus_image_id, name (alias string e.g. `oracle/7/default`), description
- `IncusOperation` — id, uuid (unique), operation_type, status (`running`/`success`/`failure`), error, created_at, updated_at; short-lived, pruned after 24 hours
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
- `lxd.tar.xz` content pushed by any user — treated as an untrusted archive; only `metadata.yaml` is extracted; all template path entries are validated against path traversal before any file operation; archive is never executed
- `metadata.yaml` field values (os, release, variant, arch, type) — stored as opaque strings, length-limited, never interpolated into queries, commands, or file paths

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
| Tar-slip via malicious lxd.tar.xz | Template path entries inside lxd.tar.xz validated against path traversal (no `..` components, no absolute paths) before extraction; metadata.yaml extracted to memory only, never written to a user-controlled path |
| Large VM image upload exhausting storage | Per-user and per-org quota limits enforced before accepting any Incus image upload; push rejected if remaining quota would be exceeded |
| Malicious rootfs tarball | Incus container rootfs tarballs scanned by Trivy on push; configurable CVE severity gate can block finalization; rootfs bytes stored opaquely and never executed by casreg |

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
- **VM disk image scanning is excluded** — scanning a VM disk image (`.qcow2`, `.img`, `.vhd`) requires mounting the filesystem, which demands kernel-level privileges incompatible with a static binary deployed in an unprivileged container. Incus container rootfs tarballs are scanned normally by Trivy.
- **Incus metadata is stored, never interpolated** — os, release, variant, arch, and type values extracted from `metadata.yaml` are stored as opaque database strings and displayed verbatim in the UI; they are never used to construct file paths, shell commands, or database queries.
- **Template entries in lxd.tar.xz are path-traversal-validated** — the Incus image format allows template files inside the metadata archive; every path entry is checked for `..` components and absolute prefixes before any extraction; violating entries cause the push to be rejected with a descriptive error.
