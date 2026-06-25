const app = document.getElementById("app");
const navLinks = Array.from(document.querySelectorAll("[data-nav]"));

const discoveryTasks = [
  "identify_device",
  "get_inventory",
  "get_neighbors",
  "get_bgp_summary",
];

const agentHistoryKey = "truthwatcher.agent.history";

window.addEventListener("hashchange", renderRoute);
window.addEventListener("DOMContentLoaded", () => {
  void loadShellVersion();
  void renderRoute();
});

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

async function loadShellVersion() {
  const shellVersion = document.getElementById("shell-version");
  if (!shellVersion) {
    return;
  }
  try {
    const payload = await apiGet("/api/version");
    const info = payload?.data || {};
    shellVersion.textContent = `${info.name || "truthwatcher"} ${info.version || "unknown"}`;
    shellVersion.title = `commit ${info.commit || "unknown"}; build ${info.build_date || "unknown"}`;
  } catch (error) {
    shellVersion.textContent = "Version unavailable";
  }
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

  if (route === "/assets") {
    await renderAssetsView();
    return;
  }

  if (route === "/identity-review" || route.startsWith("/identity-review?")) {
    await renderIdentityReviewView();
    return;
  }

  if (route === "/audit" || route.startsWith("/audit?")) {
    await renderAuditView();
    return;
  }

  if (route.startsWith("/assets/")) {
    await renderAssetDetail(route.split("/").pop());
    return;
  }

  if (route === "/discovery-plans") {
    renderDiscoveryPlansView();
    return;
  }

  if (route === "/architecture-seeds") {
    renderArchitectureSeedsView();
    return;
  }

  if (route === "/graph") {
    await renderGraphView();
    return;
  }

  if (route === "/ask") {
    renderAskView();
    return;
  }

  if (route === "/about") {
    await renderAboutView();
    return;
  }

  app.innerHTML = `<section class="panel error-state">Page not found.</section>`;
}

function setActiveNav(route) {
  for (const link of navLinks) {
    link.classList.remove("active");
    link.removeAttribute("aria-current");
  }
  let active = document.querySelector('[data-nav="dashboard"]');
  if (route.startsWith("/discovery-runs")) {
    active = document.querySelector('[data-nav="discovery-runs"]');
  }
  if (route.startsWith("/assets")) {
    active = document.querySelector('[data-nav="assets"]');
  }
  if (route.startsWith("/identity-review")) {
    active = document.querySelector('[data-nav="identity-review"]');
  }
  if (route === "/audit") {
    active = document.querySelector('[data-nav="audit"]');
  }
  if (route === "/discovery-plans") {
    active = document.querySelector('[data-nav="discovery-plans"]');
  }
  if (route === "/architecture-seeds") {
    active = document.querySelector('[data-nav="architecture-seeds"]');
  }
  if (route === "/graph") {
    active = document.querySelector('[data-nav="graph"]');
  }
  if (route === "/ask") {
    active = document.querySelector('[data-nav="ask"]');
  }
  if (route === "/about") {
    active = document.querySelector('[data-nav="about"]');
  }
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
      <article class="card">
        <span class="card-label">Assets</span>
        <h2>Browse modeled knowledge</h2>
        <p>Filter assets, inspect facts and relationships, and open linked raw evidence.</p>
      </article>
      <article class="card">
        <span class="card-label">Graph</span>
        <h2>Inspect relationships</h2>
        <p>Render a small asset neighborhood with confidence visible on every edge.</p>
      </article>
      <article class="card">
        <span class="card-label">Planner</span>
        <h2>Review safe next steps</h2>
        <p>Create read-only discovery plans that require human approval before execution.</p>
      </article>
      <article class="card">
        <span class="card-label">Architecture Seeds</span>
        <h2>Add planning context</h2>
        <p>Record ASNs, vendors, route reflectors, EMS systems, services, and regions as user-seeded context.</p>
      </article>
      <article class="card">
        <span class="card-label">Ask Truthwatcher</span>
        <h2>Deterministic workspace</h2>
        <p>Ask canned read-only questions without external LLM calls or network actions.</p>
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
    const versionPayload = await apiGet("/api/version");
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
        <p>Inspect this read-only run, its raw evidence records, audit trail, and next parsing or graph inspection steps.</p>
      </div>
      <a class="button secondary" href="#/discovery-runs">Back to runs</a>
    </section>
    <section class="detail-panel" id="run-detail">
      <div class="empty-state">Loading discovery run...</div>
    </section>
    ${evidenceDrawerMarkup()}
  `;
  setupEvidenceDrawer();

  const panel = document.getElementById("run-detail");
  try {
    const runPayload = await apiGet(`/api/v1/discovery-runs/${encodeURIComponent(id)}`);
    const run = runPayload?.data?.discovery_run;
    panel.innerHTML = discoveryRunDetailMarkup(run, [], true);

    try {
      const evidencePayload = await apiGet(`/api/v1/discovery-runs/${encodeURIComponent(id)}/evidence`);
      const evidence = evidencePayload?.data?.evidence || [];
      panel.innerHTML = discoveryRunDetailMarkup(run, evidence, false);
      setupEvidenceLinks(panel);
    } catch (evidenceError) {
      panel.innerHTML = discoveryRunDetailMarkup(run, [], false, evidenceError);
    }
  } catch (error) {
    panel.innerHTML = `<div class="error-state">Could not load discovery run: ${escapeHTML(error.message)}</div>`;
  }
}

function discoveryRunDetailMarkup(run, evidence, evidenceLoading, evidenceError) {
  const runID = run?.id || "";
  const parseRunID = escapeHTML(runID || "<run-id>");
  return `
    <div class="detail-grid">
      <div class="metric">
        <small>Status</small>
        <strong>${statusPill(run?.status)}</strong>
      </div>
      <div class="metric">
        <small>Evidence count</small>
        <strong>${evidenceLoading ? "Loading..." : evidence.length}</strong>
      </div>
      <div class="metric">
        <small>Started</small>
        <strong>${escapeHTML(formatDate(run?.started_at))}</strong>
      </div>
      <div class="metric">
        <small>Completed</small>
        <strong>${escapeHTML(formatDate(run?.completed_at))}</strong>
      </div>
    </div>
    <div class="next-step-panel">
      <strong>Next steps</strong>
      <ul>
        <li>Inspect raw evidence with the Open evidence actions below. Raw output stays hidden until the read-only drawer opens.</li>
        <li>If assets or graph data are missing, parse this run with <code>truthwatcher parse discovery-run --id ${parseRunID} --platform &lt;platform&gt;</code> or <code>POST /api/v1/discovery-runs/${parseRunID}/parse</code>.</li>
        <li><a href="#/assets">Inspect assets</a>${firstGraphLink(evidence)} after parsing.</li>
        <li><a href="#/audit?discovery_run_id=${encodeURIComponent(runID)}">Inspect audit records for this run</a>.</li>
      </ul>
    </div>
    <span class="code-block-label">Seed input</span>
    <pre class="code-block">${escapeHTML(JSON.stringify(run?.seed_input || {}, null, 2))}</pre>
    ${run?.error_message ? `<p class="error-state">${escapeHTML(run.error_message)}</p>` : ""}
    <h2>Evidence collected by this run</h2>
    <p class="muted">Discovery evidence is read-only. Command output is intentionally not rendered in this table; open a record to inspect raw output and integrity metadata.</p>
    ${discoveryRunEvidenceMarkup(evidence, evidenceLoading, evidenceError)}
  `;
}

function discoveryRunEvidenceMarkup(evidence, loading, error) {
  if (loading) {
    return `<div class="empty-state">Loading evidence records...</div>`;
  }
  if (error) {
    return `<div class="error-state">Could not load evidence for this run: ${escapeHTML(error.message)}</div>`;
  }
  if (evidence.length === 0) {
    return `<div class="empty-state">No evidence records are stored for this run yet. Evidence appears after successful discovery execution; after evidence exists, parse the discovery run before expecting assets or graph relationships.</div>`;
  }
  return `
    <div class="table-panel run-evidence-table">
      <table class="table">
        <thead>
          <tr>
            <th>Command/API</th>
            <th>Method</th>
            <th>Target</th>
            <th>Profile</th>
            <th>Collected</th>
            <th>Raw output hash</th>
            <th>Evidence ID</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          ${evidence.map((item) => `
            <tr>
              <td>${escapeHTML(item.command_or_api || "unknown")}</td>
              <td>${escapeHTML(item.method || "unknown")}</td>
              <td>${escapeHTML(item.target || "unknown")}</td>
              <td>${escapeHTML(evidenceMetadataField(item, "profile") || "not recorded")}</td>
              <td>${escapeHTML(formatDate(item.collected_at))}</td>
              <td><code>${escapeHTML(item.raw_output_hash || "missing")}</code></td>
              <td><code>${escapeHTML(item.id || "unknown")}</code></td>
              <td>${evidenceButton(item.id)}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    </div>
  `;
}

function evidenceMetadataField(item, field) {
  const metadata = parseJSONValue(item?.metadata);
  return metadata && typeof metadata[field] === "string" ? metadata[field] : "";
}

function parseJSONValue(value) {
  if (!value) {
    return null;
  }
  if (typeof value === "object") {
    return value;
  }
  try {
    return JSON.parse(value);
  } catch (_error) {
    return null;
  }
}

function firstGraphLink() {
  return ` and <a href="#/graph">inspect the graph</a>`;
}

async function renderAuditView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Audit</p>
        <h1>Audit records</h1>
        <p>Audit inspection is read-only. Records show the safe discovery actions Truthwatcher mapped and persisted.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <form class="form-panel" id="audit-filter-form">
      <div class="form-grid audit-filter-grid">
        <div class="field"><label for="audit-status-filter">Status</label><input id="audit-status-filter" name="status" placeholder="completed"></div>
        <div class="field"><label for="audit-action-filter">Action</label><input id="audit-action-filter" name="action" placeholder="discovery_command"></div>
        <div class="field"><label for="audit-target-filter">Target</label><input id="audit-target-filter" name="target" placeholder="fixture://junos-mx"></div>
        <div class="field"><label for="audit-run-filter">Discovery run ID</label><input id="audit-run-filter" name="discovery_run_id" placeholder="uuid"></div>
        <div class="field"><label for="audit-evidence-filter">Evidence ID</label><input id="audit-evidence-filter" name="evidence_id" placeholder="uuid"></div>
      </div>
      <div class="form-actions">
        <button type="submit">Apply filters</button>
        <button class="secondary" type="button" id="clear-audit-filters">Clear</button>
        <span class="muted" id="audit-filter-message">Shows up to 50 records. Audit records should not contain raw credentials or secrets.</span>
      </div>
    </form>
    <section class="table-panel audit-table-panel" id="audit-panel">
      <div class="empty-state">Loading audit records...</div>
    </section>
    <div class="drawer-backdrop hidden" id="evidence-drawer-backdrop"></div>
    <aside class="evidence-drawer hidden" id="evidence-drawer" aria-label="Evidence drawer">
      <button class="drawer-close" id="close-evidence-drawer" type="button">Close</button>
      <div id="evidence-drawer-body"></div>
    </aside>
  `;
  setupEvidenceDrawer();
  const auditQuery = new URLSearchParams((location.hash.split("?")[1] || ""));
  for (const [key, value] of auditQuery.entries()) {
    const field = document.querySelector(`#audit-filter-form [name="${CSS.escape(key)}"]`);
    if (field) {
      field.value = value;
    }
  }
  document.getElementById("audit-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadAuditRecordsFromFilters();
  });
  document.getElementById("clear-audit-filters").addEventListener("click", () => {
    document.getElementById("audit-filter-form").reset();
    loadAuditRecordsFromFilters();
  });
  await loadAuditRecordsFromFilters();
}

