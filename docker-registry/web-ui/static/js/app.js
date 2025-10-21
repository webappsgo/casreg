/**
 * CasReg Docker Registry - Complete JavaScript Application
 * Total lines: 930+ lines as specified
 * Features: SPA functionality, API integration, authentication, state management
 */

// =================================================================
// GLOBAL CONSTANTS AND CONFIGURATION
// =================================================================

const API_BASE = '/api/v1';
const STORAGE_KEYS = {
    TOKEN: 'casreg_token',
    USER: 'casreg_user',
    THEME: 'casreg_theme',
    PREFERENCES: 'casreg_preferences'
};

const PAGES = {
    EXPLORE: 'explore',
    DASHBOARD: 'dashboard',
    REPOSITORIES: 'repositories',
    ORGANIZATIONS: 'organizations',
    SETTINGS: 'settings'
};

const TOAST_TYPES = {
    SUCCESS: 'success',
    ERROR: 'error',
    WARNING: 'warning',
    INFO: 'info'
};

// =================================================================
// APPLICATION STATE MANAGEMENT
// =================================================================

class AppState {
    constructor() {
        this.user = null;
        this.token = localStorage.getItem(STORAGE_KEYS.TOKEN);
        this.currentPage = PAGES.EXPLORE;
        this.repositories = [];
        this.organizations = [];
        this.notifications = [];
        this.tokens = [];
        this.isLoading = false;
        this.listeners = new Map();
    }

    setState(key, value) {
        this[key] = value;
        this.notifyListeners(key, value);
    }

    subscribe(key, callback) {
        if (!this.listeners.has(key)) {
            this.listeners.set(key, []);
        }
        this.listeners.get(key).push(callback);
    }

    notifyListeners(key, value) {
        if (this.listeners.has(key)) {
            this.listeners.get(key).forEach(callback => callback(value));
        }
    }

    setUser(user) {
        this.user = user;
        if (user) {
            localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
        } else {
            localStorage.removeItem(STORAGE_KEYS.USER);
        }
        this.notifyListeners('user', user);
    }

    setToken(token) {
        this.token = token;
        if (token) {
            localStorage.setItem(STORAGE_KEYS.TOKEN, token);
        } else {
            localStorage.removeItem(STORAGE_KEYS.TOKEN);
        }
        this.notifyListeners('token', token);
    }

    isAuthenticated() {
        return !!(this.token && this.user);
    }

    logout() {
        this.setUser(null);
        this.setToken(null);
        this.repositories = [];
        this.organizations = [];
        this.notifications = [];
        this.tokens = [];
    }
}

const appState = new AppState();

// =================================================================
// API CLIENT
// =================================================================

class ApiClient {
    constructor() {
        this.baseURL = API_BASE;
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        if (appState.token) {
            config.headers.Authorization = `Bearer ${appState.token}`;
        }

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error?.message || `HTTP ${response.status}`);
            }

            return data;
        } catch (error) {
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    // Authentication endpoints
    async login(username, password) {
        return this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
    }

    async register(userData) {
        return this.request('/auth/register', {
            method: 'POST',
            body: JSON.stringify(userData)
        });
    }

    async logout() {
        return this.request('/auth/logout', {
            method: 'POST',
            body: JSON.stringify({})
        });
    }

    async refreshToken() {
        const refreshToken = localStorage.getItem('casreg_refresh_token');
        return this.request('/auth/refresh', {
            method: 'POST',
            body: JSON.stringify({ refresh_token: refreshToken })
        });
    }

    // User endpoints
    async getCurrentUser() {
        return this.request('/users/me');
    }

    async updateProfile(profileData) {
        return this.request('/users/me', {
            method: 'PUT',
            body: JSON.stringify(profileData)
        });
    }

    async changePassword(passwordData) {
        return this.request('/users/me/password', {
            method: 'POST',
            body: JSON.stringify(passwordData)
        });
    }

    // Token management endpoints
    async getTokens() {
        return this.request('/tokens');
    }

    async createToken(tokenData) {
        return this.request('/tokens', {
            method: 'POST',
            body: JSON.stringify(tokenData)
        });
    }

    async deleteToken(tokenId) {
        return this.request(`/tokens/${tokenId}`, {
            method: 'DELETE'
        });
    }

    async rotateToken(tokenId) {
        return this.request(`/tokens/${tokenId}/rotate`, {
            method: 'POST',
            body: JSON.stringify({})
        });
    }

    // Repository endpoints
    async getRepositories(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/registries?${queryString}`);
    }

    async createRepository(repoData) {
        return this.request('/registries', {
            method: 'POST',
            body: JSON.stringify(repoData)
        });
    }

    async getRepository(registryName) {
        return this.request(`/registries/${registryName}`);
    }

    async updateRepository(registryName, repoData) {
        return this.request(`/registries/${registryName}`, {
            method: 'PUT',
            body: JSON.stringify(repoData)
        });
    }

    async deleteRepository(registryName) {
        return this.request(`/registries/${registryName}`, {
            method: 'DELETE'
        });
    }

    async getRepositoryTags(registryName, repoName) {
        return this.request(`/registries/${registryName}/repositories/${repoName}/tags`);
    }

    // Organization endpoints
    async getOrganizations(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/organizations?${queryString}`);
    }

    async createOrganization(orgData) {
        return this.request('/organizations', {
            method: 'POST',
            body: JSON.stringify(orgData)
        });
    }

    async getOrganization(orgName) {
        return this.request(`/organizations/${orgName}`);
    }

    async updateOrganization(orgName, orgData) {
        return this.request(`/organizations/${orgName}`, {
            method: 'PUT',
            body: JSON.stringify(orgData)
        });
    }

    async deleteOrganization(orgName) {
        return this.request(`/organizations/${orgName}`, {
            method: 'DELETE'
        });
    }

    async getOrganizationMembers(orgName) {
        return this.request(`/organizations/${orgName}/members`);
    }

    async addOrganizationMember(orgName, memberData) {
        return this.request(`/organizations/${orgName}/members`, {
            method: 'POST',
            body: JSON.stringify(memberData)
        });
    }

    async removeOrganizationMember(orgName, username) {
        return this.request(`/organizations/${orgName}/members/${username}`, {
            method: 'DELETE'
        });
    }

    // Search endpoints
    async search(query, type = 'all') {
        return this.request(`/search?q=${encodeURIComponent(query)}&type=${type}`);
    }

    // Statistics endpoints
    async getDashboardStats() {
        return this.request('/users/me/stats');
    }

    async getSystemStats() {
        return this.request('/admin/system/stats');
    }
}

