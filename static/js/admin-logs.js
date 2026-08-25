const AdminLogs = {
  companies: [],
  selectedCompanyId: '',
  refreshTimer: null,

  actionLabels: {
    'user.register': 'Registered',
    'user.login': 'Logged in',
    'user.password_reset_requested': 'Requested password reset',
    'user.password_reset_completed': 'Completed password reset',
    'company.create': 'Created company',
    'company.update': 'Updated company',
    'company.suspend': 'Suspended company',
    'company.activate': 'Activated company',
    'company.delete': 'Deleted company',
    'member.invite': 'Invited member',
    'member.role_update': 'Changed member role',
    'member.remove': 'Removed member',
  },

  actionColor(action) {
    if (action.startsWith('user.login') || action.startsWith('user.register')) return 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500';
    if (action.includes('delete') || action.includes('remove') || action.includes('suspend')) return 'bg-error-50 text-error-700 dark:bg-error-500/15 dark:text-error-500';
    if (action.includes('password_reset')) return 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-400';
    return 'bg-gray-100 text-gray-600 dark:bg-white/5 dark:text-gray-400';
  },

  formatDateTime(iso) {
    if (!iso || iso.startsWith('0001-01-01')) return '—';
    try {
      return new Date(iso).toLocaleString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' });
    } catch (e) {
      return iso;
    }
  },

  async init() {
    const claims = Useria.requireCompany();
    if (!claims) return;

    await this.loadCompanies();
    this.initCompanyDropdown();

    let debounceTimer;
    const debouncedLoad = () => {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => this.load(), 300);
    };
    document.getElementById('filter-name').addEventListener('input', debouncedLoad);
    document.getElementById('filter-email').addEventListener('input', debouncedLoad);
    document.getElementById('filter-action').addEventListener('change', () => this.load());
    document.getElementById('filter-date-from').addEventListener('change', () => this.load());
    document.getElementById('filter-date-to').addEventListener('change', () => this.load());

    const intervalSelect = document.getElementById('refresh-interval');
    intervalSelect.addEventListener('change', () => this.setupAutoRefresh(intervalSelect.value));
    this.setupAutoRefresh(intervalSelect.value);

    await this.load();
  },

  setupAutoRefresh(seconds) {
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
      this.refreshTimer = null;
    }
    const n = Number(seconds);
    if (n > 0) {
      this.refreshTimer = setInterval(() => this.load(), n * 1000);
    }
  },

  async loadCompanies() {
    try {
      this.companies = await Useria.apiFetch('/admin/companies');
    } catch (err) {
      Useria.showError(document.getElementById('error'), err);
    }
  },

  initCompanyDropdown() {
    const input = document.getElementById('filter-company-input');
    const dropdown = document.getElementById('filter-company-dropdown');

    const renderOptions = (filterText) => {
      const q = (filterText || '').trim().toLowerCase();
      const matches = (this.companies || []).filter((c) => c.name.toLowerCase().includes(q));

      dropdown.innerHTML = '';

      const allOption = document.createElement('button');
      allOption.type = 'button';
      allOption.className = 'block w-full px-4 py-2 text-left text-sm text-gray-500 hover:bg-gray-50 dark:text-gray-400 dark:hover:bg-white/5';
      allOption.textContent = 'All companies';
      allOption.addEventListener('click', () => {
        this.selectedCompanyId = '';
        input.value = '';
        dropdown.classList.add('hidden');
        this.load();
      });
      dropdown.appendChild(allOption);

      matches.forEach((c) => {
        const opt = document.createElement('button');
        opt.type = 'button';
        opt.className = 'block w-full px-4 py-2 text-left text-sm text-gray-800 hover:bg-gray-50 dark:text-white/90 dark:hover:bg-white/5';
        opt.textContent = c.name;
        opt.addEventListener('click', () => {
          this.selectedCompanyId = c.id;
          input.value = c.name;
          dropdown.classList.add('hidden');
          this.load();
        });
        dropdown.appendChild(opt);
      });

      if (matches.length === 0) {
        const none = document.createElement('p');
        none.className = 'px-4 py-2 text-sm text-gray-400';
        none.textContent = 'No matching companies';
        dropdown.appendChild(none);
      }
    };

    input.addEventListener('focus', () => {
      renderOptions(input.value);
      dropdown.classList.remove('hidden');
    });
    input.addEventListener('input', () => {
      if (this.selectedCompanyId) this.selectedCompanyId = '';
      renderOptions(input.value);
      dropdown.classList.remove('hidden');
    });
    document.addEventListener('click', (e) => {
      if (!input.contains(e.target) && !dropdown.contains(e.target)) {
        dropdown.classList.add('hidden');
      }
    });
  },

  async load() {
    const errorEl = document.getElementById('error');
    const unsupportedEl = document.getElementById('unsupported');
    const tbody = document.getElementById('log-rows');
    const emptyEl = document.getElementById('log-empty');
    const countEl = document.getElementById('log-count');

    try {
      const params = new URLSearchParams();
      const name = document.getElementById('filter-name').value.trim();
      const email = document.getElementById('filter-email').value.trim();
      const action = document.getElementById('filter-action').value;
      const dateFrom = document.getElementById('filter-date-from').value;
      const dateTo = document.getElementById('filter-date-to').value;
      if (this.selectedCompanyId) params.set('company_id', this.selectedCompanyId);
      if (name) params.set('name', name);
      if (email) params.set('email', email);
      if (action) params.set('action', action);
      if (dateFrom) params.set('date_from', dateFrom);
      if (dateTo) params.set('date_to', dateTo);

      const result = await Useria.apiFetch('/admin/activity?' + params.toString());
      const logs = result.logs || [];

      unsupportedEl.classList.toggle('hidden', result.supported !== false);
      emptyEl.classList.toggle('hidden', logs.length > 0);
      countEl.textContent = logs.length > 0 ? `Showing ${logs.length} most recent` : '';

      tbody.innerHTML = '';
      logs.forEach((entry) => {
        const actorName = [entry.actor_first_name, entry.actor_last_name].filter(Boolean).join(' ') || entry.actor_email;
        const label = this.actionLabels[entry.action] || entry.action;
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${this.formatDateTime(entry.created_at)}</span></td>
          <td class="px-5 py-4 sm:px-6">
            <span class="block font-medium text-gray-800 text-theme-sm dark:text-white/90">${actorName}</span>
            <span class="block text-gray-500 text-theme-xs dark:text-gray-400">${entry.actor_email}</span>
          </td>
          <td class="px-5 py-4 sm:px-6"><p class="inline-block rounded-full px-2 py-0.5 text-theme-xs font-medium ${this.actionColor(entry.action)}">${label}</p></td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${entry.company_name || '—'}</span></td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-xs text-gray-500 dark:text-gray-400">${entry.metadata || ''}</span></td>
        `;
        tbody.appendChild(tr);
      });
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },
};