async function loadAuditRecordsFromFilters() {
  const panel = document.getElementById("audit-panel");
  const message = document.getElementById("audit-filter-message");
  const form = document.getElementById("audit-filter-form");
  const params = new URLSearchParams({ limit: "50" });
  for (const [key, value] of new FormData(form).entries()) {
    const trimmed = String(value || "").trim();
    if (trimmed) {
      params.set(key, trimmed);
    }
  }

  panel.innerHTML = `<div class="empty-state">Loading audit records...</div>`;
  try {
    const payload = await apiGet(`/api/v1/audit-records?${params.toString()}`);
    const records = payload?.data?.audit_records || [];
    message.textContent = `${records.length} audit records shown. Audit inspection is read-only.`;
    if (records.length === 0) {
      panel.innerHTML = `<div class="empty-state">No audit records match these filters. Audit records are created by discovery execution, including fake fixture-backed runs.</div>`;
      return;
    }
    panel.innerHTML = `
      <table class="table audit-table">
        <thead><tr><th>Action</th><th>Status</th><th>Target</th><th>Method</th><th>Profile</th><th>Task</th><th>Command/API</th><th>Initiator</th><th>Discovery run</th><th>Evidence</th><th>Started</th><th>Completed</th><th>Error</th></tr></thead>
        <tbody>${records.map((record) => `
          <tr>
            <td>${escapeHTML(record.action || "unknown")}</td>
            <td>${statusPill(record.status || "unknown")}</td>
            <td>${escapeHTML(record.target || "unknown")}</td>
            <td>${escapeHTML(record.method || "unknown")}</td>
            <td>${escapeHTML(record.profile || "")}</td>
            <td>${escapeHTML(record.task || "")}</td>
            <td><code>${escapeHTML(record.command_or_api || "")}</code></td>
            <td>${escapeHTML(record.initiator || "unknown")}</td>
            <td>${record.discovery_run_id ? `<a href="#/discovery-runs/${escapeHTML(record.discovery_run_id)}">${escapeHTML(shortID(record.discovery_run_id))}</a>` : ""}</td>
            <td>${record.evidence_id ? `<button class="link-button" type="button" data-evidence-id="${escapeHTML(record.evidence_id)}">${escapeHTML(shortID(record.evidence_id))}</button>` : ""}</td>
            <td>${escapeHTML(formatDate(record.started_at))}</td>
            <td>${escapeHTML(formatDate(record.completed_at))}</td>
            <td>${record.error ? `<span class="error-text">${escapeHTML(record.error)}</span>` : ""}</td>
          </tr>`).join("")}</tbody>
      </table>
    `;
    attachEvidenceButtons(panel);
  } catch (error) {
    panel.innerHTML = `<div class="error-state">Could not load audit records: ${escapeHTML(error.message)}</div>`;
  }
}