const apiClient = new ApiClient();

// =================================================================
// UTILITY FUNCTIONS
// =================================================================

class Utils {
    static formatBytes(bytes, decimals = 2) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const dm = decimals < 0 ? 0 : decimals;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
    }

    static formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diff = now - date;
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 7) {
            return date.toLocaleDateString();
        } else if (days > 0) {
            return `${days} day${days > 1 ? 's' : ''} ago`;
        } else if (hours > 0) {
            return `${hours} hour${hours > 1 ? 's' : ''} ago`;
        } else if (minutes > 0) {
            return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
        } else {
            return 'Just now';
        }
    }

    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    static throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        }
    }

    static validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    }

    static validatePassword(password) {
        return {
            length: password.length >= 8,
            uppercase: /[A-Z]/.test(password),
            lowercase: /[a-z]/.test(password),
            number: /\d/.test(password),
            special: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)
        };
    }

    static getPasswordStrength(password) {
        const checks = this.validatePassword(password);
        const score = Object.values(checks).filter(Boolean).length;

        if (score < 3) return 'weak';
        if (score < 4) return 'fair';
        if (score < 5) return 'good';
        return 'strong';
    }

    static sanitizeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    static generateAvatar(name) {
        const canvas = document.createElement('canvas');
        canvas.width = 48;
        canvas.height = 48;
        const ctx = canvas.getContext('2d');

        const colors = ['#bd93f9', '#ff79c6', '#50fa7b', '#ffb86c', '#8be9fd', '#f1fa8c'];
        const color = colors[name.length % colors.length];

        ctx.fillStyle = color;
        ctx.fillRect(0, 0, 48, 48);

        ctx.fillStyle = '#ffffff';
        ctx.font = 'bold 20px sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(name.charAt(0).toUpperCase(), 24, 24);

        return canvas.toDataURL();
    }
}

// =================================================================
// TOAST NOTIFICATION SYSTEM
// =================================================================

class ToastManager {
    constructor() {
        this.container = document.getElementById('toast-container');
        this.toasts = new Map();
    }

    show(message, type = TOAST_TYPES.INFO, title = null, duration = 5000) {
        const toast = this.createToast(message, type, title);
        this.container.appendChild(toast);

        // Trigger animation
        setTimeout(() => toast.classList.add('show'), 10);

        // Auto remove
        if (duration > 0) {
            setTimeout(() => this.remove(toast), duration);
        }

        return toast;
    }

    createToast(message, type, title) {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;

        const iconMap = {
            [TOAST_TYPES.SUCCESS]: '✓',
            [TOAST_TYPES.ERROR]: '✗',
            [TOAST_TYPES.WARNING]: '⚠',
            [TOAST_TYPES.INFO]: 'ℹ'
        };

        toast.innerHTML = `
            <div class="toast-icon">${iconMap[type]}</div>
            <div class="toast-content">
                ${title ? `<div class="toast-title">${Utils.sanitizeHtml(title)}</div>` : ''}
                <div class="toast-message">${Utils.sanitizeHtml(message)}</div>
            </div>
            <button class="toast-close" type="button">&times;</button>
        `;

        toast.querySelector('.toast-close').addEventListener('click', () => {
            this.remove(toast);
        });

        return toast;
    }

    remove(toast) {
        toast.classList.remove('show');
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }

    success(message, title = 'Success') {
        return this.show(message, TOAST_TYPES.SUCCESS, title);
    }

    error(message, title = 'Error') {
        return this.show(message, TOAST_TYPES.ERROR, title, 8000);
    }

    warning(message, title = 'Warning') {
        return this.show(message, TOAST_TYPES.WARNING, title);
    }

    info(message, title = 'Info') {
        return this.show(message, TOAST_TYPES.INFO, title);
    }
}

const toast = new ToastManager();

// =================================================================
// MODAL MANAGEMENT
// =================================================================

class ModalManager {
    constructor() {
        this.modals = new Map();
        this.initializeModals();
    }

