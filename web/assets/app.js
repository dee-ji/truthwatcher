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

  if (route === "/assets") {
    await renderAssetsView();
    return;
  }

  if (route.startsWith("/assets/")) {
    await renderAssetDetail(route.split("/").pop());
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
  if (route === "/graph") {
    active = document.querySelector('[data-nav="graph"]');
  }
  if (route === "/ask") {
    active = document.querySelector('[data-nav="ask"]');
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
    const [assetPayload, factsPayload, relationshipsPayload] = await Promise.all([
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}`),
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}/facts?limit=100`),
      apiGet(`/api/v1/assets/${encodeURIComponent(id)}/relationships?limit=100`),
    ]);
    const asset = assetPayload?.data?.asset;
    const facts = factsPayload?.data?.facts || [];
    const relationships = relationshipsPayload?.data?.relationships || [];

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
      </div>
      <div class="form-actions">
        <button type="submit">Load graph</button>
        <span class="muted" id="graph-message">Choose an asset or paste an asset ID.</span>
      </div>
    </form>
    <section class="graph-layout">
      <div class="graph-panel" id="graph-canvas">
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
    await loadGraph(assets[0].id);
  } catch (error) {
    select.innerHTML = `<option value="">Asset list unavailable</option>`;
    message.textContent = error.message;
  }
}

async function submitGraphForm(event) {
  event.preventDefault();
  const formData = new FormData(event.currentTarget);
  const assetID = String(formData.get("asset_id") || "").trim();
  if (!assetID) {
    document.getElementById("graph-message").textContent = "Asset ID is required.";
    return;
  }
  await loadGraph(assetID);
}

async function loadGraph(assetID) {
  const canvas = document.getElementById("graph-canvas");
  const detail = document.getElementById("graph-detail");
  const message = document.getElementById("graph-message");
  canvas.innerHTML = `<div class="empty-state">Loading graph...</div>`;
  detail.innerHTML = `<div class="empty-state">Click a node to show asset details.</div>`;
  message.textContent = `Loading graph for ${assetID}...`;

  try {
    const payload = await apiGet(`/api/v1/assets/${encodeURIComponent(assetID)}/graph`);
    const graph = payload?.data?.graph || { nodes: [], edges: [] };
    renderGraph(graph);
    message.textContent = `${graph.nodes?.length || 0} nodes, ${graph.edges?.length || 0} edges loaded.`;
  } catch (error) {
    canvas.innerHTML = `<div class="error-state">${escapeHTML(error.message)}</div>`;
    message.textContent = error.message;
  }
}

function renderGraph(graph) {
  const canvas = document.getElementById("graph-canvas");
  const detail = document.getElementById("graph-detail");
  const nodes = graph.nodes || [];
  const edges = graph.edges || [];
  if (nodes.length === 0) {
    canvas.innerHTML = `<div class="empty-state">Graph has no nodes.</div>`;
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
    <svg class="graph-svg" viewBox="0 0 ${width} ${height}" role="img" aria-label="Asset relationship graph">
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