async function renderAssetsView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Assets</p>
        <h1>Asset browser</h1>
        <p>Filter persisted assets using the same fields supported by the API. Identity strength, confidence, and state remain visible.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <form class="form-panel" id="asset-filter-form">
      <div class="form-grid asset-filter-grid">
        <div class="field">
          <label for="asset-type-filter">Type</label>
          <input id="asset-type-filter" name="type" placeholder="device">
        </div>
        <div class="field">
          <label for="asset-vendor-filter">Vendor</label>
          <input id="asset-vendor-filter" name="vendor" placeholder="juniper">
        </div>
        <div class="field">
          <label for="asset-serial-filter">Serial</label>
          <input id="asset-serial-filter" name="serial" placeholder="JN1234">
        </div>
        <div class="field">
          <label for="asset-identity-filter">Identity key</label>
          <input id="asset-identity-filter" name="identity_key" placeholder="device:hostname:mx-edge-01">
        </div>
      </div>
      <div class="form-actions">
        <button type="submit">Apply filters</button>
        <button class="secondary" type="button" id="clear-asset-filters">Clear</button>
        <span class="muted" id="asset-filter-message">Shows up to 100 matching assets.</span>
      </div>
    </form>
    <section class="table-panel" id="assets-panel">
      <div class="empty-state">Loading assets...</div>
    </section>
  `;

  document.getElementById("asset-filter-form").addEventListener("submit", (event) => {
    event.preventDefault();
    loadAssetsFromFilters();
  });
  document.getElementById("clear-asset-filters").addEventListener("click", () => {
    document.getElementById("asset-filter-form").reset();
    loadAssetsFromFilters();
  });
  await loadAssetsFromFilters();
}

async function loadAssetsFromFilters() {
  const panel = document.getElementById("assets-panel");
  const message = document.getElementById("asset-filter-message");
  const form = document.getElementById("asset-filter-form");
  const params = new URLSearchParams({ limit: "100" });
  for (const [key, value] of new FormData(form).entries()) {
    const trimmed = String(value || "").trim();
    if (trimmed) {
      params.set(key, trimmed);
    }
  }

  panel.innerHTML = `<div class="empty-state">Loading assets...</div>`;
  try {
    const payload = await apiGet(`/api/v1/assets?${params.toString()}`);
    const assets = payload?.data?.assets || [];
    const pagination = payload?.metadata?.pagination;
    message.textContent = `${assets.length} assets shown${pagination ? ` of ${pagination.total}` : ""}.`;
    if (assets.length === 0) {
      panel.innerHTML = `<div class="empty-state">No assets match these filters.</div>`;
      return;
    }
    panel.innerHTML = `
      <table class="table">
        <thead>
          <tr>
            <th>Asset</th>
            <th>Type</th>
            <th>Vendor</th>
            <th>Serial</th>
            <th>Identity</th>
            <th>Confidence</th>
            <th>State</th>
          </tr>
        </thead>
        <tbody>
          ${assets.map((asset) => `
            <tr>
              <td><a href="#/assets/${escapeHTML(asset.id)}">${escapeHTML(assetLabel(asset))}</a></td>
              <td>${escapeHTML(asset.type || "unknown")}</td>
              <td>${escapeHTML(asset.vendor || "unknown")}</td>
              <td>${escapeHTML(asset.serial || "unknown")}</td>
              <td>
                <code>${escapeHTML(asset.identity_key || "")}</code>
                ${identityBadge(asset)}
              </td>
              <td>${escapeHTML(confidenceLabel(asset))}</td>
              <td>${stateBadge(asset.state)}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    `;
  } catch (error) {
    message.textContent = error.message;
    panel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

async function renderAssetDetail(id) {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Asset details</p>
        <h1>Asset ${escapeHTML(shortID(id))}</h1>
        <p>Review facts, relationships, confidence, state, and evidence references without hiding uncertainty.</p>
      </div>
      <a class="button secondary" href="#/assets">Back to assets</a>
    </section>
    <section class="detail-panel" id="asset-detail">
      <div class="empty-state">Loading asset...</div>
    </section>
    ${evidenceDrawerMarkup()}
  `;
  setupEvidenceDrawer();

  const panel = document.getElementById("asset-detail");
  try {
    const [assetPayload, factsPayload, relationshipsPayload, evidencePayload] = await Promise.all([
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}`),
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}/facts?limit=100`),
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}/relationships?limit=100`),
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}/evidence?limit=100`),
    ]);
    const asset = assetPayload?.data?.asset;
    const facts = factsPayload?.data?.facts || [];
    const relationships = relationshipsPayload?.data?.relationships || [];
    const evidenceItems = evidencePayload?.data?.evidence || [];

    panel.innerHTML = `
      <div class="detail-grid">
        <div class="metric">
          <small>Type</small>
          <strong>${escapeHTML(asset.type || "unknown")}</strong>
        </div>
        <div class="metric">
          <small>Confidence</small>
          <strong>${escapeHTML(confidenceLabel(asset))}</strong>
        </div>
        <div class="metric">
          <small>State</small>
          <strong>${stateBadge(asset.state)}</strong>
        </div>
        <div class="metric">
          <small>Identity</small>
          <strong>${identityBadge(asset)}</strong>
        </div>
      </div>
      <span class="code-block-label">Identity key</span>
      <pre class="code-block">${escapeHTML(asset.identity_key || asset.id)}</pre>
      <div class="asset-meta-grid">
        <div>
          <span class="code-block-label">Facts</span>
          ${assetFactsMarkup(facts)}
        </div>
        <div>
          <span class="code-block-label">Relationships</span>
          ${assetRelationshipsMarkup(relationships, asset.id)}
        </div>
      </div>
      <span class="code-block-label">Evidence</span>
      ${assetEvidenceMarkup(evidenceItems)}
      <span class="code-block-label">Metadata</span>
      <pre class="code-block">${escapeHTML(JSON.stringify(asset.metadata || {}, null, 2))}</pre>
    `;
    attachEvidenceButtons(panel);
  } catch (error) {
    panel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

function assetFactsMarkup(facts) {
  if (facts.length === 0) {
    return `<p class="muted">No facts recorded for this asset.</p>`;
  }
  return `
    <ul class="edge-list">
      ${facts.map((fact) => `
        <li class="${fact.state === "conflicting" ? "conflict-row" : ""}">
          <strong>${escapeHTML(fact.name || "fact")}</strong>
          <span>${escapeHTML(factValueLabel(fact.value))}</span>
          <small>${escapeHTML(confidenceLabel(fact))}</small>
          <small>${escapeHTML(fact.confidence_reason || "")}</small>
          ${evidenceButton(fact.evidence_id)}
        </li>
      `).join("")}
    </ul>
  `;
}

function assetEvidenceMarkup(evidenceItems) {
  if (evidenceItems.length === 0) {
    return `<p class="muted">No evidence is linked to this asset yet.</p>`;
  }
  return `
    <ul class="edge-list">
      ${evidenceItems.map((item) => `
        <li>
          <strong>${escapeHTML(item.command_or_api || item.method || "evidence")}</strong>
          <span>${escapeHTML(item.target || "unknown target")}</span>
          <small>${escapeHTML(formatDate(item.collected_at))}</small>
          ${evidenceButton(item.id)}
        </li>
      `).join("")}
    </ul>
  `;
}

function assetRelationshipsMarkup(relationships, assetID) {
  if (relationships.length === 0) {
    return `<p class="muted">No relationships recorded for this asset.</p>`;
  }
  return `
    <ul class="edge-list">
      ${relationships.map((relationship) => {
        const direction = relationship.source_asset_id === assetID ? "outgoing" : "incoming";
        const peer = relationship.source_asset_id === assetID ? relationship.target_asset_id : relationship.source_asset_id;
        return `
          <li>
            <strong>${escapeHTML(relationship.relationship_type || "related")}</strong>
            <span>${escapeHTML(direction)} peer ${escapeHTML(peer || "unknown")}</span>
            <small>${escapeHTML(confidenceLabel(relationship))}</small>
            <small>${escapeHTML(relationship.confidence_reason || "")}</small>
            ${evidenceButton(relationship.evidence_id)}
          </li>
        `;
      }).join("")}
    </ul>
  `;
}

function renderDiscoveryPlansView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Discovery planner</p>
        <h1>Review safe discovery plans</h1>
        <p>Suggest read-only next steps from a single explicit target. Plans require human approval and are not executed automatically.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <section class="plan-layout">
      <form class="form-panel" id="discovery-plan-form">
        <div class="readonly-note">
          Planning only. This page cannot execute discovery, guess credentials, or expand target scope.
        </div>
        <div class="form-grid">
          <div class="field">
            <label for="plan-target">Target</label>
            <input id="plan-target" name="target" value="fixture://junos-mx" required>
          </div>
          <div class="field">
            <label for="plan-method">Method</label>
            <select id="plan-method" name="method">
              <option value="fake">fake</option>
              <option value="ssh">ssh</option>
            </select>
          </div>
          <div class="field">
            <label for="plan-profile">Profile</label>
            <select id="plan-profile" name="profile">
              <option value="juniper_junos">juniper_junos</option>
              <option value="cisco_iosxr">cisco_iosxr</option>
            </select>
          </div>
          <div class="field">
            <label for="plan-seed-input">Seed input JSON</label>
            <textarea id="plan-seed-input" name="seed_input" rows="5" placeholder='{"target":"fixture://junos-mx","method":"fake","profile":"juniper_junos"}'></textarea>
          </div>
          <fieldset class="task-list">
            <legend>Optional tasks</legend>
            <div class="task-options">
              ${discoveryTasks.map((task) => `
                <label>
                  <input type="checkbox" name="tasks" value="${escapeHTML(task)}">
                  ${escapeHTML(task)}
                </label>
              `).join("")}
            </div>
          </fieldset>
        </div>
        <div class="form-actions">
          <button type="submit">Create plan</button>
          <span class="muted" id="plan-message">Planner returns suggestions only.</span>
        </div>
      </form>
      <section class="detail-panel plan-result" id="plan-result">
        <div class="empty-state">Submit a seed target to review suggested read-only steps.</div>
      </section>
    </section>
  `;

  document.getElementById("discovery-plan-form").addEventListener("submit", submitDiscoveryPlan);
}

async function submitDiscoveryPlan(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const button = form.querySelector("button");
  const message = document.getElementById("plan-message");
  const resultPanel = document.getElementById("plan-result");
  const formData = new FormData(form);

  const request = {
    target: String(formData.get("target") || "").trim(),
    method: String(formData.get("method") || "").trim(),
    profile: String(formData.get("profile") || "").trim(),
    tasks: formData.getAll("tasks"),
  };
  const seedInputRaw = String(formData.get("seed_input") || "").trim();
  if (seedInputRaw) {
    try {
      request.seed_input = JSON.parse(seedInputRaw);
    } catch (error) {
      message.textContent = "Seed input must be valid JSON.";
      resultPanel.innerHTML = `<div class="error-state">Seed input must be a JSON object.</div>`;
      return;
    }
  }

  button.disabled = true;
  message.textContent = "Creating safe discovery plan...";
  resultPanel.innerHTML = `<div class="empty-state">Creating plan...</div>`;

  try {
    const payload = await apiPost("/api/v1/discovery-plans", request);
    const plan = payload?.data?.discovery_plan;
    message.textContent = "Plan created. Human approval is required before execution.";
    renderDiscoveryPlan(plan);
  } catch (error) {
    message.textContent = error.message;
    resultPanel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  } finally {
    button.disabled = false;
  }
}

