(function () {
  const TOKEN_KEY = "fm_token";
  const form = document.getElementById("login-form");
  const errorNode = document.getElementById("error");
  const submitBtn = document.getElementById("submit-btn");

  async function isLoggedIn(token) {
    const resp = await fetch("/api/v1/auth/me", {
      headers: {
        Authorization: "Bearer " + token,
      },
    });
    return resp.ok;
  }

  async function bootstrapSessionCheck() {
    const params = new URLSearchParams(window.location.search);
    const enrollToken = params.get("enroll_token") || params.get("token");
    if (enrollToken) {
      window.location.replace("/enroll?token=" + encodeURIComponent(enrollToken));
      return;
    }

    const existing = localStorage.getItem(TOKEN_KEY);
    if (!existing) {
      return;
    }

    try {
      if (await isLoggedIn(existing)) {
        window.location.assign("/admin");
      }
    } catch (_) {
      localStorage.removeItem(TOKEN_KEY);
    }
  }

  form.addEventListener("submit", async function (event) {
    event.preventDefault();
    errorNode.textContent = "";
    submitBtn.disabled = true;
    submitBtn.textContent = "Signing In...";

    const payload = {
      email: form.email.value,
      password: form.password.value,
    };

    try {
      const resp = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!resp.ok) {
        const message = (await resp.text()).trim() || "Login failed";
        errorNode.textContent = message;
        submitBtn.disabled = false;
        submitBtn.textContent = "Sign In";
        return;
      }

      const data = await resp.json();
      localStorage.setItem(TOKEN_KEY, data.token);
      window.location.assign("/admin");
    } catch (_) {
      errorNode.textContent = "Network error. Try again.";
      submitBtn.disabled = false;
      submitBtn.textContent = "Sign In";
    }
  });

  bootstrapSessionCheck();
})();
