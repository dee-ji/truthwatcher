import { Link } from 'react-router-dom';
import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { PageHeader } from '../components/PageHeader';
import { useQuery } from '../hooks/useQuery';

const client = new ApiClient();

export function DashboardPage() {
  const intents = useQuery('dashboard-intents', () => client.listIntents());
  const devices = useQuery('dashboard-devices', () => client.listTopologyDevices());
  const drift = useQuery('dashboard-drift', () => client.listDriftFindings());

  return (
    <section>
      <PageHeader title="Dashboard" subtitle="Truthwatcher control plane overview." />
      <div className="grid">
        <MetricCard title="Intents" value={intents.data?.length} loading={intents.loading} error={intents.error} to="/intents" />
        <MetricCard title="Topology devices" value={devices.data?.length} loading={devices.loading} error={devices.error} to="/topology/devices" />
        <MetricCard title="Drift findings" value={drift.data?.length} loading={drift.loading} error={drift.error} to="/drift/findings" />
      </div>
      <AsyncState loading={false} error={undefined} empty={false} emptyLabel="">
        <div className="panel muted">Use this landing page to track intent, topology, deployment, and drift at a glance.</div>
      </AsyncState>
    </section>
  );
}

function MetricCard({ title, value, loading, error, to }: { title: string; value?: number; loading: boolean; error?: string; to: string }) {
  return (
    <Link className="panel metric-card" to={to}>
      <h3>{title}</h3>
      {loading ? <p>Loading…</p> : error ? <p className="error-text">{error}</p> : <p className="metric-value">{value ?? 0}</p>}
    </Link>
  );
}
