const Members = {
  roles: [],

  closeInviteModal() {
    if (window.Alpine) Alpine.$data(document.body).isMemberInviteModal = false;
  },

  async init() {
    const claims = Useria.requireCompany();
    if (!claims) return;

    this.claims = claims;
    this.canInvite = (claims.scopes || []).includes('members:create');
    this.canWrite = (claims.scopes || []).includes('members:write');
    this.canDelete = (claims.scopes || []).includes('members:delete');

    if (this.canInvite) {
      document.getElementById('invite-btn').classList.remove('hidden');
    }

    document.getElementById('invite-form').addEventListener('submit', (e) => this.submitInvite(e));

    await this.loadRoles();
    await this.loadMembers();
  },

  async loadRoles() {
    try {
      this.roles = await Useria.apiFetch('/company/roles');
      const select = document.getElementById('invite-role');
      select.innerHTML = this.roles.map((r) => `<option value="${r.id}">${r.name}</option>`).join('');
    } catch (err) {
      Useria.showError(document.getElementById('error'), err);
    }
  },

  roleOptionsHTML(currentRoleId) {
    return this.roles
      .map((r) => `<option value="${r.id}" ${r.id === currentRoleId ? 'selected' : ''}>${r.name}</option>`)
      .join('');
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

  async loadMembers() {
    const errorEl = document.getElementById('error');
    const tbody = document.getElementById('member-rows');
    try {
      const members = await Useria.apiFetch('/company/members');
      tbody.innerHTML = '';

      (members || []).forEach((m) => {
        const name = [m.first_name, m.last_name].filter(Boolean).join(' ') || m.email;
        const tr = document.createElement('tr');

        // Roles not in the assignable list (e.g. platform SuperAdmin) are
        // not editable here — the backend rejects changes to them, and a
        // <select> with no matching <option> would silently show the wrong
        // role and risk submitting an unintended change.
        const isAssignable = this.roles.some((r) => r.id === m.role_id);
        const roleCell = this.canWrite && isAssignable
          ? `<select data-id="${m.id}" class="role-select rounded-lg border border-gray-300 dark:border-gray-700 dark:bg-gray-900 px-2 py-1.5 text-sm text-gray-700 dark:text-white/90">${this.roleOptionsHTML(m.role_id)}</select>`
          : `<span class="text-theme-sm text-gray-500 dark:text-gray-400">${m.role_name}</span>`;

        const removeCell = this.canDelete && isAssignable
          ? `<button data-id="${m.id}" class="remove-btn text-theme-sm font-medium text-error-500 hover:text-error-600">Remove</button>`
          : '';

        tr.innerHTML = `
          <td class="px-5 py-4 sm:px-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-full bg-brand-500 text-xs font-semibold text-white">${Useria.initials(m.first_name, m.last_name, m.email)}</div>
              <div>
                <span class="block font-medium text-gray-800 text-theme-sm dark:text-white/90">${name}</span>
                <span class="block text-gray-500 text-theme-xs dark:text-gray-400">${m.email}</span>
              </div>
            </div>
          </td>
          <td class="px-5 py-4 sm:px-6">${roleCell}</td>
          <td class="px-5 py-4 sm:px-6">${this.statusBadge(m.status)}</td>
          <td class="px-5 py-4 sm:px-6 text-right">${removeCell}</td>
        `;
        tbody.appendChild(tr);
      });

      tbody.querySelectorAll('.role-select').forEach((sel) => {
        sel.addEventListener('change', () => this.updateRole(sel.dataset.id, sel.value));
      });
      tbody.querySelectorAll('.remove-btn').forEach((btn) => {
        btn.addEventListener('click', () => this.removeMember(btn.dataset.id));
      });
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },

  async updateRole(memberId, roleId) {
    const errorEl = document.getElementById('error');
    Useria.hideError(errorEl);
    try {
      await Useria.apiFetch(`/company/members/${memberId}`, {
        method: 'PUT',
        body: JSON.stringify({ role_id: roleId }),
      });
      await this.loadMembers();
    } catch (err) {
      Useria.showError(errorEl, err);
      await this.loadMembers();
    }
  },

  async removeMember(memberId) {
    if (!confirm('Remove this member from the company?')) return;
    const errorEl = document.getElementById('error');
    Useria.hideError(errorEl);
    try {
      await Useria.apiFetch(`/company/members/${memberId}`, { method: 'DELETE' });
      await this.loadMembers();
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },

  async submitInvite(e) {
    e.preventDefault();
    const errorEl = document.getElementById('invite-error');
    Useria.hideError(errorEl);
    try {
      await Useria.apiFetch('/company/members/invite', {
        method: 'POST',
        body: JSON.stringify({
          email: document.getElementById('invite-email').value,
          role_id: document.getElementById('invite-role').value,
        }),
      });
      this.closeInviteModal();
      document.getElementById('invite-form').reset();
      await this.loadMembers();
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },
};
