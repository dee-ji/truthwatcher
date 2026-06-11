const app = document.getElementById("app");
const navLinks = Array.from(document.querySelectorAll("[data-nav]"));

const discoveryTasks = [
  "identify_device",
  "get_inventory",
  "get_neighbors",
  "get_bgp_summary",
];

window.addEventListener("hashchange", renderRoute);
window.addEventListener("DOMContentLoaded", renderRoute);

async function apiGet(path) {
  const response = await fetch(path, { headers: { Accept: "application/json" } });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(payload?.error?.message || `GET ${path} failed`);
  }
  return payload;
}

async function apiPost(path, body) {
  const response = await fetch(path, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(payload?.error?.message || `POST ${path} failed`);
  }
  return payload;
}

async function renderRoute() {
  const route = location.hash.replace(/^#/, "") || "/";
  setActiveNav(route);

  if (route === "/" || route === "") {
    await renderDashboard();
    return;
  }

  if (route === "/discovery-runs") {
    await renderDiscoveryRuns();
    return;
  }

  if (route.startsWith("/discovery-runs/")) {
    await renderDiscoveryRunDetail(route.split("/").pop());
    return;
  }

  app.innerHTML = `<section class="panel error-state">Page not found.</section>`;
}

function setActiveNav(route) {
  for (const link of navLinks) {
    link.classList.remove("active");
    link.removeAttribute("aria-current");
  }
  const active = route.startsWith("/discovery-runs")
    ? document.querySelector('[data-nav="discovery-runs"]')
    : document.querySelector('[data-nav="dashboard"]');
  if (active) {
    active.classList.add("active");
    active.setAttribute("aria-current", "page");
  }
}

async function renderDashboard() {
  app.innerHTML = `
    <section class="hero-panel">
      <div>
        <p class="eyebrow">Single binary UI foundation</p>
        <h1>Network truth starts with evidence.</h1>
        <p class="lede">
          This embedded frontend now exposes discovery runs and safe fixture-backed discovery execution.
          The next useful UI step is rendering assets and evidence without adding chat or realtime infrastructure.
        </p>
      </div>
      <aside class="status-card" aria-live="polite">
        <span class="status-dot" id="status-dot"></span>
        <div>
          <strong id="api-status">Checking API...</strong>
          <small id="api-version">Version unavailable</small>
        </div>
      </aside>
    </section>

    <section class="grid">
      <article class="card">
        <span class="card-label">Discovery Runs</span>
        <h2>Review collection history</h2>
        <p>List runs, inspect seed input, timestamps, and evidence counts.</p>
      </article>
      <article class="card">
        <span class="card-label">Fake Collector</span>
        <h2>Safe local execution</h2>
        <p>Start fixture-backed runs without touching a network device.</p>
      </article>
      <article class="card">
        <span class="card-label">Evidence First</span>
        <h2>Raw output remains primary</h2>
        <p>Every run stores evidence before model facts are created.</p>
      </article>
    </section>
  `;
  await checkAPI();
}

async function checkAPI() {
  const status = document.getElementById("api-status");
  const version = document.getElementById("api-version");
  const dot = document.getElementById("status-dot");
  if (!status || !version || !dot) {
    return;
  }

  try {
    const versionPayload = await apiGet("/api/v1/version");
    const appVersion = versionPayload?.data?.version || "unknown";

    status.textContent = "API ready";
    version.textContent = `truthwatcher ${appVersion}`;
    dot.classList.add("ok");
    dot.classList.remove("bad");
  } catch (error) {
    status.textContent = "API unavailable";
    version.textContent = "Check server logs";
    dot.classList.add("bad");
    dot.classList.remove("ok");
  }
}

async function renderDiscoveryRuns() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Discovery</p>
        <h1>Discovery runs</h1>
        <p>Review evidence-first collection attempts and start a safe fake run from local fixtures.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    ${fakeDiscoveryForm()}
    <section class="table-panel" id="runs-panel">
      <div class="empty-state">Loading discovery runs...</div>
    </section>
  `;

  document.getElementById("fake-discovery-form").addEventListener("submit", startFakeDiscovery);
  await loadDiscoveryRuns();
}

function fakeDiscoveryForm() {
  const taskControls = discoveryTasks.map((task) => `
    <label>
      <input type="checkbox" name="tasks" value="${escapeHTML(task)}" checked>
      ${escapeHTML(task)}
    </label>
  `).join("");

  return `
    <form class="form-panel" id="fake-discovery-form">
      <div class="form-grid">
        <div class="field">
          <label for="target">Fixture target</label>
          <input id="target" name="target" value="fixture://junos-mx" required>
        </div>
        <div class="field">
          <label for="profile">Profile</label>
          <select id="profile" name="profile">
            <option value="juniper_junos">juniper_junos</option>
            <option value="cisco_iosxr">cisco_iosxr</option>
          </select>
        </div>
        <div class="field">
          <label for="fixture-root">Fixture root</label>
          <input id="fixture-root" name="fixture_root" value="examples/fixtures">
        </div>
        <fieldset class="task-list">
          <legend>Tasks</legend>
          <div class="task-options">${taskControls}</div>
        </fieldset>
      </div>
      <div class="form-actions">
        <button type="submit">Start fake discovery</button>
        <span class="muted" id="form-message">Fake collector only. No network device will be contacted.</span>
      </div>
    </form>
  `;
}

async function startFakeDiscovery(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const button = form.querySelector("button");
  const message = document.getElementById("form-message");
  const formData = new FormData(form);
  const tasks = formData.getAll("tasks");

  button.disabled = true;
  message.textContent = "Starting fake discovery...";

  try {
    const payload = await apiPost("/api/v1/discovery-runs/execute", {
      collector: "fake",
      target: String(formData.get("target") || "").trim(),
      profile: String(formData.get("profile") || "").trim(),
      fixture_root: String(formData.get("fixture_root") || "").trim(),
      tasks,
    });
    const runID = payload?.data?.discovery_run?.id;
    message.textContent = "Fake discovery completed.";
    if (runID) {
      location.hash = `#/discovery-runs/${runID}`;
      return;
    }
    await loadDiscoveryRuns();
  } catch (error) {
    message.textContent = error.message;
  } finally {
    button.disabled = false;
  }
}

