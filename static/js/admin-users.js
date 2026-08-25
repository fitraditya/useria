const AdminUsers = {
  companies: [],
  selectedCompanyId: '',

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

    await this.load();
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

  statusBadge(status) {
    const colors = {
      active: 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500',
      invited: 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-400',
      inactive: 'bg-gray-100 text-gray-600 dark:bg-white/5 dark:text-gray-400',
    };
    const cls = colors[status] || colors.inactive;
    return `<p class="inline-block rounded-full px-2 py-0.5 text-theme-xs font-medium ${cls}">${status}</p>`;
  },

  async load() {
    const errorEl = document.getElementById('error');
    const tbody = document.getElementById('user-rows');
    const hintEl = document.getElementById('user-hint');
    const emptyEl = document.getElementById('user-empty');

    const name = document.getElementById('filter-name').value.trim();
    const email = document.getElementById('filter-email').value.trim();
    const hasFilter = Boolean(this.selectedCompanyId || name || email);

    tbody.innerHTML = '';
    if (!hasFilter) {
      hintEl.classList.remove('hidden');
      emptyEl.classList.add('hidden');
      return;
    }
    hintEl.classList.add('hidden');

    try {
      const params = new URLSearchParams();
      if (this.selectedCompanyId) params.set('company_id', this.selectedCompanyId);
      if (name) params.set('name', name);
      if (email) params.set('email', email);

      const users = await Useria.apiFetch('/admin/users?' + params.toString());
      emptyEl.classList.toggle('hidden', (users || []).length > 0);

      (users || []).forEach((m) => {
        const displayName = [m.first_name, m.last_name].filter(Boolean).join(' ') || m.email;
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td class="px-5 py-4 sm:px-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-full bg-brand-500 text-xs font-semibold text-white">${Useria.initials(m.first_name, m.last_name, m.email)}</div>
              <div>
                <span class="block font-medium text-gray-800 text-theme-sm dark:text-white/90">${displayName}</span>
                <span class="block text-gray-500 text-theme-xs dark:text-gray-400">${m.email}</span>
              </div>
            </div>
          </td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${m.company_name}</span></td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${m.role_name}</span></td>
          <td class="px-5 py-4 sm:px-6">${this.statusBadge(m.status)}</td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${Useria.formatDate(m.joined_at)}</span></td>
        `;
        tbody.appendChild(tr);
      });
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },
};
