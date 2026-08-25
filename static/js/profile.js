const Profile = {
  async init() {
    if (!Useria.requireAuth()) return;

    const form = document.getElementById('profile-form');
    const errorEl = document.getElementById('error');
    const modalErrorEl = document.getElementById('modal-error');
    const successEl = document.getElementById('success');

    const render = (user) => {
      const name = [user.first_name, user.last_name].filter(Boolean).join(' ') || user.email;
      document.getElementById('profile-avatar').textContent = Useria.initials(user.first_name, user.last_name, user.email);
      document.getElementById('profile-name').textContent = name;
      document.getElementById('profile-status').textContent = user.status;
      document.getElementById('info-first-name').textContent = user.first_name || '—';
      document.getElementById('info-last-name').textContent = user.last_name || '—';
      document.getElementById('info-email').textContent = user.email;
      form.first_name.value = user.first_name || '';
      form.last_name.value = user.last_name || '';
    };

    try {
      render(await Useria.apiFetch('/profile'));
    } catch (err) {
      Useria.showError(errorEl, err);
    }

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      Useria.hideError(modalErrorEl);
      successEl.classList.add('hidden');
      try {
        const user = await Useria.apiFetch('/profile', {
          method: 'PUT',
          body: JSON.stringify({
            first_name: form.first_name.value,
            last_name: form.last_name.value,
          }),
        });
        render(user);
        if (window.Alpine) Alpine.$data(document.body).isProfileInfoModal = false;
        successEl.textContent = 'Profile updated.';
        successEl.classList.remove('hidden');
      } catch (err) {
        Useria.showError(modalErrorEl, err);
      }
    });
  },
};
