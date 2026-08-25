const Dashboard = {
  async init() {
    const claims = Useria.requireCompany();
    if (!claims) return;

    const errorEl = document.getElementById('error');
    try {
      const [profile, companies] = await Promise.all([
        Useria.apiFetch('/profile'),
        Useria.apiFetch('/auth/select-company'),
      ]);

      const name = [profile.first_name, profile.last_name].filter(Boolean).join(' ') || profile.email;
      document.getElementById('welcome-name').textContent = name;

      const current = (companies || []).find((c) => c.company_id === claims.company_id);
      document.getElementById('welcome-company').textContent = current ? current.company_name : '—';
      document.getElementById('welcome-role').textContent = (claims.roles && claims.roles[0]) || (current && current.role_name) || '—';
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },
};