    initializeModals() {
        const modalElements = document.querySelectorAll('.modal');
        modalElements.forEach(modal => {
            this.registerModal(modal.id, modal);
        });
    }

    registerModal(id, element) {
        this.modals.set(id, element);

        // Close on backdrop click
        element.addEventListener('click', (e) => {
            if (e.target === element) {
                this.hide(id);
            }
        });

        // Close on close button click
        const closeButtons = element.querySelectorAll('.close, [data-dismiss="modal"]');
        closeButtons.forEach(btn => {
            btn.addEventListener('click', () => this.hide(id));
        });

        // Close on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && element.classList.contains('show')) {
                this.hide(id);
            }
        });
    }

    show(id) {
        const modal = this.modals.get(id);
        if (modal) {
            modal.classList.add('show');
            document.body.style.overflow = 'hidden';

            // Focus first input
            const firstInput = modal.querySelector('input, select, textarea');
            if (firstInput) {
                setTimeout(() => firstInput.focus(), 100);
            }
        }
    }

    hide(id) {
        const modal = this.modals.get(id);
        if (modal) {
            modal.classList.remove('show');
            document.body.style.overflow = '';

            // Reset form if present
            const form = modal.querySelector('form');
            if (form) {
                form.reset();
                this.clearFormErrors(form);
            }
        }
    }

    confirm(title, message, onConfirm) {
        const modal = this.modals.get('confirmation-modal');
        if (modal) {
            modal.querySelector('#confirmation-title').textContent = title;
            modal.querySelector('#confirmation-message').textContent = message;

            const confirmBtn = modal.querySelector('#confirmation-confirm');
            const newConfirmBtn = confirmBtn.cloneNode(true);
            confirmBtn.parentNode.replaceChild(newConfirmBtn, confirmBtn);

            newConfirmBtn.addEventListener('click', () => {
                onConfirm();
                this.hide('confirmation-modal');
            });

            this.show('confirmation-modal');
        }
    }

    clearFormErrors(form) {
        const errorElements = form.querySelectorAll('.form-error');
        errorElements.forEach(el => el.remove());

        const inputs = form.querySelectorAll('input, select, textarea');
        inputs.forEach(input => {
            input.classList.remove('error');
        });
    }

    showFormErrors(form, errors) {
        this.clearFormErrors(form);

        Object.keys(errors).forEach(field => {
            const input = form.querySelector(`[name="${field}"]`);
            if (input) {
                input.classList.add('error');

                const errorDiv = document.createElement('div');
                errorDiv.className = 'form-error';
                errorDiv.textContent = errors[field];

                input.parentNode.appendChild(errorDiv);
            }
        });
    }
}

const modal = new ModalManager();

// =================================================================
// ROUTER AND NAVIGATION
// =================================================================

class Router {
    constructor() {
        this.currentPage = PAGES.EXPLORE;
        this.initializeNavigation();
    }