function renderDiscoveryPlan(plan) {
  const resultPanel = document.getElementById("plan-result");
  if (!resultPanel || !plan) {
    return;
  }
  const steps = plan.steps || [];
  resultPanel.innerHTML = `
    <div class="approval-banner">
      <strong>Human approval required</strong>
      <span>Execution allowed: ${plan.execution_allowed === true ? "yes" : "no"}. This UI does not execute plans.</span>
    </div>
    <div class="detail-grid compact">
      <div class="metric">
        <small>Approval required</small>
        <strong>${plan.approval_required === true ? "yes" : "no"}</strong>
      </div>
      <div class="metric">
        <small>Execution allowed</small>
        <strong>${plan.execution_allowed === true ? "yes" : "no"}</strong>
      </div>
      <div class="metric">
        <small>Suggested steps</small>
        <strong>${steps.length}</strong>
      </div>
    </div>
    ${plan.warnings?.length ? `
      <span class="code-block-label">Planner warnings</span>
      <ul class="edge-list">
        ${plan.warnings.map((warning) => `<li><span>${escapeHTML(warning)}</span></li>`).join("")}
      </ul>
    ` : ""}
    <span class="code-block-label">Suggested steps</span>
    ${steps.length === 0 ? `<p class="muted">No steps suggested.</p>` : `
      <ul class="plan-step-list">
        ${steps.map((step) => `
          <li class="plan-step">
            <div>
              <strong>${escapeHTML(step.task || "task")}</strong>
              <span class="risk-badge">${escapeHTML(step.risk_level || "risk unknown")}</span>
            </div>
            <dl>
              <dt>Target</dt>
              <dd>${escapeHTML(step.target || "unknown")}</dd>
              <dt>Method</dt>
              <dd>${escapeHTML(step.method || "unknown")}</dd>
              <dt>Profile</dt>
              <dd>${escapeHTML(step.profile || "unknown")}</dd>
              <dt>Reason</dt>
              <dd>${escapeHTML(step.reason || "not provided")}</dd>
              <dt>Expected evidence</dt>
              <dd>${escapeHTML(step.expected_evidence || "not provided")}</dd>
            </dl>
          </li>
        `).join("")}
      </ul>
    `}
    <span class="code-block-label">Seed input</span>
    <pre class="code-block">${escapeHTML(JSON.stringify(plan.seed_input || {}, null, 2))}</pre>
  `;
}

function renderArchitectureSeedsView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Architecture seeding</p>
        <h1>Seed planning context</h1>
        <p>Submit known network context for planning. Seeded hints are context, not observed proof, and this page does not trigger discovery.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <section class="plan-layout">
      <form class="form-panel" id="architecture-seed-form">
        <div class="readonly-note">
          Seeded hints are context, not observed proof. They are stored as user_seeded facts with low confidence and cannot execute discovery.
        </div>
        <div class="form-grid">
          <div class="field">
            <label for="seed-network-type">Organization/network type</label>
            <input id="seed-network-type" name="organization_network_type" placeholder="service_provider">
          </div>
          <div class="field">
            <label for="seed-asns">Known ASNs</label>
            <textarea id="seed-asns" name="known_asns" rows="4" placeholder="65000, 65001"></textarea>
          </div>
          <div class="field">
            <label for="seed-route-reflectors">Known route reflectors</label>
            <textarea id="seed-route-reflectors" name="known_route_reflectors" rows="4" placeholder="rr1.example.net"></textarea>
          </div>
          <div class="field">
            <label for="seed-vendors">Known vendors</label>
            <textarea id="seed-vendors" name="known_vendors" rows="4" placeholder="juniper, cisco"></textarea>
          </div>
          <div class="field">
            <label for="seed-ems">Known EMS systems</label>
            <textarea id="seed-ems" name="known_ems_systems" rows="4" placeholder="ems-a"></textarea>
          </div>
          <div class="field">
            <label for="seed-services">Known services</label>
            <textarea id="seed-services" name="known_services" rows="4" placeholder="l3vpn, internet"></textarea>
          </div>
          <div class="field">
            <label for="seed-regions">Known regions/markets</label>
            <textarea id="seed-regions" name="known_regions_markets" rows="4" placeholder="nyc, dfw"></textarea>
          </div>
        </div>
        <div class="form-actions">
          <button type="submit">Save seed hints</button>
          <span class="muted" id="architecture-seed-message">At least one hint is required.</span>
        </div>
      </form>
      <section class="detail-panel plan-result" id="architecture-seed-result">
        <div class="empty-state">Submit architecture hints to see the stored user_seeded facts.</div>
      </section>
    </section>
  `;

  document.getElementById("architecture-seed-form").addEventListener("submit", submitArchitectureSeed);
}

async function submitArchitectureSeed(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const button = form.querySelector("button");
  const message = document.getElementById("architecture-seed-message");
  const resultPanel = document.getElementById("architecture-seed-result");
  const formData = new FormData(form);
  const request = {
    organization_network_type: String(formData.get("organization_network_type") || "").trim(),
    known_asns: splitSeedList(formData.get("known_asns")),
    known_route_reflectors: splitSeedList(formData.get("known_route_reflectors")),
    known_vendors: splitSeedList(formData.get("known_vendors")),
    known_ems_systems: splitSeedList(formData.get("known_ems_systems")),
    known_services: splitSeedList(formData.get("known_services")),
    known_regions_markets: splitSeedList(formData.get("known_regions_markets")),
  };

  button.disabled = true;
  message.textContent = "Saving architecture seed hints...";
  resultPanel.innerHTML = `<div class="empty-state">Saving seed hints...</div>`;

  try {
    const payload = await apiPost("/api/v1/architecture-seeds", request);
    const seed = payload?.data?.architecture_seed;
    message.textContent = "Seed hints stored as user_seeded context.";
    renderArchitectureSeed(seed);
  } catch (error) {
    message.textContent = error.message;
    resultPanel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  } finally {
    button.disabled = false;
  }
}

function renderArchitectureSeed(seed) {
  const resultPanel = document.getElementById("architecture-seed-result");
  if (!resultPanel || !seed) {
    return;
  }
  const asset = seed.asset || {};
  const facts = seed.facts || [];
  resultPanel.innerHTML = `
    <div class="approval-banner">
      <strong>Context only</strong>
      <span>${escapeHTML(seed.warning || "Seeded hints are context, not observed proof.")}</span>
    </div>
    <div class="detail-grid compact">
      <div class="metric">
        <small>Asset type</small>
        <strong>${escapeHTML(asset.type || "architecture_context")}</strong>
      </div>
      <div class="metric">
        <small>State</small>
        <strong>${stateBadge(asset.state || "user_seeded")}</strong>
      </div>
      <div class="metric">
        <small>Confidence</small>
        <strong>${escapeHTML(confidenceLabel(asset))}</strong>
      </div>
    </div>
    <span class="code-block-label">Identity key</span>
    <pre class="code-block">${escapeHTML(asset.identity_key || "architecture:seed:default")}</pre>
    <span class="code-block-label">Seeded facts</span>
    ${seededFactsMarkup(facts)}
  `;
}

function seededFactsMarkup(facts) {
  if (facts.length === 0) {
    return `<p class="muted">No facts were returned for this seed request.</p>`;
  }
  return `
    <ul class="edge-list">
      ${facts.map((fact) => `
        <li>
          <strong>${escapeHTML(fact.name || "seeded_fact")}</strong>
          <span>${escapeHTML(factValueLabel(fact.value))}</span>
          <small>${escapeHTML(confidenceLabel(fact))}</small>
          <small>source ${escapeHTML(fact.source || "user_seeded")} / state ${escapeHTML(fact.state || "user_seeded")}</small>
        </li>
      `).join("")}
    </ul>
  `;
}

function splitSeedList(value) {
  return String(value || "")
    .split(/[\n,]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function graphLegendMarkup() {
  return `
    <div class="graph-legend" aria-label="Graph legend">
      <span><i class="legend-root"></i> Root asset</span>
      <span><i class="legend-peer"></i> Related asset</span>
      <span>Edge labels show relationship type and confidence.</span>
    </div>
  `;
}

async function renderGraphView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Graph</p>
        <h1>Asset graph</h1>
        <p>Load a small asset neighborhood from the API, inspect relationships, and review edge confidence.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <form class="form-panel" id="graph-form">
      <div class="form-grid">
        <div class="field">
          <label for="asset-select">Known assets</label>
          <select id="asset-select" name="asset_select">
            <option value="">Loading assets...</option>
          </select>
        </div>
        <div class="field">
          <label for="asset-id">Asset ID</label>
          <input id="asset-id" name="asset_id" placeholder="asset UUID or ID" required>
        </div>
        <div class="field">
          <label for="graph-depth">Depth</label>
          <select id="graph-depth" name="depth">
            <option value="1">1 hop (focused)</option>
            <option value="2">2 hops (expanded)</option>
          </select>
        </div>
      </div>
      <div class="form-actions">
        <button type="submit">Load graph</button>
        <span class="muted" id="graph-message">Choose an asset or paste an asset ID. Depth is capped at 2 hops for readability.</span>
      </div>
    </form>
    <section class="graph-layout">
      <div class="graph-panel" id="graph-canvas">
        ${graphLegendMarkup()}
        <div class="empty-state">Select an asset to render its graph.</div>
      </div>
      <aside class="detail-panel graph-detail" id="graph-detail">
        <div class="empty-state">Click a node to show asset details.</div>
      </aside>
    </section>
    ${evidenceDrawerMarkup()}
  `;

  document.getElementById("graph-form").addEventListener("submit", submitGraphForm);
  document.getElementById("asset-select").addEventListener("change", (event) => {
    const selectedID = event.currentTarget.value;
    if (selectedID) {
      document.getElementById("asset-id").value = selectedID;
    }
  });
  setupEvidenceDrawer();
  await loadAssetOptions();
}

