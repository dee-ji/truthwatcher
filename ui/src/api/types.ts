export type ApiError = {
  error?: string;
};

export type Intent = {
  id: string;
  name: string;
  description: string;
  revision: number;
  created_at: string;
  spec?: Record<string, unknown>;
};

export type Deployment = {
  id: string;
  intent_id: string;
  status: string;
  idempotency_key: string;
  mode: string;
  artifact_refs?: string[];
  targets?: string[];
  created_at: string;
  rollout?: { waves?: Array<{ name: string; order: number; planned_target_count: number }> };
};

export type Device = {
  id: string;
  hostname: string;
  vendor: string;
  platform?: string;
  site?: string;
};

export type DriftFinding = {
  id: string;
  device_id: string;
  severity: string;
  kind: string;
  summary: string;
  created_at: string;
};

export type HealthStatus = {
  name: string;
  ok: boolean;
  detail: string;
};