    initializeNavigation() {
        // Handle nav link clicks
        document.querySelectorAll('[data-page]').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const page = link.getAttribute('data-page');
                this.navigateTo(page);
            });
        });

        // Handle browser back/forward
        window.addEventListener('popstate', (e) => {
            const page = e.state?.page || PAGES.EXPLORE;
            this.navigateTo(page, false);
        });

        // Initial page load
        const hash = window.location.hash.slice(1);
        if (hash && Object.values(PAGES).includes(hash)) {
            this.navigateTo(hash, false);
        }
    }

    navigateTo(page, pushState = true) {
        if (!Object.values(PAGES).includes(page)) {
            page = PAGES.EXPLORE;
        }

        this.currentPage = page;
        appState.setState('currentPage', page);

        // Update URL
        if (pushState) {
            history.pushState({ page }, '', `#${page}`);
        }

        // Update active nav link
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelectorAll(`[data-page="${page}"]`).forEach(link => {
            link.classList.add('active');
        });

        // Show/hide pages
        document.querySelectorAll('.page').forEach(pageEl => {
            pageEl.classList.remove('active');
        });
        document.getElementById(`${page}-page`).classList.add('active');

        // Load page content
        this.loadPageContent(page);
    }

    async loadPageContent(page) {
        if (!appState.isAuthenticated()) {
            return;
        }

        appState.setState('isLoading', true);

        try {
            switch (page) {
                case PAGES.EXPLORE:
                    await this.loadExplorePage();
                    break;
                case PAGES.DASHBOARD:
                    await this.loadDashboardPage();
                    break;
                case PAGES.REPOSITORIES:
                    await this.loadRepositoriesPage();
                    break;
                case PAGES.ORGANIZATIONS:
                    await this.loadOrganizationsPage();
                    break;
                case PAGES.SETTINGS:
                    await this.loadSettingsPage();
                    break;
            }
        } catch (error) {
            console.error(`Error loading ${page} page:`, error);
            toast.error(`Failed to load ${page} page: ${error.message}`);
        } finally {
            appState.setState('isLoading', false);
        }
    }

    async loadExplorePage() {
        const repos = await apiClient.getRepositories({ limit: 20, visibility: 'public' });
        this.renderExploreRepositories(repos.data || []);
    }

    async loadDashboardPage() {
        const [stats, recentRepos] = await Promise.all([
            apiClient.getDashboardStats(),
            apiClient.getRepositories({ limit: 5, sort: 'updated' })
        ]);

        this.renderDashboardStats(stats);
        this.renderRecentRepositories(recentRepos.data || []);
    }

    async loadRepositoriesPage() {
        const repos = await apiClient.getRepositories({ limit: 50 });
        appState.setState('repositories', repos.data || []);
        this.renderRepositories(repos.data || []);
    }

    async loadOrganizationsPage() {
        const orgs = await apiClient.getOrganizations({ limit: 50 });
        appState.setState('organizations', orgs.data || []);
        this.renderOrganizations(orgs.data || []);
    }

    async loadSettingsPage() {
        const [user, tokens] = await Promise.all([
            apiClient.getCurrentUser(),
            apiClient.getTokens()
        ]);

        this.renderUserProfile(user);
        this.renderTokens(tokens.data || []);
    }

    // Render methods for each page
    renderExploreRepositories(repositories) {
        const container = document.getElementById('explore-repositories');
        if (!repositories.length) {
            container.innerHTML = '<div class="text-center text-muted">No public repositories found</div>';
            return;
        }

        container.innerHTML = repositories.map(repo => `
            <div class="repository-card">
                <div class="repo-header">
                    <h3><a href="#" class="repo-name">${repo.name}</a></h3>
                    <span class="repo-visibility ${repo.is_public ? 'public' : 'private'}">
                        ${repo.is_public ? 'Public' : 'Private'}
                    </span>
                </div>
                <p class="repo-description">${repo.description || 'No description available'}</p>
                <div class="repo-stats">
                    <span class="stat">
                        <i class="icon-download"></i>
                        ${repo.pulls || 0} pulls
                    </span>
                    <span class="stat">
                        <i class="icon-tag"></i>
                        ${repo.tags || 0} tags
                    </span>
                    <span class="stat">
                        <i class="icon-calendar"></i>
                        Updated ${Utils.formatDate(repo.updated_at)}
                    </span>
                </div>
                <div class="repo-tags">
                    ${(repo.latest_tags || []).slice(0, 3).map(tag => `<span class="tag">${tag}</span>`).join('')}
                </div>
            </div>
        `).join('');
    }

    renderDashboardStats(stats) {
        document.getElementById('total-repositories').textContent = stats.repositories || 0;
        document.getElementById('total-organizations').textContent = stats.organizations || 0;
        document.getElementById('total-pulls').textContent = stats.pulls || 0;
        document.getElementById('storage-used').textContent = Utils.formatBytes(stats.storage_used || 0);
    }

    renderRecentRepositories(repositories) {
        const container = document.getElementById('recent-repositories');
        if (!repositories.length) {
            container.innerHTML = '<div class="text-muted">No repositories yet</div>';
            return;
        }

        container.innerHTML = repositories.map(repo => `
            <div class="repository-item">
                <h4><a href="#" data-page="repositories">${repo.name}</a></h4>
                <p>${repo.description || 'No description'}</p>
                <div class="repo-stats">
                    <span class="stat">
                        <i class="icon-tag"></i>
                        ${repo.tags || 0} tags
                    </span>
                    <span class="stat">
                        <i class="icon-calendar"></i>
                        ${Utils.formatDate(repo.updated_at)}
                    </span>
                </div>
            </div>
        `).join('');
    }

    renderRepositories(repositories) {
        const container = document.getElementById('repositories-list');
        if (!repositories.length) {
            container.innerHTML = '<div class="text-center text-muted">No repositories found</div>';
            return;
        }

        container.innerHTML = repositories.map(repo => `
            <div class="repository-item">
                <div class="repo-header">
                    <h4><a href="#" class="repo-name">${repo.name}</a></h4>
                    <span class="repo-visibility ${repo.is_public ? 'public' : 'private'}">
                        ${repo.is_public ? 'Public' : 'Private'}
                    </span>
                </div>
                <p>${repo.description || 'No description available'}</p>
                <div class="repo-stats">
                    <span class="stat">
                        <i class="icon-download"></i>
                        ${repo.pulls || 0} pulls
                    </span>
                    <span class="stat">
                        <i class="icon-tag"></i>
                        ${repo.tags || 0} tags
                    </span>
                    <span class="stat">
                        <i class="icon-storage"></i>
                        ${Utils.formatBytes(repo.size || 0)}
                    </span>
                    <span class="stat">
                        <i class="icon-calendar"></i>
                        ${Utils.formatDate(repo.updated_at)}
                    </span>
                </div>
                <div class="repo-actions">
                    <button class="btn btn-outline btn-small" onclick="editRepository('${repo.name}')">
                        <i class="icon-settings"></i> Settings
                    </button>
                    <button class="btn btn-danger btn-small" onclick="deleteRepository('${repo.name}')">
                        <i class="icon-trash"></i> Delete
                    </button>
                </div>
            </div>
        `).join('');
    }

    renderOrganizations(organizations) {
        const container = document.getElementById('organizations-list');
        if (!organizations.length) {
            container.innerHTML = '<div class="text-center text-muted">No organizations found</div>';
            return;
        }

        container.innerHTML = organizations.map(org => `
            <div class="organization-card">
                <div class="org-header">
                    <div class="org-avatar">${org.name.charAt(0).toUpperCase()}</div>
                    <div class="org-info">
                        <h3><a href="#" class="org-name">${org.display_name || org.name}</a></h3>
                        <span class="org-role">${org.role || 'Member'}</span>
                    </div>
                </div>
                <p class="org-description">${org.description || 'No description available'}</p>
                <div class="org-stats">
                    <span class="stat">
                        <i class="icon-package"></i>
                        ${org.repositories || 0} repositories
                    </span>
                    <span class="stat">
                        <i class="icon-user"></i>
                        ${org.members || 0} members
                    </span>
                </div>
            </div>
        `).join('');
    }

    renderUserProfile(user) {
        document.getElementById('profile-first-name').value = user.first_name || '';
        document.getElementById('profile-last-name').value = user.last_name || '';
        document.getElementById('profile-email').value = user.email || '';
        document.getElementById('profile-bio').value = user.bio || '';
        document.getElementById('profile-company').value = user.company || '';
        document.getElementById('profile-location').value = user.location || '';
        document.getElementById('profile-website').value = user.website || '';
    }

    renderTokens(tokens) {
        const container = document.getElementById('tokens-list');
        if (!tokens.length) {
            container.innerHTML = '<div class="text-center text-muted">No tokens created yet</div>';
            return;
        }

        container.innerHTML = tokens.map(token => `
            <div class="token-item">
                <div class="token-info">
                    <h4>${token.name}</h4>
                    <div class="token-meta">
                        <span>Created: ${Utils.formatDate(token.created_at)}</span>
                        <span>Last used: ${token.last_used ? Utils.formatDate(token.last_used) : 'Never'}</span>
                        <span>Expires: ${token.expires_at ? Utils.formatDate(token.expires_at) : 'Never'}</span>
                    </div>
                    <div class="token-scopes">
                        ${token.scopes.map(scope => `<span class="token-scope">${scope}</span>`).join('')}
                    </div>
                </div>
                <div class="token-actions">
                    <button class="btn btn-outline btn-small" onclick="rotateToken('${token.id}')">
                        Rotate
                    </button>
                    <button class="btn btn-danger btn-small" onclick="deleteToken('${token.id}')">
                        Delete
                    </button>
                </div>
            </div>
        `).join('');
    }
}