function evidenceDrawerMarkup() {
  return `
    <div class="drawer-backdrop hidden" id="evidence-drawer-backdrop"></div>
    <aside class="evidence-drawer hidden" id="evidence-drawer" aria-label="Evidence drawer" aria-live="polite">
      <div class="drawer-header">
        <div>
          <p class="eyebrow">Read-only evidence</p>
          <h2>Evidence</h2>
        </div>
        <button class="secondary" type="button" id="close-evidence-drawer">Close</button>
      </div>
      <div id="evidence-drawer-body" class="drawer-body">
        <div class="empty-state">Open evidence from a fact or relationship.</div>
      </div>
    </aside>
  `;
}

async function loadAssetOptions() {
  const select = document.getElementById("asset-select");
  const input = document.getElementById("asset-id");
  const message = document.getElementById("graph-message");
  try {
    const payload = await apiGet("/api/v1/assets?limit=100");
    const assets = payload?.data?.assets || [];
    if (assets.length === 0) {
      select.innerHTML = `<option value="">No assets available</option>`;
      message.textContent = "Create or persist assets before loading a graph.";
      return;
    }
    select.innerHTML = `
      <option value="">Select an asset...</option>
      ${assets.map((asset) => `
        <option value="${escapeHTML(asset.id)}">${escapeHTML(assetLabel(asset))}</option>
      `).join("")}
    `;
    input.value = assets[0].id;
    await loadGraph(assets[0].id, document.getElementById("graph-depth")?.value || "1");
  } catch (error) {
    select.innerHTML = `<option value="">Asset list unavailable</option>`;
    message.textContent = error.message;
  }
}

async function submitGraphForm(event) {
  event.preventDefault();
  const formData = new FormData(event.currentTarget);
  const assetID = String(formData.get("asset_id") || "").trim();
  const depth = String(formData.get("depth") || "1").trim();
  if (!assetID) {
    document.getElementById("graph-message").textContent = "Asset ID is required.";
    return;
  }
  await loadGraph(assetID, depth);
}

async function loadGraph(assetID, depth = "1") {
  const canvas = document.getElementById("graph-canvas");
  const detail = document.getElementById("graph-detail");
  const message = document.getElementById("graph-message");
  canvas.innerHTML = `${graphLegendMarkup()}<div class="empty-state">Loading graph...</div>`;
  detail.innerHTML = `<div class="empty-state">Click a node to show asset details.</div>`;
  message.textContent = `Loading graph for ${assetID}...`;

  try {
    const params = new URLSearchParams({ depth });
    const payload = await apiGet(`/api/v1/assets/${encodeURIComponent(assetID)}/graph?${params.toString()}`);
    const graph = payload?.data?.graph || { nodes: [], edges: [] };
    renderGraph(graph);
    message.textContent = `${graph.nodes?.length || 0} nodes, ${graph.edges?.length || 0} edges loaded.`;
  } catch (error) {
    canvas.innerHTML = `${graphLegendMarkup()}<div class="error-state">${escapeHTML(error.message)}</div>`;
    message.textContent = error.message;
  }
}

function renderGraph(graph) {
  const canvas = document.getElementById("graph-canvas");
  const detail = document.getElementById("graph-detail");
  const nodes = graph.nodes || [];
  const edges = graph.edges || [];
  if (nodes.length === 0) {
    canvas.innerHTML = `${graphLegendMarkup()}<div class="empty-state">Graph has no nodes.</div>`;
    return;
  }

  const width = 860;
  const height = 520;
  const centerX = width / 2;
  const centerY = height / 2;
  const radius = Math.min(width, height) * 0.32;
  const positions = new Map();
  nodes.forEach((node, index) => {
    if (index === 0) {
      positions.set(node.id, { x: centerX, y: centerY });
      return;
    }
    const angle = ((index - 1) / Math.max(nodes.length - 1, 1)) * Math.PI * 2 - Math.PI / 2;
    positions.set(node.id, {
      x: centerX + Math.cos(angle) * radius,
      y: centerY + Math.sin(angle) * radius,
    });
  });

  const edgeMarkup = edges.map((edge) => {
    const source = positions.get(edge.source);
    const target = positions.get(edge.target);
    if (!source || !target) {
      return "";
    }
    const midX = (source.x + target.x) / 2;
    const midY = (source.y + target.y) / 2;
    return `
      <g class="graph-edge">
        <line x1="${source.x}" y1="${source.y}" x2="${target.x}" y2="${target.y}"></line>
        <text x="${midX}" y="${midY - 8}">${escapeHTML(edge.relationship_type || "related")}</text>
        <text class="edge-confidence" x="${midX}" y="${midY + 12}">${escapeHTML(confidenceLabel(edge))}</text>
      </g>
    `;
  }).join("");

  const nodeMarkup = nodes.map((node, index) => {
    const position = positions.get(node.id);
    const rootClass = index === 0 ? " root" : "";
    return `
      <g class="graph-node${rootClass}" data-node-id="${escapeHTML(node.id)}" tabindex="0" role="button" aria-label="${escapeHTML(assetLabel(node))}">
        <circle cx="${position.x}" cy="${position.y}" r="${index === 0 ? 42 : 34}"></circle>
        <text x="${position.x}" y="${position.y - 2}">${escapeHTML(truncate(assetLabel(node), 18))}</text>
        <text class="node-type" x="${position.x}" y="${position.y + 16}">${escapeHTML(node.type || "asset")}</text>
      </g>
    `;
  }).join("");

  canvas.innerHTML = `
    ${graphLegendMarkup()}
    <svg class="graph-svg" viewBox="0 0 ${width} ${height}" role="img" aria-label="Asset relationship graph with ${nodes.length} nodes and ${edges.length} edges">
      ${edgeMarkup}
      ${nodeMarkup}
    </svg>
  `;

  const nodesByID = new Map(nodes.map((node) => [node.id, node]));
  for (const element of canvas.querySelectorAll(".graph-node")) {
    element.addEventListener("click", () => selectGraphNode(nodesByID.get(element.dataset.nodeId), edges));
    element.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        selectGraphNode(nodesByID.get(element.dataset.nodeId), edges);
      }
    });
  }
  selectGraphNode(nodes[0], edges);
}

function selectGraphNode(node, edges) {
  if (!node) {
    return;
  }
  const detail = document.getElementById("graph-detail");
  const relatedEdges = edges.filter((edge) => edge.source === node.id || edge.target === node.id);
  const facts = node.facts || [];
  detail.innerHTML = `
    <p class="eyebrow">Selected asset</p>
    <h2>${escapeHTML(assetLabel(node))}</h2>
    <div class="detail-grid compact">
      <div class="metric">
        <small>Type</small>
        <strong>${escapeHTML(node.type || "unknown")}</strong>
      </div>
      <div class="metric">
        <small>State</small>
        <strong>${escapeHTML(node.state || "unknown")}</strong>
      </div>
      <div class="metric">
        <small>Confidence</small>
        <strong>${escapeHTML(confidenceLabel(node))}</strong>
      </div>
    </div>
    <span class="code-block-label">Identity key</span>
    <pre class="code-block">${escapeHTML(node.identity_key || node.id)}</pre>
    <span class="code-block-label">Facts</span>
    ${facts.length === 0 ? `<p class="muted">No facts included in this graph response.</p>` : `
      <ul class="edge-list">
        ${facts.map((fact) => `
          <li>
            <strong>${escapeHTML(fact.name || "fact")}</strong>
            <span>${escapeHTML(factValueLabel(fact.value))}</span>
            <small>${escapeHTML(confidenceLabel(fact))}</small>
            ${evidenceButton(fact.evidence_id)}
          </li>
        `).join("")}
      </ul>
    `}
    <span class="code-block-label">Relationships</span>
    ${relatedEdges.length === 0 ? `<p class="muted">No relationships in this graph.</p>` : `
      <ul class="edge-list">
        ${relatedEdges.map((edge) => `
          <li>
            <strong>${escapeHTML(edge.relationship_type || "related")}</strong>
            <span>${escapeHTML(edge.source)} -> ${escapeHTML(edge.target)}</span>
            <small>${escapeHTML(confidenceLabel(edge))}</small>
            ${evidenceButton(edge.evidence_id)}
          </li>
        `).join("")}
      </ul>
    `}
  `;
  attachEvidenceButtons(detail);
}


