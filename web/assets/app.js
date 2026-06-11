(async function checkAPI() {
  const status = document.getElementById("api-status");
  const version = document.getElementById("api-version");
  const dot = document.getElementById("status-dot");

  try {
    const [readyResponse, versionResponse] = await Promise.all([
      fetch("/readyz", { headers: { Accept: "application/json" } }),
      fetch("/api/v1/version", { headers: { Accept: "application/json" } }),
    ]);

    if (!readyResponse.ok || !versionResponse.ok) {
      throw new Error("API returned an unsuccessful status");
    }

    const versionPayload = await versionResponse.json();
    const appVersion = versionPayload?.data?.version || "unknown";

    status.textContent = "API ready";
    version.textContent = `truthwatcher ${appVersion}`;
    dot.classList.add("ok");
  } catch (error) {
    status.textContent = "API unavailable";
    version.textContent = "Check server logs";
    dot.classList.add("bad");
  }
})();