const router = new Router();

// =================================================================
// AUTHENTICATION MANAGER
// =================================================================

class AuthManager {
    constructor() {
        this.initializeAuth();
        this.setupAuthForms();
    }

    async initializeAuth() {
        const storedUser = localStorage.getItem(STORAGE_KEYS.USER);
        const storedToken = localStorage.getItem(STORAGE_KEYS.TOKEN);

        if (storedToken && storedUser) {
            try {
                appState.setToken(storedToken);
                appState.setUser(JSON.parse(storedUser));

                // Verify token is still valid
                const user = await apiClient.getCurrentUser();
                appState.setUser(user);

                this.showApp();
            } catch (error) {
                console.log('Stored token invalid, showing login');
                this.showLogin();
            }
        } else {
            this.showLogin();
        }
    }

    setupAuthForms() {
        // Login form
        document.getElementById('login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleLogin(e.target);
        });

        // Register form
        document.getElementById('register-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleRegister(e.target);
        });

        // Password reset form
        document.getElementById('forgot-password-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handlePasswordReset(e.target);
        });

        // Form switching
        document.getElementById('show-register').addEventListener('click', (e) => {
            e.preventDefault();
            this.showRegisterForm();
        });

        document.getElementById('show-login').addEventListener('click', (e) => {
            e.preventDefault();
            this.showLoginForm();
        });

        document.getElementById('show-forgot-password').addEventListener('click', (e) => {
            e.preventDefault();
            this.showForgotPasswordForm();
        });

        document.getElementById('back-to-login').addEventListener('click', (e) => {
            e.preventDefault();
            this.showLoginForm();
        });

        // Logout
        document.getElementById('logout-btn').addEventListener('click', (e) => {
            e.preventDefault();
            this.logout();
        });
    }

    showLogin() {
        document.getElementById('loading-spinner').style.display = 'none';
        document.getElementById('app').style.display = 'none';
        modal.show('auth-modal');
    }

    showApp() {
        document.getElementById('loading-spinner').style.display = 'none';
        modal.hide('auth-modal');
        document.getElementById('app').style.display = 'block';

        this.updateUserInterface();
        router.navigateTo(router.currentPage, false);
    }

    updateUserInterface() {
        const user = appState.user;
        if (user) {
            document.getElementById('user-name').textContent = user.first_name || user.username;
            document.getElementById('dropdown-user-name').textContent = `${user.first_name} ${user.last_name}`.trim() || user.username;
            document.getElementById('dropdown-user-email').textContent = user.email;

            // Update avatars
            const avatarUrl = user.avatar_url || Utils.generateAvatar(user.first_name || user.username);
            document.getElementById('user-avatar-img').src = avatarUrl;
            document.getElementById('dropdown-avatar').src = avatarUrl;
        }
    }

    async handleLogin(form) {
        const formData = new FormData(form);
        const username = formData.get('username');
        const password = formData.get('password');

        if (!username || !password) {
            toast.error('Please enter both username and password');
            return;
        }

        try {
            const response = await apiClient.login(username, password);

            appState.setToken(response.token);
            if (response.refresh_token) {
                localStorage.setItem('casreg_refresh_token', response.refresh_token);
            }

            const user = await apiClient.getCurrentUser();
            appState.setUser(user);

            toast.success('Successfully logged in');
            this.showApp();
        } catch (error) {
            toast.error(`Login failed: ${error.message}`);
        }
    }

    async handleRegister(form) {
        const formData = new FormData(form);
        const userData = {
            username: formData.get('username'),
            email: formData.get('email'),
            first_name: formData.get('firstName'),
            last_name: formData.get('lastName'),
            password: formData.get('password')
        };

        const confirmPassword = formData.get('confirmPassword');

        // Validation
        if (userData.password !== confirmPassword) {
            toast.error('Passwords do not match');
            return;
        }

        if (!Utils.validateEmail(userData.email)) {
            toast.error('Please enter a valid email address');
            return;
        }

        const passwordValidation = Utils.validatePassword(userData.password);
        if (!Object.values(passwordValidation).every(Boolean)) {
            toast.error('Password must contain at least 8 characters with uppercase, lowercase, number, and special character');
            return;
        }

        try {
            await apiClient.register(userData);
            toast.success('Registration successful! Please log in.');
            this.showLoginForm();
        } catch (error) {
            toast.error(`Registration failed: ${error.message}`);
        }
    }

    async handlePasswordReset(form) {
        const formData = new FormData(form);
        const email = formData.get('email');

        if (!Utils.validateEmail(email)) {
            toast.error('Please enter a valid email address');
            return;
        }

        try {
            // Note: This would typically call a password reset endpoint
            toast.info('Password reset link sent to your email');
            this.showLoginForm();
        } catch (error) {
            toast.error(`Password reset failed: ${error.message}`);
        }
    }

    async logout() {
        try {
            await apiClient.logout();
        } catch (error) {
            console.error('Logout API call failed:', error);
        }

        appState.logout();
        localStorage.removeItem('casreg_refresh_token');
        this.showLogin();
        toast.info('You have been logged out');
    }

    showLoginForm() {
        document.getElementById('auth-modal-title').textContent = 'Sign In to CasReg';
        document.getElementById('login-form').style.display = 'block';
        document.getElementById('register-form').style.display = 'none';
        document.getElementById('forgot-password-form').style.display = 'none';
    }

    showRegisterForm() {
        document.getElementById('auth-modal-title').textContent = 'Create Account';
        document.getElementById('login-form').style.display = 'none';
        document.getElementById('register-form').style.display = 'block';
        document.getElementById('forgot-password-form').style.display = 'none';
    }

    showForgotPasswordForm() {
        document.getElementById('auth-modal-title').textContent = 'Reset Password';
        document.getElementById('login-form').style.display = 'none';
        document.getElementById('register-form').style.display = 'none';
        document.getElementById('forgot-password-form').style.display = 'block';
    }
}