async function renderIdentityReviewView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Identity Review</p>
        <h1>Review parser-derived identity candidates</h1>
        <p>Inspect evidence-backed identity clues before accepting, rejecting, deferring, or requesting more evidence. Review actions are non-destructive and never silently merge assets.</p>
      </div>
      <a class="button secondary" href="/api/v1/identity-candidates/handoff-report">Handoff API</a>
    </section>
    <form class="form-panel" id="identity-filter-form">
      <div class="form-grid identity-filter-grid">
        <div class="field"><label for="identity-review-state">Review state</label><select id="identity-review-state" name="review_state">
          <option value="pending">pending</option><option value="">all states</option><option value="accepted">accepted</option><option value="rejected">rejected</option><option value="deferred">deferred</option><option value="more_evidence_requested">more evidence requested</option><option value="auto_accepted">auto accepted</option>
        </select></div>
        <div class="field"><label for="identity-strength">Strength</label><select id="identity-strength" name="strength"><option value="">all strengths</option><option value="strong">strong</option><option value="provisional">provisional</option><option value="weak">weak</option></select></div>
        <div class="field"><label for="identity-asset-type">Asset type</label><input id="identity-asset-type" name="asset_type" placeholder="device"></div>
        <div class="field"><label for="identity-run-id">Discovery run ID</label><input id="identity-run-id" name="discovery_run_id" placeholder="uuid"></div>
        <div class="field"><label for="identity-proposed-asset-id">Proposed asset ID</label><input id="identity-proposed-asset-id" name="proposed_asset_id" placeholder="uuid"></div>
        <div class="field"><label for="identity-search">Search</label><input id="identity-search" name="search" placeholder="hostname, serial, MAC, identity key"></div>
      </div>
      <div class="form-actions">
        <button type="submit">Apply filters</button>
        <button class="secondary" type="button" id="clear-identity-filters">Clear</button>
        <span class="muted" id="identity-result-message">Pending candidates are shown by default. Asset type, proposed asset, and search filters are applied in this UI.</span>
      </div>
    </form>
    <section class="identity-review-layout">
      <div class="table-panel" id="identity-candidates-panel"><div class="empty-state">Loading identity candidates...</div></div>
      <aside class="detail-panel identity-detail-panel" id="identity-candidate-detail"><div class="empty-state">Select a candidate to inspect provenance and review actions.</div></aside>
    </section>
    ${evidenceDrawerMarkup()}
  `;
  setupEvidenceDrawer();
  const query = new URLSearchParams((location.hash.split("?")[1] || ""));
  for (const [key, value] of query.entries()) {
    const field = document.querySelector(`#identity-filter-form [name="${CSS.escape(key)}"]`);
    if (field) field.value = value;
  }
  document.getElementById("identity-filter-form").addEventListener("submit", (event) => { event.preventDefault(); loadIdentityCandidatesFromFilters(); });
  document.getElementById("clear-identity-filters").addEventListener("click", () => { location.hash = "#/identity-review"; void loadIdentityCandidates({ review_state: "pending" }); });
  await loadIdentityCandidatesFromFilters();
}

async function loadIdentityCandidatesFromFilters() {
  const form = document.getElementById("identity-filter-form");
  const params = new URLSearchParams(new FormData(form));
  for (const [key, value] of Array.from(params.entries())) if (!String(value).trim()) params.delete(key);
  if (!params.has("review_state")) params.set("review_state", "pending");
  history.replaceState(null, "", `#/identity-review?${params.toString()}`);
  await loadIdentityCandidates(Object.fromEntries(params.entries()));
}

async function loadIdentityCandidates(filters) {
  const panel = document.getElementById("identity-candidates-panel");
  panel.innerHTML = `<div class="empty-state">Loading identity candidates...</div>`;
  try {
    const serverParams = new URLSearchParams();
    for (const key of ["review_state", "strength", "discovery_run_id"]) if (filters[key]) serverParams.set(key, filters[key]);
    const endpoint = filters.review_state === "pending" ? "/api/v1/identity-candidates/review-queue" : "/api/v1/identity-candidates";
    const payload = await apiGet(`${endpoint}${serverParams.toString() ? `?${serverParams}` : ""}`);
    let candidates = payload?.data?.identity_candidates || [];
    candidates = filterIdentityCandidatesClientSide(candidates, filters);
    renderIdentityCandidateList(candidates);
    if (candidates[0]) renderIdentityCandidateDetail(candidates[0]);
  } catch (error) {
    panel.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

function filterIdentityCandidatesClientSide(candidates, filters) {
  const includes = (value, needle) => String(value || "").toLowerCase().includes(String(needle || "").toLowerCase());
  return candidates.filter((candidate) => {
    if (filters.asset_type && candidate.asset_type !== filters.asset_type) return false;
    if (filters.proposed_asset_id && candidate.proposed_asset_id !== filters.proposed_asset_id) return false;
    if (filters.search) {
      const haystack = [candidate.candidate_identity_key, candidate.vendor, candidate.model, candidate.serial, candidate.system_mac, candidate.hostname, candidate.parser_name, candidate.evidence_id, candidate.discovery_run_id, candidate.proposed_asset_id].join(" ");
      if (!includes(haystack, filters.search)) return false;
    }
    return true;
  });
}

function renderIdentityCandidateList(candidates) {
  const panel = document.getElementById("identity-candidates-panel");
  if (!candidates.length) { panel.innerHTML = `<div class="empty-state">No identity candidates match these filters.</div>`; return; }
  panel.innerHTML = `<table class="table identity-table"><thead><tr><th>Candidate</th><th>Identity clues</th><th>Provenance</th><th>Created</th><th></th></tr></thead><tbody>${candidates.map((candidate) => `
    <tr>
      <td><strong>${escapeHTML(candidate.candidate_identity_key)}</strong><br>${identityValueBadge(candidate.strength)} ${reviewStateBadge(candidate.review_state)}<br><small>${escapeHTML(candidate.asset_type || "asset")} · ${escapeHTML(confidenceLabel(candidate))}</small></td>
      <td><small>Vendor</small> ${escapeHTML(candidate.vendor || "—")}<br><small>Model</small> ${escapeHTML(candidate.model || "—")}<br><small>Serial</small> ${escapeHTML(candidate.serial || "—")}<br><small>MAC</small> ${escapeHTML(candidate.system_mac || "—")}<br><small>Hostname</small> ${escapeHTML(candidate.hostname || "—")}</td>
      <td><small>Parser</small> ${escapeHTML(candidate.parser_name || "—")}<br><small>Run</small> ${discoveryRunLink(candidate.discovery_run_id)}<br><small>Evidence</small> ${evidenceButton(candidate.evidence_id)}<br><small>Proposed asset</small> ${assetLink(candidate.proposed_asset_id)}</td>
      <td>${escapeHTML(formatDate(candidate.created_at))}</td>
      <td><button type="button" data-candidate-id="${escapeHTML(candidate.id)}">Review</button></td>
    </tr>`).join("")}</tbody></table>`;
  for (const button of panel.querySelectorAll("[data-candidate-id]")) button.addEventListener("click", () => renderIdentityCandidateDetail(candidates.find((candidate) => candidate.id === button.dataset.candidateId)));
  attachEvidenceButtons(panel);
}

function renderIdentityCandidateDetail(candidate) {
  const detail = document.getElementById("identity-candidate-detail");
  if (!candidate) return;
  detail.innerHTML = `
    <div class="identity-detail-content">
      <p class="eyebrow">Candidate detail</p><h2>${escapeHTML(candidate.candidate_identity_key)}</h2>
      <p class="readonly-note">Review records an explicit decision only. Truthwatcher does not silently merge assets or rewrite canonical identity.</p>
      <div class="detail-grid compact">
        <div class="metric"><small>Strength</small><strong>${identityValueBadge(candidate.strength)}</strong></div>
        <div class="metric"><small>Review state</small><strong>${reviewStateBadge(candidate.review_state)}</strong></div>
        <div class="metric"><small>Confidence</small><strong>${escapeHTML(confidenceLabel(candidate))}</strong></div>
        <div class="metric"><small>Reason</small><strong>${escapeHTML(candidate.reason || "—")}</strong></div>
      </div>
      <span class="code-block-label">Evidence and provenance</span>
      <ul class="edge-list"><li><strong>Discovery run</strong><span>${discoveryRunLink(candidate.discovery_run_id)}</span></li><li><strong>Evidence</strong><span>${evidenceButton(candidate.evidence_id)}</span></li><li><strong>Parser</strong><span>${escapeHTML(candidate.parser_name || "—")}</span></li><li><strong>Proposed asset</strong><span>${assetLink(candidate.proposed_asset_id)}</span></li></ul>
      <span class="code-block-label">Metadata JSON</span><pre class="code-block">${escapeHTML(formatJSON(candidate.metadata))}</pre>
      ${identityReviewFormMarkup(candidate)}
    </div>`;
  attachEvidenceButtons(detail);
  document.getElementById("identity-review-action-form")?.addEventListener("submit", (event) => submitIdentityReview(event, candidate));
}

function identityReviewFormMarkup(candidate) {
  return `<form class="form-panel identity-action-form" id="identity-review-action-form">
    <div class="field"><label for="identity-action">Review action</label><select id="identity-action" name="action"><option value="accept">accept</option><option value="reject">reject</option><option value="defer">defer</option><option value="request_more_evidence">request more evidence</option></select></div>
    <div class="field"><label for="identity-reviewer">Reviewer</label><input id="identity-reviewer" name="reviewer" value="ui:identity-review" required></div>
    <div class="field"><label for="identity-rationale">Review note / reason</label><textarea id="identity-rationale" name="rationale" rows="3" placeholder="Required for reject, defer, and request more evidence"></textarea></div>
    <p class="muted">Accept requires an existing proposed asset ID. Current proposed asset: ${assetLink(candidate.proposed_asset_id)}</p>
    <div class="form-actions"><button type="submit">Submit review</button><span class="muted" id="identity-action-message"></span></div>
  </form>`;
}

async function submitIdentityReview(event, candidate) {
  event.preventDefault();
  const form = event.currentTarget;
  const action = form.elements.action.value;
  const rationale = form.elements.rationale.value.trim();
  const message = document.getElementById("identity-action-message");
  if (["reject", "defer", "request_more_evidence"].includes(action) && !rationale) { message.textContent = "A short review note is required for this action."; return; }
  if (action === "accept" && !candidate.proposed_asset_id) { message.textContent = "Accept requires a proposed_asset_id; request more evidence or defer instead."; return; }
  try {
    const payload = await apiPost(`/api/v1/identity-candidates/${encodeURIComponent(candidate.id)}/review`, { action, reviewer: form.elements.reviewer.value.trim(), rationale, metadata: { source: "embedded_identity_review_ui" } });
    const review = payload?.data?.identity_candidate_review;
    document.getElementById("identity-result-message").textContent = `Recorded ${review?.action || action}; resulting state ${review?.resulting_review_state || "updated"}.`;
    await loadIdentityCandidatesFromFilters();
  } catch (error) { message.textContent = error.message; }
}

function identityValueBadge(value) { return `<span class="identity-badge ${escapeHTML(value || "unknown")}">${escapeHTML(value || "unknown")}</span>`; }
function reviewStateBadge(value) { return `<span class="review-state-badge ${escapeHTML(value || "unknown")}">${escapeHTML(value || "unknown")}</span>`; }
function discoveryRunLink(id) { return id ? `<a href="#/discovery-runs/${encodeURIComponent(id)}">${escapeHTML(shortID(id))}</a>` : "—"; }
function assetLink(id) { return id ? `<a href="#/assets/${encodeURIComponent(id)}">${escapeHTML(shortID(id))}</a>` : "—"; }
function formatJSON(value) { try { return JSON.stringify(typeof value === "string" ? JSON.parse(value) : (value || {}), null, 2); } catch { return String(value || "{}"); } }

function setupEvidenceDrawer() {
  const drawer = document.getElementById("evidence-drawer");
  const backdrop = document.getElementById("evidence-drawer-backdrop");
  const close = document.getElementById("close-evidence-drawer");
  if (!drawer || !backdrop || !close) {
    return;
  }
  close.addEventListener("click", closeEvidenceDrawer);
  backdrop.addEventListener("click", closeEvidenceDrawer);
  window.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && !drawer.classList.contains("hidden")) {
      closeEvidenceDrawer();
    }
  });
}

