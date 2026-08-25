const Companies = {
  editingId: null,

  closeModal() {
    if (window.Alpine) Alpine.$data(document.body).isCompanyModalOpen = false;
  },

  async init() {
    const claims = Useria.requireCompany();
    if (!claims) return;

    this.canWrite = (claims.scopes || []).includes('companies:write');

    document.getElementById('add-company-btn').addEventListener('click', () => {
      this.editingId = null;
      document.getElementById('company-modal-title').textContent = 'Add company';
      document.getElementById('company-form').reset();
    });
    document.getElementById('company-form').addEventListener('submit', (e) => this.submit(e));

    let debounceTimer;
    const debouncedLoad = () => {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => this.load(), 300);
    };
    document.getElementById('filter-name').addEventListener('input', debouncedLoad);
    document.getElementById('filter-admin-email').addEventListener('input', debouncedLoad);
    document.getElementById('filter-status').addEventListener('change', () => this.load());

    await this.load();
  },

  statusBadge(status) {
    const colors = {
      active: 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500',
      suspended: 'bg-error-50 text-error-700 dark:bg-error-500/15 dark:text-error-500',
      inactive: 'bg-gray-100 text-gray-600 dark:bg-white/5 dark:text-gray-400',
    };
    const cls = colors[status] || colors.inactive;
    return `<p class="inline-block rounded-full px-2 py-0.5 text-theme-xs font-medium capitalize ${cls}">${status}</p>`;
  },

  async load() {
    const errorEl = document.getElementById('error');
    const tbody = document.getElementById('company-rows');
    const emptyEl = document.getElementById('company-empty');
    try {
      const params = new URLSearchParams();
      const name = document.getElementById('filter-name').value.trim();
      const adminEmail = document.getElementById('filter-admin-email').value.trim();
      const status = document.getElementById('filter-status').value;
      if (name) params.set('name', name);
      if (adminEmail) params.set('admin_email', adminEmail);
      if (status) params.set('status', status);
      const qs = params.toString();

      const companies = await Useria.apiFetch('/admin/companies' + (qs ? '?' + qs : ''));
      tbody.innerHTML = '';
      emptyEl.classList.toggle('hidden', (companies || []).length > 0);

      (companies || []).forEach((c) => {
        const tr = document.createElement('tr');
        const suspendLabel = c.status === 'suspended' ? 'Activate' : 'Suspend';

        const actions = this.canWrite
          ? `<button data-id="${c.id}" class="edit-btn text-theme-sm font-medium text-brand-500 hover:text-brand-600 mr-4">Edit</button>
             <button data-id="${c.id}" data-status="${c.status}" class="suspend-btn text-theme-sm font-medium text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">${suspendLabel}</button>`
          : '';

        tr.innerHTML = `
          <td class="px-5 py-4 sm:px-6">
            <span class="block font-medium text-gray-800 text-theme-sm dark:text-white/90">${c.name}</span>
            <span class="block text-gray-500 text-theme-xs dark:text-gray-400">${c.slug}</span>
          </td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${c.admin_email || '—'}</span></td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400 capitalize">${c.plan}</span></td>
          <td class="px-5 py-4 sm:px-6">${this.statusBadge(c.status)}</td>
          <td class="px-5 py-4 sm:px-6"><span class="text-theme-sm text-gray-500 dark:text-gray-400">${Useria.formatDate(c.created_at)}</span></td>
          <td class="px-5 py-4 sm:px-6 text-right">${actions}</td>
        `;
        tbody.appendChild(tr);
        this._company = this._company || {};
        this._company[c.id] = c;
      });

      tbody.querySelectorAll('.edit-btn').forEach((btn) => {
        btn.addEventListener('click', () => this.openEdit(btn.dataset.id));
      });
      tbody.querySelectorAll('.suspend-btn').forEach((btn) => {
        btn.addEventListener('click', () => this.toggleSuspend(btn.dataset.id, btn.dataset.status));
      });
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },

  openEdit(id) {
    const c = this._company && this._company[id];
    if (!c) return;
    this.editingId = id;
    document.getElementById('company-modal-title').textContent = 'Edit company';
    const form = document.getElementById('company-form');
    form.name.value = c.name || '';
    form.plan.value = c.plan || 'free';
    form.website.value = c.website || '';
    form.description.value = c.description || '';
    if (window.Alpine) Alpine.$data(document.body).isCompanyModalOpen = true;
  },

  async toggleSuspend(id, currentStatus) {
    const nextStatus = currentStatus === 'suspended' ? 'active' : 'suspended';
    const verb = nextStatus === 'suspended' ? 'suspend' : 'reactivate';
    if (!confirm(`${verb.charAt(0).toUpperCase() + verb.slice(1)} this company?`)) return;

    const errorEl = document.getElementById('error');
    Useria.hideError(errorEl);
    try {
      await Useria.apiFetch(`/admin/companies/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ status: nextStatus }),
      });
      await this.load();
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },

  async submit(e) {
    e.preventDefault();
    const errorEl = document.getElementById('company-modal-error');
    Useria.hideError(errorEl);
    const form = document.getElementById('company-form');

    try {
      if (this.editingId) {
        await Useria.apiFetch(`/admin/companies/${this.editingId}`, {
          method: 'PUT',
          body: JSON.stringify({
            name: form.name.value,
            plan: form.plan.value,
            website: form.website.value,
            description: form.description.value,
          }),
        });
      } else {
        await Useria.apiFetch('/admin/companies', {
          method: 'POST',
          body: JSON.stringify({
            name: form.name.value,
            plan: form.plan.value,
            website: form.website.value,
            description: form.description.value,
          }),
        });
      }
      this.closeModal();
      form.reset();
      this.editingId = null;
      await this.load();
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },
};
