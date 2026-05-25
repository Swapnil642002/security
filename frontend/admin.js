(function () {
  const TOKEN_KEY = "fm_token";

  const el = {
    userMeta: document.getElementById("user-meta"),
    logoutBtn: document.getElementById("logout"),
    metricActivePolicies: document.getElementById("metric-active-policies"),
    metricProvider: document.getElementById("metric-provider"),
    statusProvider: document.getElementById("status-provider"),
    statusLastSync: document.getElementById("status-last-sync"),
    statusPending: document.getElementById("status-pending"),
    countWebsite: document.getElementById("count-website"),
    countPort: document.getElementById("count-port"),
    countDepartment: document.getElementById("count-department"),
    auditList: document.getElementById("audit-list"),
    syncFirewallBtn: document.getElementById("sync-firewall"),
    syncMessage: document.getElementById("sync-message"),
    syncOutput: document.getElementById("sync-output"),

    jumpPolicyFormBtn: document.getElementById("jump-policy-form"),
    refreshPoliciesBtn: document.getElementById("refresh-policies"),
    policyPanel: document.getElementById("policy-panel"),
    policyForm: document.getElementById("policy-form"),
    policyId: document.getElementById("policy-id"),
    policyName: document.getElementById("policy-name"),
    policyType: document.getElementById("policy-type"),
    policyAction: document.getElementById("policy-action"),
    policyTarget: document.getElementById("policy-target"),
    policyDepartment: document.getElementById("policy-department"),
    policyEnabled: document.getElementById("policy-enabled"),
    policySchedule: document.getElementById("policy-schedule"),
    policySubmit: document.getElementById("policy-submit"),
    policyCancel: document.getElementById("policy-cancel"),
    policyMessage: document.getElementById("policy-message"),
    policyTableBody: document.getElementById("policy-table-body"),

    refreshDepartmentsBtn: document.getElementById("refresh-departments"),
    departmentForm: document.getElementById("department-form"),
    departmentName: document.getElementById("department-name"),
    departmentDescription: document.getElementById("department-description"),
    departmentMessage: document.getElementById("department-message"),
    departmentList: document.getElementById("department-list"),

    refreshLaptopsBtn: document.getElementById("refresh-laptops"),
    laptopForm: document.getElementById("laptop-form"),
    laptopHostname: document.getElementById("laptop-hostname"),
    laptopOS: document.getElementById("laptop-os"),
    laptopEmployeeName: document.getElementById("laptop-employee-name"),
    laptopEmployeeEmail: document.getElementById("laptop-employee-email"),
    laptopDepartment: document.getElementById("laptop-department"),
    laptopActive: document.getElementById("laptop-active"),
    laptopMessage: document.getElementById("laptop-message"),
    laptopTableBody: document.getElementById("laptop-table-body"),

    assignmentForm: document.getElementById("assignment-form"),
    assignPolicy: document.getElementById("assign-policy"),
    assignType: document.getElementById("assign-type"),
    assignDepartmentWrap: document.getElementById("assign-department-wrap"),
    assignDepartment: document.getElementById("assign-department"),
    assignLaptopWrap: document.getElementById("assign-laptop-wrap"),
    assignLaptop: document.getElementById("assign-laptop"),
    assignEnabled: document.getElementById("assign-enabled"),
    assignmentMessage: document.getElementById("assignment-message"),

    notifBell: document.getElementById("notif-bell"),
    notifBadge: document.getElementById("notif-badge"),
    notifList: document.getElementById("notif-list"),
    refreshNotificationsBtn: document.getElementById("refresh-notifications"),

    enrollLinkForm: document.getElementById("enroll-link-form"),
    enrollExpires: document.getElementById("enroll-expires"),
    enrollMaxUses: document.getElementById("enroll-max-uses"),
    enrollLinkMessage: document.getElementById("enroll-link-message"),
    enrollLinkOutput: document.getElementById("enroll-link-output"),
    refreshEnrollmentsBtn: document.getElementById("refresh-enrollments"),
    enrollmentTableBody: document.getElementById("enrollment-table-body"),
    enrollmentMessage: document.getElementById("enrollment-message"),
  };

  const state = {
    policies: [],
    departments: [],
    laptops: [],
    enrollments: [],
    unreadNotifCount: 0,
  };

  function redirectToLogin() {
    window.location.assign("/login");
  }

  function getTokenOrRedirect() {
    const token = localStorage.getItem(TOKEN_KEY);
    if (!token) {
      redirectToLogin();
      return "";
    }
    return token;
  }

  function authHeader(token) {
    return { Authorization: "Bearer " + token };
  }

  function isAuthStatus(status) {
    return status === 401 || status === 403;
  }

  function isAuthError(err) {
    return !!(err && isAuthStatus(err.status));
  }

  function escapeHtml(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/\"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function titleCase(value) {
    return String(value || "")
      .replace(/_/g, " ")
      .replace(/\b\w/g, function (c) {
        return c.toUpperCase();
      });
  }

  function formatDate(value) {
    if (!value) {
      return "Not available";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "Not available";
    }
    return date.toLocaleString();
  }

  function setMessage(node, message, isError) {
    node.textContent = message;
    node.style.color = isError ? "#b91c1c" : "#0f766e";
  }

  function renderAuditLogs(items) {
    if (!Array.isArray(items) || items.length === 0) {
      el.auditList.innerHTML = '<div class="timeline-item"><p class="muted">No audit records yet.</p></div>';
      return;
    }

    el.auditList.innerHTML = items
      .map(function (item) {
        return (
          '<div class="timeline-item">' +
          "<p><strong>" + titleCase(item.action || "unknown") + "</strong> on " + titleCase(item.entity || "entity") + "</p>" +
          '<p class="muted">' + formatDate(item.created_at) + "</p>" +
          "</div>"
        );
      })
      .join("");
  }

  async function apiRequest(path, options) {
    const token = getTokenOrRedirect();
    if (!token) {
      throw new Error("missing token");
    }
    const requestOptions = options || {};
    const headers = Object.assign({}, requestOptions.headers || {}, authHeader(token));
    const resp = await fetch(path, Object.assign({}, requestOptions, { headers: headers }));
    if (!resp.ok) {
      const text = (await resp.text()).trim();
      const err = new Error(text || "request failed");
      err.status = resp.status;
      throw err;
    }
    return resp;
  }

  function resetPolicyForm() {
    el.policyId.value = "";
    el.policyName.value = "";
    el.policyType.value = "website_category";
    el.policyAction.value = "block";
    el.policyTarget.value = "";
    el.policyDepartment.value = "";
    el.policyEnabled.value = "true";
    el.policySchedule.value = "{}";
    el.policySubmit.textContent = "Create Policy";
    el.policyCancel.hidden = true;
  }

  function renderPolicies(items) {
    if (!Array.isArray(items) || items.length === 0) {
      el.policyTableBody.innerHTML = '<tr><td colspan="7" class="muted">No policies yet.</td></tr>';
      return;
    }
    el.policyTableBody.innerHTML = items
      .map(function (item) {
        return (
          "<tr>" +
          "<td>" + escapeHtml(item.name) + "</td>" +
          "<td>" + escapeHtml(titleCase(item.policy_type)) + "</td>" +
          "<td>" + escapeHtml(titleCase(item.action)) + "</td>" +
          "<td>" + escapeHtml(item.target) + "</td>" +
          "<td>" + escapeHtml(item.department || "-") + "</td>" +
          '<td><span class="status-pill ' + (item.is_enabled ? "enabled" : "disabled") + '">' + (item.is_enabled ? "Enabled" : "Disabled") + "</span></td>" +
          '<td><div class="table-actions">' +
          '<button class="tiny-btn" data-action="edit" data-id="' + item.id + '">Edit</button>' +
          '<button class="tiny-btn" data-action="delete" data-id="' + item.id + '">Delete</button>' +
          "</div></td>" +
          "</tr>"
        );
      })
      .join("");
  }

  function renderDepartments(items) {
    if (!Array.isArray(items) || items.length === 0) {
      el.departmentList.innerHTML = '<li class="muted">No departments yet.</li>';
      return;
    }
    el.departmentList.innerHTML = items
      .map(function (item) {
        return "<li>" + escapeHtml(item.name) + "</li>";
      })
      .join("");
  }

  function renderLaptops(items) {
    if (!Array.isArray(items) || items.length === 0) {
      el.laptopTableBody.innerHTML = '<tr><td colspan="9" class="muted">No laptops yet.</td></tr>';
      return;
    }
    el.laptopTableBody.innerHTML = items
      .map(function (item) {
        const dept = getDepartmentName(item.department_id) || "-";
        const usbState = item.usb_storage_blocked
          ? '<span class="status-pill disabled">Blocked</span>'
          : '<span class="status-pill enabled">Allowed</span>';
        const tokenShort = item.agent_token ? item.agent_token.slice(0, 12) + "..." : "missing";
        return (
          "<tr>" +
          "<td>" + escapeHtml(item.hostname) + "</td>" +
          "<td>" + escapeHtml(item.employee_name) + "</td>" +
          "<td>" + escapeHtml(item.employee_email) + "</td>" +
          "<td>" + escapeHtml(titleCase(item.os_type)) + "</td>" +
          "<td>" + escapeHtml(dept) + "</td>" +
          '<td><span class="status-pill ' + (item.is_active ? "enabled" : "disabled") + '">' + (item.is_active ? "Active" : "Inactive") + "</span></td>" +
          "<td>" + usbState + '<div class="table-actions" style="margin-top:6px;">' +
            '<button class="tiny-btn" data-action="usb-block" data-id="' + item.id + '">Block USB</button>' +
            '<button class="tiny-btn" data-action="usb-unblock" data-id="' + item.id + '">Unblock USB</button>' +
          "</div></td>" +
          "<td><div style=\"font-family:monospace;font-size:0.78rem;\">" + escapeHtml(tokenShort) + "</div>" +
          '<button class="tiny-btn" data-action="copy-token" data-token="' + escapeHtml(item.agent_token || "") + '">Copy Token</button></td>' +
          '<td><button class="tiny-btn" data-action="delete-laptop" data-id="' + item.id + '">Delete</button></td>' +
          "</tr>"
        );
      })
      .join("");
  }

  async function queueUSBCommand(laptopID, block) {
    const endpoint = "/api/v1/laptops/" + laptopID + (block ? "/usb/block" : "/usb/unblock");
    await apiRequest(endpoint, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
    setMessage(el.laptopMessage, block ? "USB block command queued." : "USB unblock command queued.", false);
    await loadLaptops();
  }

  function getDepartmentName(id) {
    const found = state.departments.find(function (d) {
      return String(d.id) === String(id);
    });
    return found ? found.name : "";
  }

  function renderAssignmentOptions() {
    const policyOptions = state.policies.length
      ? state.policies.map(function (p) { return '<option value="' + p.id + '">' + escapeHtml(p.name) + "</option>"; }).join("")
      : '<option value="">No policies</option>';
    el.assignPolicy.innerHTML = policyOptions;

    const deptOptions = state.departments.length
      ? state.departments.map(function (d) { return '<option value="' + d.id + '">' + escapeHtml(d.name) + "</option>"; }).join("")
      : '<option value="">No departments</option>';
    el.assignDepartment.innerHTML = deptOptions;
    el.laptopDepartment.innerHTML = '<option value="">None</option>' + deptOptions;

    const laptopOptions = state.laptops.length
      ? state.laptops.map(function (l) { return '<option value="' + l.id + '">' + escapeHtml(l.hostname) + "</option>"; }).join("")
      : '<option value="">No laptops</option>';
    el.assignLaptop.innerHTML = laptopOptions;
  }

  function fillPolicyFormForEdit(item) {
    el.policyId.value = String(item.id);
    el.policyName.value = item.name || "";
    el.policyType.value = item.policy_type || "website_category";
    el.policyAction.value = item.action || "block";
    el.policyTarget.value = item.target || "";
    el.policyDepartment.value = item.department || "";
    el.policyEnabled.value = item.is_enabled ? "true" : "false";
    el.policySchedule.value = item.schedule_json || "{}";
    el.policySubmit.textContent = "Update Policy";
    el.policyCancel.hidden = false;
    el.policyPanel.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function normalizeWebsiteCategoryTarget(value) {
    return String(value || "")
      .trim()
      .toLowerCase()
      .replace(/-/g, "_")
      .replace(/\s+/g, "_");
  }

  function collectPolicyPayload() {
    const policyType = (el.policyType.value || "").trim().toLowerCase();
    const rawTarget = el.policyTarget.value.trim();
    const normalizedTarget = policyType === "website_category"
      ? normalizeWebsiteCategoryTarget(rawTarget)
      : rawTarget;

    return {
      name: el.policyName.value.trim(),
      policy_type: policyType,
      action: (el.policyAction.value || "").trim().toLowerCase(),
      target: normalizedTarget,
      department: el.policyDepartment.value.trim(),
      schedule_json: (el.policySchedule.value || "{}").trim() || "{}",
      is_enabled: el.policyEnabled.value === "true",
    };
  }

  async function loadCurrentUser() {
    const token = getTokenOrRedirect();
    if (!token) {
      return;
    }
    const resp = await fetch("/api/v1/auth/me", { headers: authHeader(token) });
    if (!resp.ok) {
      if (isAuthStatus(resp.status)) {
        localStorage.removeItem(TOKEN_KEY);
        redirectToLogin();
        return;
      }
      const text = (await resp.text()).trim();
      const err = new Error(text || "failed to load current user");
      err.status = resp.status;
      throw err;
    }
    const body = await resp.json();
    const user = body.user;
    el.userMeta.textContent = user.full_name + " (" + user.role + ")";
    document.title = "Firewall Manager - " + user.full_name;
  }

  async function loadDashboardSummary() {
    const resp = await apiRequest("/api/v1/dashboard/summary");
    const body = await resp.json();
    const summary = body.summary || {};

    const provider = titleCase(summary.firewall_provider || "not_configured");
    el.metricProvider.textContent = provider;
    el.statusProvider.textContent = provider;
    el.statusLastSync.textContent = formatDate(summary.last_sync_at);

    const counts = summary.policy_counts || {};
    el.metricActivePolicies.textContent = String(counts.total || 0);
    el.statusPending.textContent = String(counts.pending_changes || 0);
    el.countWebsite.textContent = String(counts.website_category || 0) + " rules";
    el.countPort.textContent = String(counts.port || 0) + " rules";
    el.countDepartment.textContent = String(counts.department_group || 0) + " groups";

    renderAuditLogs(summary.recent_audit_logs || []);
  }

  async function loadPolicies() {
    const resp = await apiRequest("/api/v1/policies");
    const body = await resp.json();
    state.policies = body.items || [];
    renderPolicies(state.policies);
    renderAssignmentOptions();
  }

  async function loadDepartments() {
    const resp = await apiRequest("/api/v1/departments");
    const body = await resp.json();
    state.departments = body.items || [];
    renderDepartments(state.departments);
    renderAssignmentOptions();
  }

  async function loadLaptops() {
    const resp = await apiRequest("/api/v1/laptops");
    const body = await resp.json();
    state.laptops = body.items || [];
    renderLaptops(state.laptops);
    renderAssignmentOptions();
  }

  async function loadEnrollments() {
    const resp = await apiRequest("/api/v1/enrollments");
    const body = await resp.json();
    state.enrollments = body.items || [];
    renderEnrollments(state.enrollments);
  }

  function renderEnrollments(items) {
    if (!Array.isArray(items) || items.length === 0) {
      el.enrollmentTableBody.innerHTML = '<tr><td colspan="9" class="muted">No enrollment requests yet.</td></tr>';
      return;
    }
    el.enrollmentTableBody.innerHTML = items
      .map(function (item) {
        const isPending = item.status === "pending";
        const isApproved = item.status === "approved";
        const statusClass = isPending ? "status-pill disabled" : (isApproved ? "status-pill enabled" : "status-pill disabled");
        const hasConsent = !!item.permission;
        const consentPill = hasConsent
          ? '<span class="status-pill enabled">Granted</span>'
          : '<span class="status-pill disabled">Missing</span>';
        const actions = isPending
          ? '<button class="approve-btn" data-action="approve" data-id="' + item.id + '" ' + (hasConsent ? "" : "disabled") + '>Approve</button>' +
            ' <button class="disable-btn" data-action="reject" data-id="' + item.id + '">Reject</button>'
          : isApproved
            ? '<button class="approve-btn" data-action="block-port" data-id="' + item.id + '" data-laptop-id="' + (item.laptop_id || "") + '">Block Port</button>' +
              ' <button class="disable-btn" data-action="disable" data-id="' + item.id + '">Disable</button>'
            : '<span class="muted">—</span>';
        return (
          "<tr>" +
          "<td>" + escapeHtml(item.hostname) + "</td>" +
          "<td>" + escapeHtml(item.employee_name) + "</td>" +
          "<td>" + escapeHtml(item.employee_email) + "</td>" +
          "<td>" + escapeHtml(titleCase(item.os_type)) + "</td>" +
          "<td>" + consentPill + "</td>" +
          "<td>" + escapeHtml(item.current_ip || "—") + "</td>" +
          "<td>" + formatDate(item.created_at) + "</td>" +
          '<td><span class="' + statusClass + '">' + escapeHtml(titleCase(item.status)) + "</span></td>" +
          "<td><div class=\"table-actions\">" + actions + "</div></td>" +
          "</tr>"
        );
      })
      .join("");
  }

  async function generateEnrollmentLink() {
    const hours = parseInt(el.enrollExpires.value, 10) || 48;
    const maxUses = parseInt(el.enrollMaxUses.value, 10) || 1;
    const resp = await apiRequest("/api/v1/enrollment-links", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expires_hours: hours, max_uses: maxUses }),
    });
    const body = await resp.json();
    setMessage(el.enrollLinkMessage, "Link generated. Copy and send to the employee.", false);
    el.enrollLinkOutput.hidden = false;
    el.enrollLinkOutput.textContent = body.link || "(no link returned)";
    el.enrollLinkForm.reset();
    el.enrollExpires.value = "48";
    el.enrollMaxUses.value = "1";
  }

  async function handleEnrollmentAction(enrollmentID, action) {
    if (action === "approve" || action === "reject") {
      const endpoint = action === "approve"
        ? "/api/v1/enrollments/" + enrollmentID + "/approve"
        : "/api/v1/enrollments/" + enrollmentID + "/disable";
      await apiRequest(endpoint, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      setMessage(el.enrollmentMessage, action === "approve" ? "Device approved and added to fleet." : "Request rejected.", false);
    } else if (action === "block-port") {
      const row = state.enrollments.find(function (entry) {
        return String(entry.id) === String(enrollmentID);
      });
      if (!row || !row.laptop_id) {
        throw new Error("Laptop not linked yet; approve the request first.");
      }

      const rawPort = window.prompt("Enter port/protocol to block (example: 445/tcp)");
      const target = (rawPort || "").trim();
      if (!target) {
        throw new Error("Port block canceled");
      }

      const policyResp = await apiRequest("/api/v1/policies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: "Auto Block " + target + " - " + row.hostname,
          policy_type: "port",
          action: "block",
          target: target,
          department: "",
          schedule_json: "{}",
          is_enabled: true,
        }),
      });
      const createdPolicy = (await policyResp.json()).item;

      await apiRequest("/api/v1/policy-assignments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          policy_id: Number(createdPolicy.id),
          assignment_type: "laptop",
          laptop_id: Number(row.laptop_id),
          department_id: null,
          is_enabled: true,
        }),
      });

      const syncResp = await apiRequest("/api/v1/firewall/sync", { method: "POST" });
      const syncBody = await syncResp.json();
      const syncResult = syncBody.result || {};
      if (syncResult.dry_run) {
        throw new Error("Firewall sync ran in dry-run mode. Set FIREWALL_DRY_RUN=false and retry.");
      }
      if ((syncResult.applied || 0) <= 0) {
        throw new Error("No firewall rules were applied. Check provider configuration and sync warnings.");
      }

      setMessage(el.enrollmentMessage, "Port " + target + " blocked for " + row.hostname + ".", false);
    } else if (action === "disable") {
      await apiRequest("/api/v1/enrollments/" + enrollmentID + "/disable", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      setMessage(el.enrollmentMessage, "Device access disabled.", false);
    }
    await Promise.all([loadEnrollments(), loadLaptops(), loadNotifications()]);
  }

  async function loadNotifications() {
    const resp = await apiRequest("/api/v1/notifications");
    const body = await resp.json();
    const items = body.items || [];
    const unread = items.filter(function (n) { return !n.is_read; }).length;
    state.unreadNotifCount = unread;
    if (unread > 0) {
      el.notifBadge.textContent = String(unread);
      el.notifBadge.hidden = false;
    } else {
      el.notifBadge.hidden = true;
    }
    if (!el.notifList) { return; }
    if (items.length === 0) {
      el.notifList.innerHTML = '<div class="timeline-item"><p class="muted">No notifications yet.</p></div>';
      return;
    }
    el.notifList.innerHTML = items.map(function (n) {
      const unreadMark = n.is_read ? "" : ' <span style="color:#f59e0b;font-weight:800;">•</span>';
      return '<div class="timeline-item"><p><strong>' + escapeHtml(titleCase(n.type)) + '</strong>' + unreadMark + '</p><p class="muted">' + escapeHtml(n.message) + '</p><p class="muted" style="font-size:0.78rem;">' + formatDate(n.created_at) + '</p></div>';
    }).join("");
  }

  async function savePolicy() {
    const payload = collectPolicyPayload();
    const id = el.policyId.value.trim();
    const isEdit = id !== "";

    await apiRequest(isEdit ? "/api/v1/policies/" + id : "/api/v1/policies", {
      method: isEdit ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    setMessage(el.policyMessage, isEdit ? "Policy updated." : "Policy created.", false);
    resetPolicyForm();
    await Promise.all([loadPolicies(), loadDashboardSummary()]);
  }

  async function deletePolicy(id) {
    await apiRequest("/api/v1/policies/" + id, { method: "DELETE" });
    setMessage(el.policyMessage, "Policy deleted.", false);
    await Promise.all([loadPolicies(), loadDashboardSummary()]);
  }

  async function createDepartment() {
    await apiRequest("/api/v1/departments", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: el.departmentName.value.trim(), description: el.departmentDescription.value.trim() }),
    });
    el.departmentForm.reset();
    setMessage(el.departmentMessage, "Department added.", false);
    await Promise.all([loadDepartments(), loadDashboardSummary()]);
  }

  async function createLaptop() {
    const deptID = el.laptopDepartment.value ? Number(el.laptopDepartment.value) : null;
    await apiRequest("/api/v1/laptops", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        hostname: el.laptopHostname.value.trim(),
        os_type: el.laptopOS.value,
        employee_name: el.laptopEmployeeName.value.trim(),
        employee_email: el.laptopEmployeeEmail.value.trim(),
        department_id: deptID,
        is_active: el.laptopActive.value === "true",
      }),
    });
    el.laptopForm.reset();
    el.laptopOS.value = "windows";
    el.laptopActive.value = "true";
    setMessage(el.laptopMessage, "Laptop added.", false);
    await loadLaptops();
  }

  async function deleteLaptop(id) {
    await apiRequest("/api/v1/laptops/" + id, { method: "DELETE" });
    setMessage(el.laptopMessage, "Laptop removed from management.", false);
    await Promise.all([loadLaptops(), loadEnrollments(), loadDashboardSummary()]);
  }

  async function createAssignment() {
    const type = el.assignType.value;
    const payload = {
      policy_id: Number(el.assignPolicy.value),
      assignment_type: type,
      department_id: type === "department" ? Number(el.assignDepartment.value) : null,
      laptop_id: type === "laptop" ? Number(el.assignLaptop.value) : null,
      is_enabled: el.assignEnabled.value === "true",
    };

    await apiRequest("/api/v1/policy-assignments", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    setMessage(el.assignmentMessage, "Policy assigned.", false);
    await loadDashboardSummary();
  }

  function toggleAssignmentInputs() {
    const isDepartment = el.assignType.value === "department";
    el.assignDepartmentWrap.hidden = !isDepartment;
    el.assignLaptopWrap.hidden = isDepartment;
  }

  async function syncFirewall() {
    el.syncFirewallBtn.disabled = true;
    setMessage(el.syncMessage, "Running firewall sync...", false);
    el.syncOutput.hidden = true;

    try {
      const resp = await apiRequest("/api/v1/firewall/sync", { method: "POST" });
      const body = await resp.json();
      const result = body.result || {};

      setMessage(el.syncMessage, "Firewall sync completed.", false);
      el.syncOutput.hidden = false;
      el.syncOutput.textContent = [
        "Provider: " + (result.provider || "unknown"),
        "Applied: " + String(result.applied || 0),
        "Skipped: " + String(result.skipped || 0),
        "Dry Run: " + String(!!result.dry_run),
        "Warnings: " + ((result.warnings || []).join(" | ") || "none"),
      ].join("\n");

      await loadDashboardSummary();
    } catch (err) {
      setMessage(el.syncMessage, err.message || "Firewall sync failed", true);
    } finally {
      el.syncFirewallBtn.disabled = false;
    }
  }

  el.logoutBtn.addEventListener("click", function () {
    localStorage.removeItem(TOKEN_KEY);
    redirectToLogin();
  });

  el.jumpPolicyFormBtn.addEventListener("click", function () {
    el.policyPanel.scrollIntoView({ behavior: "smooth", block: "start" });
    el.policyName.focus();
  });

  el.refreshPoliciesBtn.addEventListener("click", function () {
    loadPolicies().catch(function (err) {
      setMessage(el.policyMessage, err.message || "Failed to refresh policies", true);
    });
  });

  el.policyCancel.addEventListener("click", function () {
    resetPolicyForm();
    setMessage(el.policyMessage, "Edit canceled.", false);
  });

  el.policyForm.addEventListener("submit", function (event) {
    event.preventDefault();
    savePolicy().catch(function (err) {
      setMessage(el.policyMessage, err.message || "Failed to save policy", true);
    });
  });

  el.policyTableBody.addEventListener("click", function (event) {
    const button = event.target.closest("button[data-action]");
    if (!button) {
      return;
    }
    const id = button.getAttribute("data-id");
    const action = button.getAttribute("data-action");
    if (!id || !action) {
      return;
    }

    if (action === "delete") {
      if (!window.confirm("Delete this policy?")) {
        return;
      }
      deletePolicy(id).catch(function (err) {
        setMessage(el.policyMessage, err.message || "Failed to delete policy", true);
      });
      return;
    }

    if (action === "edit") {
      const item = state.policies.find(function (entry) {
        return String(entry.id) === String(id);
      });
      if (item) {
        fillPolicyFormForEdit(item);
      }
    }
  });

  el.refreshDepartmentsBtn.addEventListener("click", function () {
    loadDepartments().catch(function (err) {
      setMessage(el.departmentMessage, err.message || "Failed to refresh departments", true);
    });
  });

  el.departmentForm.addEventListener("submit", function (event) {
    event.preventDefault();
    createDepartment().catch(function (err) {
      setMessage(el.departmentMessage, err.message || "Failed to create department", true);
    });
  });

  el.refreshLaptopsBtn.addEventListener("click", function () {
    loadLaptops().catch(function (err) {
      setMessage(el.laptopMessage, err.message || "Failed to refresh laptops", true);
    });
  });

  el.laptopForm.addEventListener("submit", function (event) {
    event.preventDefault();
    createLaptop().catch(function (err) {
      setMessage(el.laptopMessage, err.message || "Failed to create laptop", true);
    });
  });

  el.laptopTableBody.addEventListener("click", function (event) {
    const button = event.target.closest("button[data-action]");
    if (!button) {
      return;
    }
    const action = button.getAttribute("data-action");
    if (action === "copy-token") {
      const token = button.getAttribute("data-token") || "";
      if (!token) {
        setMessage(el.laptopMessage, "Agent token is missing for this laptop.", true);
        return;
      }
      navigator.clipboard.writeText(token).then(function () {
        setMessage(el.laptopMessage, "Agent token copied.", false);
      }).catch(function () {
        setMessage(el.laptopMessage, "Copy failed. Please copy token manually.", true);
      });
      return;
    }

    const laptopID = button.getAttribute("data-id");
    if (!laptopID) {
      return;
    }
    if (action === "usb-block") {
      if (!window.confirm("Queue USB storage block command for this laptop?")) {
        return;
      }
      queueUSBCommand(laptopID, true).catch(function (err) {
        setMessage(el.laptopMessage, err.message || "Failed to queue USB block", true);
      });
      return;
    }
    if (action === "usb-unblock") {
      if (!window.confirm("Queue USB storage unblock command for this laptop?")) {
        return;
      }
      queueUSBCommand(laptopID, false).catch(function (err) {
        setMessage(el.laptopMessage, err.message || "Failed to queue USB unblock", true);
      });
      return;
    }
    if (action === "delete-laptop") {
      if (!window.confirm("Delete this laptop from company management? This removes policy assignment and command history for that device.")) {
        return;
      }
      deleteLaptop(laptopID).catch(function (err) {
        setMessage(el.laptopMessage, err.message || "Failed to delete laptop", true);
      });
    }
  });

  el.assignType.addEventListener("change", toggleAssignmentInputs);

  el.assignmentForm.addEventListener("submit", function (event) {
    event.preventDefault();
    createAssignment().catch(function (err) {
      setMessage(el.assignmentMessage, err.message || "Failed to assign policy", true);
    });
  });

  el.syncFirewallBtn.addEventListener("click", function () {
    syncFirewall();
  });

  el.enrollLinkForm.addEventListener("submit", function (event) {
    event.preventDefault();
    generateEnrollmentLink().catch(function (err) {
      setMessage(el.enrollLinkMessage, err.message || "Failed to generate link", true);
    });
  });

  el.refreshEnrollmentsBtn.addEventListener("click", function () {
    loadEnrollments().catch(function (err) {
      setMessage(el.enrollmentMessage, err.message || "Failed to refresh enrollments", true);
    });
  });

  el.enrollmentTableBody.addEventListener("click", function (event) {
    const button = event.target.closest("button[data-action]");
    if (!button) { return; }
    const id = button.getAttribute("data-id");
    const action = button.getAttribute("data-action");
    if (!id || !action) { return; }
    if (action !== "block-port") {
      const confirmMsg = action === "approve" ? "Approve this device for policy control?" : action === "disable" ? "Disable this device's access?" : "Reject this enrollment request?";
      if (!window.confirm(confirmMsg)) { return; }
    }
    handleEnrollmentAction(id, action).catch(function (err) {
      if (action === "block-port" && /canceled/i.test(err.message || "")) {
        setMessage(el.enrollmentMessage, "Port block canceled.", false);
        return;
      }
      setMessage(el.enrollmentMessage, err.message || "Action failed", true);
    });
  });

  el.notifBell.addEventListener("click", function () {
    document.getElementById("enrollment-panel") && document.getElementById("enrollment-panel").scrollIntoView({ behavior: "smooth", block: "start" });
    loadNotifications().catch(function () {});
  });

  if (el.refreshNotificationsBtn) {
    el.refreshNotificationsBtn.addEventListener("click", function () {
      loadNotifications().catch(function (err) {
        console.error(err);
      });
    });
  }

  toggleAssignmentInputs();
  resetPolicyForm();

  Promise.all([loadCurrentUser(), loadDashboardSummary(), loadPolicies(), loadDepartments(), loadLaptops(), loadEnrollments(), loadNotifications()]).catch(function (err) {
    if (isAuthError(err)) {
      localStorage.removeItem(TOKEN_KEY);
      redirectToLogin();
      return;
    }
    console.error(err);
    el.userMeta.textContent = "Signed in, but some dashboard data failed to load.";
  });
})();