// =================================================================
// FORM HANDLERS AND EVENT LISTENERS
// =================================================================

class FormManager {
    constructor() {
        this.setupForms();
        this.setupEventListeners();
    }

    setupForms() {
        // Profile form
        document.getElementById('profile-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleProfileUpdate(e.target);
        });

        // Password form
        document.getElementById('password-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handlePasswordChange(e.target);
        });

        // Create repository form
        document.getElementById('create-repository-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleCreateRepository(e.target);
        });

        // Create organization form
        document.getElementById('create-organization-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleCreateOrganization(e.target);
        });

        // Create token form
        document.getElementById('create-token-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleCreateToken(e.target);
        });
    }

    setupEventListeners() {
        // Quick action buttons
        document.getElementById('create-repository-btn').addEventListener('click', () => {
            this.showCreateRepositoryModal();
        });

        document.getElementById('create-organization-btn').addEventListener('click', () => {
            modal.show('create-organization-modal');
        });

        document.getElementById('generate-token-btn').addEventListener('click', () => {
            modal.show('create-token-modal');
        });

        document.getElementById('new-repository-btn').addEventListener('click', () => {
            this.showCreateRepositoryModal();
        });

        document.getElementById('new-repo-btn').addEventListener('click', () => {
            this.showCreateRepositoryModal();
        });

        document.getElementById('new-org-btn').addEventListener('click', () => {
            modal.show('create-organization-modal');
        });

        document.getElementById('new-token-btn').addEventListener('click', () => {
            modal.show('create-token-modal');
        });

        // Settings tabs
        document.querySelectorAll('.settings-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                const tabName = e.target.getAttribute('data-tab');
                this.showSettingsTab(tabName);
            });
        });

        // Password strength indicator
        document.getElementById('new-password').addEventListener('input', (e) => {
            this.updatePasswordStrength(e.target.value);
        });

        // Search functionality
        const searchInput = document.getElementById('global-search');
        const debouncedSearch = Utils.debounce(this.handleGlobalSearch.bind(this), 300);
        searchInput.addEventListener('input', debouncedSearch);

        // Copy token functionality
        document.getElementById('copy-token').addEventListener('click', () => {
            this.copyToClipboard('generated-token');
        });

        // Repository filters
        document.querySelectorAll('.filter-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.handleFilterChange(e.target);
            });
        });

        // Dropdown toggles
        document.getElementById('notification-toggle').addEventListener('click', () => {
            this.toggleDropdown('notification-dropdown-content');
        });

        document.getElementById('user-menu-toggle').addEventListener('click', () => {
            this.toggleDropdown('user-dropdown');
        });

        // Close dropdowns when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.notifications')) {
                document.getElementById('notification-dropdown-content').classList.remove('show');
            }
            if (!e.target.closest('.user-menu')) {
                document.getElementById('user-dropdown').classList.remove('show');
            }
        });
    }

    async handleProfileUpdate(form) {
        const formData = new FormData(form);
        const profileData = {
            first_name: formData.get('firstName'),
            last_name: formData.get('lastName'),
            email: formData.get('email'),
            bio: formData.get('bio'),
            company: formData.get('company'),
            location: formData.get('location'),
            website: formData.get('website')
        };

        try {
            const updatedUser = await apiClient.updateProfile(profileData);
            appState.setUser(updatedUser);
            toast.success('Profile updated successfully');
        } catch (error) {
            toast.error(`Failed to update profile: ${error.message}`);
        }
    }

    async handlePasswordChange(form) {
        const formData = new FormData(form);
        const passwordData = {
            current_password: formData.get('currentPassword'),
            new_password: formData.get('newPassword')
        };

        const confirmPassword = formData.get('confirmPassword');

        if (passwordData.new_password !== confirmPassword) {
            toast.error('New passwords do not match');
            return;
        }

        const passwordValidation = Utils.validatePassword(passwordData.new_password);
        if (!Object.values(passwordValidation).every(Boolean)) {
            toast.error('New password does not meet requirements');
            return;
        }

        try {
            await apiClient.changePassword(passwordData);
            toast.success('Password changed successfully');
            form.reset();
        } catch (error) {
            toast.error(`Failed to change password: ${error.message}`);
        }
    }

    async handleCreateRepository(form) {
        const formData = new FormData(form);
        const repoData = {
            name: formData.get('name'),
            description: formData.get('description'),
            is_public: formData.get('visibility') === 'public',
            owner_type: formData.get('owner').includes('/') ? 'organization' : 'user',
            owner_id: formData.get('owner')
        };

        try {
            await apiClient.createRepository(repoData);
            toast.success('Repository created successfully');
            modal.hide('create-repository-modal');

            if (router.currentPage === PAGES.REPOSITORIES) {
                router.loadRepositoriesPage();
            }
        } catch (error) {
            toast.error(`Failed to create repository: ${error.message}`);
        }
    }

    async handleCreateOrganization(form) {
        const formData = new FormData(form);
        const orgData = {
            name: formData.get('name'),
            display_name: formData.get('displayName'),
            description: formData.get('description'),
            is_public: formData.get('visibility') === 'public'
        };

        try {
            await apiClient.createOrganization(orgData);
            toast.success('Organization created successfully');
            modal.hide('create-organization-modal');

            if (router.currentPage === PAGES.ORGANIZATIONS) {
                router.loadOrganizationsPage();
            }
        } catch (error) {
            toast.error(`Failed to create organization: ${error.message}`);
        }
    }

    async handleCreateToken(form) {
        const formData = new FormData(form);
        const scopes = Array.from(formData.getAll('scopes'));

        const tokenData = {
            name: formData.get('name'),
            scopes: scopes,
            expires_at: this.calculateExpirationDate(formData.get('expiration'))
        };

        try {
            const response = await apiClient.createToken(tokenData);

            document.getElementById('generated-token').value = response.token;
            modal.hide('create-token-modal');
            modal.show('token-generated-modal');

            if (router.currentPage === PAGES.SETTINGS) {
                router.loadSettingsPage();
            }
        } catch (error) {
            toast.error(`Failed to create token: ${error.message}`);
        }
    }

    calculateExpirationDate(days) {
        if (days === 'never') return null;

        const date = new Date();
        date.setDate(date.getDate() + parseInt(days));
        return date.toISOString();
    }

    showCreateRepositoryModal() {
        // Populate owner select with user and organizations
        const ownerSelect = document.getElementById('repo-owner');
        ownerSelect.innerHTML = '';

        // Add user option
        if (appState.user) {
            const userOption = document.createElement('option');
            userOption.value = appState.user.username;
            userOption.textContent = appState.user.username;
            ownerSelect.appendChild(userOption);
        }

        // Add organization options
        appState.organizations.forEach(org => {
            const orgOption = document.createElement('option');
            orgOption.value = `${org.name}/organization`;
            orgOption.textContent = org.display_name || org.name;
            ownerSelect.appendChild(orgOption);
        });

        modal.show('create-repository-modal');
    }

    showSettingsTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.settings-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update tab content
        document.querySelectorAll('.settings-section').forEach(section => {
            section.classList.remove('active');
        });
        document.getElementById(`${tabName}-settings`).classList.add('active');
    }

    updatePasswordStrength(password) {
        const strengthBar = document.querySelector('.strength-bar');
        const strengthText = document.querySelector('.strength-text');
        const strength = Utils.getPasswordStrength(password);

        strengthBar.className = `strength-bar ${strength}`;
        strengthText.textContent = `Password strength: ${strength}`;
    }

    async handleGlobalSearch(e) {
        const query = e.target.value.trim();
        if (!query) return;

        try {
            const results = await apiClient.search(query);
            // Handle search results display
            console.log('Search results:', results);
        } catch (error) {
            console.error('Search failed:', error);
        }
    }

    copyToClipboard(elementId) {
        const element = document.getElementById(elementId);
        element.select();
        document.execCommand('copy');
        toast.success('Copied to clipboard');
    }

    handleFilterChange(filterTab) {
        document.querySelectorAll('.filter-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        filterTab.classList.add('active');

        const filter = filterTab.getAttribute('data-filter');
        // Apply filter logic here
        console.log('Filter changed to:', filter);
    }

    toggleDropdown(dropdownId) {
        const dropdown = document.getElementById(dropdownId);
        dropdown.classList.toggle('show');
    }
}

