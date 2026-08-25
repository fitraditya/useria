const Auth = {
  initLogin() {
    const form = document.getElementById('login-form');
    const errorEl = document.getElementById('error');
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      Useria.hideError(errorEl);
      try {
        const data = await Useria.apiFetch('/auth/login', {
          method: 'POST',
          body: JSON.stringify({
            email: form.email.value,
            password: form.password.value,
          }),
        });
        const claims = Useria.decodeToken(data.token);
        if (window.USERIA_APP === 'admin') {
          if (claims && claims.company_id) {
            Useria.setToken(data.token);
            window.location.href = '/admin/companies';
          } else {
            throw new Error('This account does not have platform admin access.');
          }
        } else {
          Useria.setToken(data.token);
          window.location.href = claims && claims.company_id ? '/dashboard' : '/select-company';
        }
      } catch (err) {
        Useria.showError(errorEl, err);
      }
    });
  },

  initRegister() {
    const form = document.getElementById('register-form');
    const errorEl = document.getElementById('error');
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      Useria.hideError(errorEl);
      try {
        const data = await Useria.apiFetch('/auth/register', {
          method: 'POST',
          body: JSON.stringify({
            email: form.email.value,
            password: form.password.value,
            first_name: form.first_name.value,
            last_name: form.last_name.value,
            company_name: form.company_name.value,
          }),
        });
        Useria.setToken(data.token);
        window.location.href = '/select-company';
      } catch (err) {
        Useria.showError(errorEl, err);
      }
    });
  },

  async initSelectCompany() {
    if (!Useria.requireAuth()) return;

    const errorEl = document.getElementById('error');
    const companyList = document.getElementById('company-list');
    const companyEmpty = document.getElementById('company-empty');
    const inviteSection = document.getElementById('invite-section');
    const inviteList = document.getElementById('invite-list');

    document.getElementById('logout-btn').addEventListener('click', () => Useria.logout());

    const selectCompany = async (companyId) => {
      Useria.hideError(errorEl);
      try {
        const data = await Useria.apiFetch('/auth/select-company', {
          method: 'POST',
          body: JSON.stringify({ company_id: companyId }),
        });
        Useria.setToken(data.token);
        window.location.href = '/dashboard';
      } catch (err) {
        Useria.showError(errorEl, err);
      }
    };

    const acceptInvite = async (companyId, btn) => {
      Useria.hideError(errorEl);
      btn.disabled = true;
      btn.textContent = 'Accepting…';
      try {
        await Useria.apiFetch('/invitations/accept', {
          method: 'POST',
          body: JSON.stringify({ company_id: companyId }),
        });
        await load();
      } catch (err) {
        Useria.showError(errorEl, err);
        btn.disabled = false;
        btn.textContent = 'Accept';
      }
    };

    const load = async () => {
      try {
        const [companies, invites] = await Promise.all([
          Useria.apiFetch('/auth/select-company'),
          Useria.apiFetch('/invitations'),
        ]);

        companyList.innerHTML = '';
        if (!companies || companies.length === 0) {
          companyEmpty.classList.remove('hidden');
        } else {
          companyEmpty.classList.add('hidden');
          companies.forEach((c) => {
            const row = document.createElement('button');
            row.type = 'button';
            row.className = 'w-full flex items-center justify-between rounded-lg border border-gray-200 dark:border-gray-700 px-4 py-3 text-left hover:border-brand-300 hover:bg-brand-50 dark:hover:bg-brand-500/10 transition-colors';
            row.innerHTML = `
              <span>
                <span class="block font-medium text-gray-800 dark:text-white/90">${c.company_name}</span>
                <span class="block text-xs text-gray-500 dark:text-gray-400">${c.role_name}</span>
              </span>
              <span class="text-brand-500 text-sm">Continue &rarr;</span>
            `;
            row.addEventListener('click', () => selectCompany(c.company_id));
            companyList.appendChild(row);
          });
        }

        inviteList.innerHTML = '';
        if (invites && invites.length > 0) {
          inviteSection.classList.remove('hidden');
          invites.forEach((inv) => {
            const row = document.createElement('div');
            row.className = 'flex items-center justify-between rounded-lg border border-warning-200 dark:border-warning-800 bg-warning-50 dark:bg-warning-500/15 px-4 py-3';
            row.innerHTML = `
              <span>
                <span class="block font-medium text-gray-800 dark:text-white/90">${inv.company_name}</span>
                <span class="block text-xs text-gray-500 dark:text-gray-400">Invited as ${inv.role_name}</span>
              </span>
            `;
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.textContent = 'Accept';
            btn.className = 'ml-3 shrink-0 bg-warning-500 hover:bg-warning-600 text-white text-sm font-medium rounded-lg px-3 py-1.5';
            btn.addEventListener('click', () => acceptInvite(inv.company_id, btn));
            row.appendChild(btn);
            inviteList.appendChild(row);
          });
        } else {
          inviteSection.classList.add('hidden');
        }
      } catch (err) {
        Useria.showError(errorEl, err);
      }
    };

    load();
  },

  initForgotPassword() {
    const form = document.getElementById('forgot-password-form');
    const errorEl = document.getElementById('error');
    const successEl = document.getElementById('success');
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      Useria.hideError(errorEl);
      successEl.classList.add('hidden');
      try {
        await Useria.apiFetch('/auth/forgot-password', {
          method: 'POST',
          body: JSON.stringify({ email: form.email.value }),
        });
        successEl.classList.remove('hidden');
        form.reset();
      } catch (err) {
        Useria.showError(errorEl, err);
      }
    });
  },

  initResetPassword() {
    const form = document.getElementById('reset-password-form');
    const errorEl = document.getElementById('error');
    const noTokenEl = document.getElementById('no-token');
    const token = new URLSearchParams(window.location.search).get('token');

    if (!token) {
      noTokenEl.classList.remove('hidden');
      form.querySelector('button[type=submit]').disabled = true;
      return;
    }

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      Useria.hideError(errorEl);
      if (form.password.value !== form.password_confirm.value) {
        Useria.showError(errorEl, new Error('Passwords do not match.'));
        return;
      }
      try {
        await Useria.apiFetch('/auth/reset-password', {
          method: 'POST',
          body: JSON.stringify({ token, password: form.password.value }),
        });
        window.location.href = '/login';
      } catch (err) {
        Useria.showError(errorEl, err);
      }
    });
  },
};