function attachEvidenceButtons(scope) {
  for (const button of scope.querySelectorAll("[data-evidence-id]")) {
    button.addEventListener("click", () => openEvidenceDrawer(button.dataset.evidenceId));
  }
}

async function openEvidenceDrawer(evidenceID) {
  const drawer = document.getElementById("evidence-drawer");
  const backdrop = document.getElementById("evidence-drawer-backdrop");
  const body = document.getElementById("evidence-drawer-body");
  if (!drawer || !backdrop || !body || !evidenceID) {
    return;
  }

  drawer.classList.remove("hidden");
  backdrop.classList.remove("hidden");
  body.innerHTML = `<div class="empty-state">Loading evidence...</div>`;

  try {
    const payload = await apiGet(`/api/v1/evidence/${encodeURIComponent(evidenceID)}`);
    const evidence = payload?.data?.evidence;
    renderEvidenceDrawer(evidence);
  } catch (error) {
    body.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

function closeEvidenceDrawer() {
  document.getElementById("evidence-drawer")?.classList.add("hidden");
  document.getElementById("evidence-drawer-backdrop")?.classList.add("hidden");
}

function renderEvidenceDrawer(evidence) {
  const body = document.getElementById("evidence-drawer-body");
  if (!body || !evidence) {
    return;
  }
  body.innerHTML = `
    <p class="readonly-note">Evidence is read-only. Raw output is preserved exactly as stored.</p>
    <div class="detail-grid compact">
      <div class="metric">
        <small>Method</small>
        <strong>${escapeHTML(evidence.method || "unknown")}</strong>
      </div>
      <div class="metric">
        <small>Target</small>
        <strong>${escapeHTML(evidence.target || "unknown")}</strong>
      </div>
      <div class="metric">
        <small>Command/API</small>
        <strong>${escapeHTML(evidence.command_or_api || "unknown")}</strong>
      </div>
      <div class="metric">
        <small>Timestamp</small>
        <strong>${escapeHTML(formatDate(evidence.collected_at))}</strong>
      </div>
      <div class="metric">
        <small>Hash</small>
        <strong>${escapeHTML(evidence.raw_output_hash || "missing")}</strong>
      </div>
    </div>
    <div class="drawer-actions">
      <button type="button" id="copy-evidence-output">Copy raw output</button>
      <span class="muted" id="copy-evidence-message">Copies the stored raw output only.</span>
    </div>
    <span class="code-block-label">Raw output</span>
    <pre class="code-block raw-output" id="evidence-raw-output">${escapeHTML(evidence.raw_output || "")}</pre>
  `;
  document.getElementById("copy-evidence-output").addEventListener("click", () => copyEvidenceRawOutput(evidence.raw_output || ""));
}

function renderAskView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">Ask Truthwatcher</p>
        <h1>Read-only agent shell</h1>
        <p>Ask deterministic questions about known assets, asset evidence, or discovery runs. This shell does not call an external LLM and cannot execute discovery.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <section class="ask-layout">
      <div class="chat-panel">
        <div class="readonly-note">
          Deterministic canned responses only. No network actions, discovery execution, or external model calls.
        </div>
        <div class="prompt-chips" aria-label="Example prompts">
          <button class="secondary" type="button" data-agent-prompt="list known assets">List known assets</button>
          <button class="secondary" type="button" data-agent-prompt="explain asset evidence">Explain asset evidence</button>
          <button class="secondary" type="button" data-agent-prompt="summarize discovery run">Summarize discovery run</button>
          <button class="secondary" type="button" data-agent-prompt="what do we know about asset-a">What do we know about X</button>
          <button class="secondary" type="button" data-agent-prompt="show neighbors for asset-a">Show neighbors for X</button>
          <button class="secondary" type="button" data-agent-prompt="why do we believe asset-a exists">Why does X exist</button>
          <button class="secondary" type="button" data-agent-prompt="what is unknown">What is unknown</button>
        </div>
        <div class="chat-history" id="chat-history" aria-live="polite"></div>
        <form class="chat-form" id="agent-form">
          <label class="sr-only" for="agent-message">Message</label>
          <textarea id="agent-message" name="message" rows="3" placeholder="Try: list known assets"></textarea>
          <button type="submit">Ask</button>
        </form>
      </div>
      <aside class="detail-panel">
        <p class="eyebrow">Capabilities</p>
        <ul class="capability-list">
          <li>list known assets</li>
          <li>explain asset evidence</li>
          <li>summarize discovery run</li>
          <li>what do we know about X</li>
          <li>show neighbors for X</li>
          <li>why do we believe X exists</li>
          <li>what is unknown</li>
        </ul>
        <p class="muted">Conversation history is stored in this browser only.</p>
      </aside>
    </section>
  `;

  document.getElementById("agent-form").addEventListener("submit", submitAgentMessage);
  for (const button of document.querySelectorAll("[data-agent-prompt]")) {
    button.addEventListener("click", () => {
      document.getElementById("agent-message").value = button.dataset.agentPrompt;
      document.getElementById("agent-message").focus();
    });
  }
  renderAgentHistory();
}

async function submitAgentMessage(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const button = form.querySelector("button");
  const input = document.getElementById("agent-message");
  const message = input.value.trim();
  if (!message) {
    return;
  }

  appendAgentHistory({ role: "user", message, at: new Date().toISOString() });
  input.value = "";
  button.disabled = true;
  renderAgentHistory("Thinking deterministically...");

  try {
    const payload = await apiPost("/api/v1/agent/messages", { message });
    const response = payload?.data?.agent_message;
    appendAgentHistory({
      role: "truthwatcher",
      message: response?.message || "No response.",
      intent: response?.intent || "unknown",
      read_only: response?.read_only === true,
      actions: response?.actions || [],
      at: new Date().toISOString(),
    });
  } catch (error) {
    appendAgentHistory({
      role: "truthwatcher",
      message: error.message,
      intent: "error",
      read_only: true,
      actions: [],
      at: new Date().toISOString(),
    });
  } finally {
    button.disabled = false;
    renderAgentHistory();
  }
}

function renderAgentHistory(statusMessage = "") {
  const historyPanel = document.getElementById("chat-history");
  if (!historyPanel) {
    return;
  }
  const history = loadAgentHistory();
  const messages = history.map((item) => `
    <article class="chat-message ${escapeHTML(item.role)}">
      <small>${escapeHTML(item.role)}${item.intent ? ` / ${escapeHTML(item.intent)}` : ""}</small>
      <pre>${escapeHTML(item.message)}</pre>
      ${item.read_only ? `<span class="readonly-badge">read-only</span>` : ""}
      ${item.actions?.length ? `<p class="muted">Actions: ${escapeHTML(item.actions.join(", "))}</p>` : ""}
    </article>
  `).join("");
  historyPanel.innerHTML = `
    ${messages || `<div class="empty-state">No local conversation yet.</div>`}
    ${statusMessage ? `<div class="empty-state">${escapeHTML(statusMessage)}</div>` : ""}
  `;
  historyPanel.scrollTop = historyPanel.scrollHeight;
}

function appendAgentHistory(item) {
  const history = loadAgentHistory();
  history.push(item);
  localStorage.setItem(agentHistoryKey, JSON.stringify(history.slice(-30)));
}

function loadAgentHistory() {
  try {
    const raw = localStorage.getItem(agentHistoryKey);
    if (!raw) {
      return [];
    }
    const history = JSON.parse(raw);
    return Array.isArray(history) ? history : [];
  } catch (error) {
    return [];
  }
}

async function copyEvidenceRawOutput(rawOutput) {
  const message = document.getElementById("copy-evidence-message");
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(rawOutput);
    } else {
      const textArea = document.createElement("textarea");
      textArea.value = rawOutput;
      textArea.setAttribute("readonly", "");
      textArea.style.position = "fixed";
      textArea.style.left = "-9999px";
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
    }
    message.textContent = "Raw output copied.";
  } catch (error) {
    message.textContent = "Copy failed.";
  }
}

function evidenceButton(evidenceID) {
  if (!evidenceID) {
    return `<small>No evidence reference</small>`;
  }
  return `<button class="secondary evidence-link" type="button" data-evidence-id="${escapeHTML(evidenceID)}">Open evidence</button>`;
}

function factValueLabel(value) {
  if (value === undefined || value === null) {
    return "null";
  }
  if (typeof value === "string") {
    return value;
  }
  return JSON.stringify(value);
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

function stateBadge(state) {
  const safeState = escapeHTML(state || "unknown");
  return `<span class="status-pill ${safeState}">${safeState}</span>`;
}

async function renderAboutView() {
  app.innerHTML = `
    <section class="section-header">
      <div>
        <p class="eyebrow">About this system</p>
        <h1>Truthwatcher philosophy and glossary</h1>
        <p>Truthwatcher is an evidence-first network cartography tool. It separates observed evidence, parsed facts, user-provided context, and human-approved action.</p>
      </div>
      <a class="button secondary" href="#/">Back to dashboard</a>
    </section>
    <section class="about-layout">
      <article class="detail-panel">
        <p class="eyebrow">Philosophy</p>
        <h2>Evidence before inference</h2>
        <p class="muted">The app treats raw evidence as the durable record. Facts, relationships, graph views, and planner suggestions must remain explainable by evidence or clearly labeled as seeded context.</p>
        <ul class="principle-list">
          <li><strong>Read-only by default.</strong><span>Collection and planning features prefer safe observation and human approval over automatic execution.</span></li>
          <li><strong>Confidence is visible.</strong><span>Modeled knowledge should carry confidence, source, state, and conflict information instead of pretending every claim is equally true.</span></li>
          <li><strong>Humans own intent.</strong><span>Seeds, reviews, and approvals are explicit human inputs; the system should not silently expand scope or infer permission.</span></li>
        </ul>
      </article>
      <article class="detail-panel" id="system-info-panel">
        <p class="eyebrow">System stats</p>
        <div class="empty-state">Loading CPU, memory, disk, and build information...</div>
      </article>
    </section>
    <section class="detail-panel glossary-panel">
      <p class="eyebrow">Glossary</p>
      <h2>Terms used across the app</h2>
      <dl class="glossary-list">
        ${glossaryTerms.map((term) => `
          <div>
            <dt>${escapeHTML(term.term)}</dt>
            <dd>${escapeHTML(term.definition)}</dd>
          </div>
        `).join("")}
      </dl>
    </section>
  `;
  await loadSystemInfo();
}

const glossaryTerms = [
  { term: "Asset", definition: "A modeled network object such as a device, service, architecture context record, or other entity tracked by Truthwatcher." },
  { term: "Evidence", definition: "Raw observed output or API data collected from an allowed source. Evidence is read-only and remains the primary record." },
  { term: "Fact", definition: "A specific claim about an asset, usually parsed from evidence or created as explicitly user-seeded context." },
  { term: "Relationship", definition: "A typed edge between assets, such as a neighbor, ownership, or service relationship." },
  { term: "Confidence", definition: "A numeric indication of how strongly the system trusts a fact, relationship, or asset identity." },
  { term: "Discovery run", definition: "A bounded collection attempt with seed input, status, timestamps, and linked evidence." },
  { term: "Discovery plan", definition: "A read-only proposal for next collection steps. It is not execution and requires human approval." },
  { term: "Architecture seed", definition: "Known context supplied by a user. Seeds help planning but are labeled as context rather than observed proof." },
  { term: "Provisional identity", definition: "An identity that may need review or merging before it should be treated as authoritative." },
  { term: "Graph", definition: "A neighborhood view that renders assets and relationships with confidence visible on edges." },
];

async function loadSystemInfo() {
  const panel = document.getElementById("system-info-panel");
  if (!panel) {
    return;
  }
  try {
    const payload = await apiGet("/api/v1/system-info");
    const info = payload?.data?.system_info || {};
    panel.innerHTML = systemInfoMarkup(info);
  } catch (error) {
    panel.innerHTML = `<p class="eyebrow">System stats</p><div class="error-state">${escapeHTML(error.message)}</div>`;
  }
}

function systemInfoMarkup(info) {
  const runtime = info.runtime || {};
  const memory = info.memory || {};
  const disk = info.disk || {};
  const build = info.build || {};
  return `
    <p class="eyebrow">System stats</p>
    <h2>${escapeHTML(info.name || "truthwatcher")} ${escapeHTML(info.version || "unknown")}</h2>
    <div class="detail-grid compact">
      <div class="metric"><small>CPUs</small><strong>${escapeHTML(runtime.cpus || "unknown")}</strong></div>
      <div class="metric"><small>Goroutines</small><strong>${escapeHTML(runtime.goroutines || "unknown")}</strong></div>
      <div class="metric"><small>Memory alloc</small><strong>${escapeHTML(formatBytes(memory.alloc_bytes))}</strong></div>
      <div class="metric"><small>Heap sys</small><strong>${escapeHTML(formatBytes(memory.heap_sys_bytes))}</strong></div>
      <div class="metric"><small>Disk used</small><strong>${escapeHTML(formatBytes(disk.used_bytes))}</strong></div>
      <div class="metric"><small>Disk free</small><strong>${escapeHTML(formatBytes(disk.free_bytes))}</strong></div>
    </div>
    <span class="code-block-label">Runtime</span>
    <pre class="code-block">${escapeHTML(`${runtime.go_version || "unknown"} ${runtime.os || "unknown"}/${runtime.arch || "unknown"}`)}</pre>
    <span class="code-block-label">Build details</span>
    <pre class="code-block">${escapeHTML(JSON.stringify({ main_path: build.main_path || "unknown", go_version: build.go_version || "unknown", settings: build.settings || {}, generated_at: info.generated_at || "unknown" }, null, 2))}</pre>
  `;
}

function formatBytes(value) {
  const bytes = Number(value || 0);
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return "0 B";
  }
  const units = ["B", "KiB", "MiB", "GiB", "TiB"];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / (1024 ** index)).toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
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

function assetLabel(asset) {
  if (!asset) {
    return "unknown asset";
  }
  return asset.label || asset.identity_key || asset.serial || asset.id || "unknown asset";
}

function confidenceLabel(item) {
  if (!item || item.confidence === undefined || item.confidence === null) {
    return "confidence unknown";
  }
  const percent = Math.round(Number(item.confidence) * 100);
  const state = item.state ? ` ${item.state}` : "";
  return `${percent}%${state}`;
}

function identityBadge(asset) {
  const strength = asset?.metadata?.identity_strength || "unknown";
  const reason = asset?.metadata?.identity_reason || "";
  return `<span class="identity-badge ${escapeHTML(strength)}" title="${escapeHTML(reason)}">${escapeHTML(strength)}</span>`;
}

function truncate(value, length) {
  const text = String(value || "");
  if (text.length <= length) {
    return text;
  }
  return `${text.slice(0, length - 1)}...`;
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
