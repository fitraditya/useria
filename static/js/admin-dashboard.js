const AdminDashboard = {
  async init() {
    const claims = Useria.requireCompany();
    if (!claims) return;

    const errorEl = document.getElementById('error');
    try {
      const stats = await Useria.apiFetch('/admin/stats');
      document.getElementById('stat-companies').textContent = stats.total_companies;
      document.getElementById('stat-users').textContent = stats.total_users;
    } catch (err) {
      Useria.showError(errorEl, err);
    }
  },
};