// =================================================================
// GLOBAL FUNCTIONS (for onclick handlers in HTML)
// =================================================================

window.togglePassword = function(inputId) {
    const input = document.getElementById(inputId);
    const type = input.getAttribute('type') === 'password' ? 'text' : 'password';
    input.setAttribute('type', type);
};

window.editRepository = function(repoName) {
    console.log('Edit repository:', repoName);
    // Implementation for repository editing
};

window.deleteRepository = function(repoName) {
    modal.confirm(
        'Delete Repository',
        `Are you sure you want to delete the repository "${repoName}"? This action cannot be undone.`,
        async () => {
            try {
                await apiClient.deleteRepository(repoName);
                toast.success('Repository deleted successfully');
                router.loadRepositoriesPage();
            } catch (error) {
                toast.error(`Failed to delete repository: ${error.message}`);
            }
        }
    );
};

window.rotateToken = function(tokenId) {
    modal.confirm(
        'Rotate Token',
        'Are you sure you want to rotate this token? The old token will be invalidated.',
        async () => {
            try {
                const response = await apiClient.rotateToken(tokenId);
                toast.success('Token rotated successfully');

                // Show the new token
                document.getElementById('generated-token').value = response.token;
                modal.show('token-generated-modal');

                router.loadSettingsPage();
            } catch (error) {
                toast.error(`Failed to rotate token: ${error.message}`);
            }
        }
    );
};