async function loadDiscoveryRuns() {
  const panel = document.getElementById("runs-panel");
  try {
    const payload = await apiGet("/api/v1/discovery-runs");
    const runs = payload?.data?.discovery_runs || [];
    if (runs.length === 0) {
      panel.innerHTML = `<div class="empty-state">No discovery runs yet. Start a fake run above.</div>`;
      return;
    }

    const rows = await Promise.all(runs.map(async (run) => {
      const evidenceCount = await evidenceCountForRun(run.id);
      return `
        <tr>
          <td><a href="#/discovery-runs/${escapeHTML(run.id)}">${escapeHTML(shortID(run.id))}</a></td>
          <td>${statusPill(run.status)}</td>
          <td>${escapeHTML(seedTarget(run.seed_input))}</td>
          <td>${escapeHTML(formatDate(run.created_at))}</td>
          <td>${escapeHTML(formatDate(run.completed_at))}</td>
          <td>${evidenceCount}</td>
        </tr>
      `;
    }));

    panel.innerHTML = `
      <table class="table">
        <thead>
          <tr>
            <th>Run</th>
            <th>Status</th>
            <th>Target</th>
            <th>Created</th>
            <th>Completed</th>
            <th>Evidence</th>
          </tr>
        </thead>
        <tbody>${rows.join("")}</tbody>
      </table>
    `;
  } catch (error) {
    panel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

async function renderDiscoveryRunDetail(id) {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Discovery details</p>
        <h1>Run ${escapeHTML(shortID(id))}</h1>
        <p>Inspect status, seed input, timestamps, and evidence count for one discovery run.</p>
      </div>
      <a class="button secondary" href="#/discovery-runs">Back to runs</a>
    </section>
    <section class="detail-panel" id="run-detail">
      <div class="empty-state">Loading discovery run...</div>
    </section>
  `;

  const panel = document.getElementById("run-detail");
  try {
    const [runPayload, evidencePayload] = await Promise.all([
      apiGet(`/api/v1/discovery-runs/${encodeURIComponent(id)}`),
      apiGet(`/api/v1/discovery-runs/${encodeURIComponent(id)}/evidence`),
    ]);
    const run = runPayload?.data?.discovery_run;
    const evidence = evidencePayload?.data?.evidence || [];

    panel.innerHTML = `
      <div class="detail-grid">
        <div class="metric">
          <small>Status</small>
          <strong>${statusPill(run.status)}</strong>
        </div>
        <div class="metric">
          <small>Evidence count</small>
          <strong>${evidence.length}</strong>
        </div>
        <div class="metric">
          <small>Started</small>
          <strong>${escapeHTML(formatDate(run.started_at))}</strong>
        </div>
        <div class="metric">
          <small>Completed</small>
          <strong>${escapeHTML(formatDate(run.completed_at))}</strong>
        </div>
      </div>
      <span class="code-block-label">Seed input</span>
      <pre class="code-block">${escapeHTML(JSON.stringify(run.seed_input || {}, null, 2))}</pre>
      ${run.error_message ? `<p class="error-state">${escapeHTML(run.error_message)}</p>` : ""}
    `;
  } catch (error) {
    panel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

async function evidenceCountForRun(id) {
  try {
    const payload = await apiGet(`/api/v1/discovery-runs/${encodeURIComponent(id)}/evidence`);
    return (payload?.data?.evidence || []).length;
  } catch (error) {
    return "Unavailable";
  }
}

function seedTarget(seedInput) {
  if (!seedInput || typeof seedInput !== "object") {
    return "Unknown";
  }
  return seedInput.target || seedInput.Target || "Unknown";
}

function statusPill(status) {
  const safeStatus = escapeHTML(status || "unknown");
  return `<span class="status-pill ${safeStatus}">${safeStatus}</span>`;
}

function formatDate(value) {
  if (!value) {
    return "Not set";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return date.toLocaleString();
}

function shortID(id) {
  if (!id) {
    return "unknown";
  }
  return String(id).slice(0, 8);
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
