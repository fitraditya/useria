// Shared helpers for every Useria page. Loaded before page-specific scripts.

const Useria = {
  TOKEN_KEY: 'useria_token',

  getToken() {
    return localStorage.getItem(this.TOKEN_KEY);
  },

  setToken(token) {
    localStorage.setItem(this.TOKEN_KEY, token);
  },

  clearToken() {
    localStorage.removeItem(this.TOKEN_KEY);
  },

  // Decodes the JWT payload for UI decisions only (expiry, company_id
  // presence, role display). The server is the source of truth; this never
  // gates access on its own.
  decodeToken(token) {
    token = token || this.getToken();
    if (!token) return null;
    try {
      const payload = token.split('.')[1];
      const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
      return JSON.parse(json);
    } catch (e) {
      return null;
    }
  },

  isExpired(claims) {
    if (!claims || !claims.exp) return true;
    return Date.now() >= claims.exp * 1000;
  },

  logout() {
    this.clearToken();
    window.location.href = '/login';
  },

  // Redirects to /login if there is no valid token. Call at the top of
  // any protected page. Returns the decoded claims when authenticated.
  requireAuth() {
    const claims = this.decodeToken();
    if (!claims || this.isExpired(claims)) {
      this.clearToken();
      window.location.href = '/login';
      return null;
    }
    return claims;
  },

  // Redirects to /select-company if the current token has no company
  // context selected yet.
  requireCompany() {
    const claims = this.requireAuth();
    if (claims && !claims.company_id) {
      window.location.href = '/select-company';
      return null;
    }
    return claims;
  },

  async apiFetch(path, options = {}) {
    const token = this.getToken();
    const headers = Object.assign(
      { 'Content-Type': 'application/json' },
      options.headers || {},
    );
    if (token) {
      headers['Authorization'] = 'Bearer ' + token;
    }

    const res = await fetch('/api' + path, Object.assign({}, options, { headers }));
    if (res.status === 401) {
      this.clearToken();
      window.location.href = '/login';
      throw new Error('unauthenticated');
    }

    const body = await res.json().catch(() => ({}));
    if (!res.ok || body.success === false) {
      throw new Error(body.error || ('request failed: ' + res.status));
    }
    return body.data;
  },

  showError(el, err) {
    if (!el) return;
    el.textContent = err.message || String(err);
    el.classList.remove('hidden');
  },

  hideError(el) {
    if (!el) return;
    el.classList.add('hidden');
  },

  formatDate(iso) {
    if (!iso || iso.startsWith('0001-01-01')) return '—';
    try {
      return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    } catch (e) {
      return iso;
    }
  },

  initials(firstName, lastName, email) {
    const f = (firstName || '').trim();
    const l = (lastName || '').trim();
    if (f || l) return ((f[0] || '') + (l[0] || '')).toUpperCase();
    return (email || '?').slice(0, 2).toUpperCase();
  },

  // Populates the header's avatar/name/email and wires sign-out, on any
  // dashboard-family page that includes the shared header partial.
  async initHeader() {
    const avatarEl = document.getElementById('user-avatar');
    const nameEl = document.getElementById('user-name');
    const signoutBtn = document.getElementById('header-signout-btn');
    if (!avatarEl && !nameEl && !signoutBtn) return;

    if (signoutBtn) signoutBtn.addEventListener('click', () => this.logout());

    try {
      const user = await this.apiFetch('/profile');
      const name = [user.first_name, user.last_name].filter(Boolean).join(' ') || user.email;
      if (avatarEl) avatarEl.textContent = this.initials(user.first_name, user.last_name, user.email);
      if (nameEl) nameEl.textContent = name;
      const nameFullEl = document.getElementById('user-name-full');
      const emailEl = document.getElementById('user-email');
      if (nameFullEl) nameFullEl.textContent = name;
      if (emailEl) emailEl.textContent = user.email;
    } catch (e) {
      // requireAuth()/requireCompany() on the page itself handles redirects;
      // nothing further to do here if this fails.
    }
  },

  // Shows any [data-require-scope] element whose listed scope (or any of a
  // comma-separated list) is present on the current token. Purely cosmetic —
  // the API enforces the real access control.
  applyScopeVisibility() {
    const claims = this.decodeToken();
    const scopes = (claims && claims.scopes) || [];
    document.querySelectorAll('[data-require-scope]').forEach((el) => {
      const required = el.getAttribute('data-require-scope').split(',').map((s) => s.trim());
      if (required.some((s) => scopes.includes(s))) {
        el.classList.remove('hidden');
      }
    });
  },
};

document.addEventListener('DOMContentLoaded', () => {
  Useria.initHeader();
  Useria.applyScopeVisibility();
});