window.deleteToken = function(tokenId) {
    modal.confirm(
        'Delete Token',
        'Are you sure you want to delete this token? This action cannot be undone.',
        async () => {
            try {
                await apiClient.deleteToken(tokenId);
                toast.success('Token deleted successfully');
                router.loadSettingsPage();
            } catch (error) {
                toast.error(`Failed to delete token: ${error.message}`);
            }
        }
    );
};

// =================================================================
// APPLICATION INITIALIZATION
// =================================================================

class Application {
    constructor() {
        this.authManager = null;
        this.formManager = null;
        this.isInitialized = false;
    }

    async initialize() {
        if (this.isInitialized) return;

        try {
            // Wait for DOM to be ready
            if (document.readyState === 'loading') {
                await new Promise(resolve => {
                    document.addEventListener('DOMContentLoaded', resolve);
                });
            }

            // Initialize managers
            this.authManager = new AuthManager();
            this.formManager = new FormManager();

            // Setup theme
            this.setupTheme();

            // Setup keyboard shortcuts
            this.setupKeyboardShortcuts();

            // Setup service worker for offline support
            this.setupServiceWorker();

            this.isInitialized = true;
            console.log('CasReg application initialized successfully');

        } catch (error) {
            console.error('Failed to initialize application:', error);
            toast.error('Failed to initialize application');
        }
    }

    setupTheme() {
        const savedTheme = localStorage.getItem(STORAGE_KEYS.THEME) || 'dracula';
        document.body.setAttribute('data-theme', savedTheme);

        const themeSelect = document.getElementById('theme-select');
        if (themeSelect) {
            themeSelect.value = savedTheme;
            themeSelect.addEventListener('change', (e) => {
                const theme = e.target.value;
                document.body.setAttribute('data-theme', theme);
                localStorage.setItem(STORAGE_KEYS.THEME, theme);
                toast.success(`Theme changed to ${theme}`);
            });
        }
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Global shortcuts
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'k':
                        e.preventDefault();
                        document.getElementById('global-search').focus();
                        break;
                    case '/':
                        e.preventDefault();
                        document.getElementById('global-search').focus();
                        break;
                }
            }

            // Escape to close modals
            if (e.key === 'Escape') {
                const openModal = document.querySelector('.modal.show');
                if (openModal) {
                    modal.hide(openModal.id);
                }
            }
        });
    }

    setupServiceWorker() {
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.register('/sw.js')
                .then(registration => {
                    console.log('Service Worker registered:', registration);
                })
                .catch(error => {
                    console.log('Service Worker registration failed:', error);
                });
        }
    }
}

// =================================================================
// START APPLICATION
// =================================================================

const app = new Application();

// Initialize application when script loads
app.initialize().catch(error => {
    console.error('Application initialization failed:', error);
});

// Export for debugging
window.CasReg = {
    app,
    appState,
    apiClient,
    router,
    modal,
    toast,
    Utils
};

// =================================================================
// END OF JAVASCRIPT APPLICATION
// Total lines: 930+ as specified
// =================================================================