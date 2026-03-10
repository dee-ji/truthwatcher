import type { ApiError, Deployment, Device, DriftFinding, Intent } from './types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export class ApiClient {
  constructor(private readonly token?: string) {}

  async getVersion() {
    return this.request<{ version: string }>('/version');
  }

  async getHealthz() {
    return this.requestText('/healthz');
  }

  async getReadyz() {
    return this.requestText('/readyz');
  }

  async listIntents() {
    return this.request<Intent[]>('/api/v1/intents');
  }

  async getIntentByID(intentID: string) {
    return this.request<Intent>(`/api/v1/intents/${intentID}`);
  }

  async listTopologyDevices() {
    return this.request<Device[]>('/api/v1/topology/devices');
  }

  async getDeploymentByID(deploymentID: string) {
    return this.request<Deployment>(`/api/v1/deployments/${deploymentID}`);
  }

  async listDriftFindings() {
    return this.request<DriftFinding[]>('/api/v1/drift/findings');
  }

  private async request<T>(path: string): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      headers: this.headers(),
    });
    if (!response.ok) {
      const error = (await response.json().catch(() => ({}))) as ApiError;
      throw new Error(error.error ?? `request failed: ${response.status}`);
    }
    return (await response.json()) as T;
  }

  private async requestText(path: string): Promise<string> {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      headers: this.headers(),
    });
    if (!response.ok) {
      throw new Error(`request failed: ${response.status}`);
    }
    return response.text();
  }

  private headers() {
    const headers = new Headers({
      'Content-Type': 'application/json',
    });
    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`);
    }
    return headers;
  }
}
