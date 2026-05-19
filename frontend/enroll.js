document.addEventListener('DOMContentLoaded', function () {
  const form = document.getElementById('enroll-form');
  const errorNode = document.getElementById('enroll-error');
  const successNode = document.getElementById('enroll-success');
  const submitBtn = document.getElementById('enroll-submit');

  const urlParams = new URLSearchParams(window.location.search);
  const enrollToken = urlParams.get('token') || urlParams.get('enroll_token');

  if (!enrollToken) {
    errorNode.textContent = 'Invalid or missing enrollment link.';
    form.hidden = true;
    return;
  }

  form.addEventListener('submit', async function (event) {
    event.preventDefault();
    errorNode.textContent = '';
    successNode.textContent = '';
    submitBtn.disabled = true;
    submitBtn.textContent = 'Submitting...';

    const payload = {
      token: enrollToken,
      hostname: form.hostname.value,
      employee_name: form.employee_name.value,
      employee_email: form.employee_email.value,
      os_type: form.os_type.value,
      permission: form.permission.checked,
    };

    try {
      const resp = await fetch('/api/v1/public/enroll/accept', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        const message = (await resp.text()).trim() || 'Enrollment failed.';
        errorNode.textContent = message;
        submitBtn.disabled = false;
        submitBtn.textContent = 'Submit Request';
        return;
      }
      successNode.textContent = 'Enrollment request submitted! Await admin approval.';
      form.reset();
    } catch (e) {
      errorNode.textContent = 'Network error. Try again.';
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Submit Request';
    }
  });
});
