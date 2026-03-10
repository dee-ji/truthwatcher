import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { PageHeader } from '../components/PageHeader';
import { StatusBadge } from '../components/StatusBadge';
import { Table } from '../components/Table';
import { useQuery } from '../hooks/useQuery';

const client = new ApiClient();

export function SystemHealthPage() {
  const health = useQuery('healthz', () => client.getHealthz());
  const ready = useQuery('readyz', () => client.getReadyz());
  const version = useQuery('version', () => client.getVersion());

  const rows = [
    { name: 'healthz', ok: health.data === 'ok', detail: health.data ?? health.error ?? 'pending' },
    { name: 'readyz', ok: ready.data === 'ready', detail: ready.data ?? ready.error ?? 'pending' },
    { name: 'version', ok: Boolean(version.data?.version), detail: version.data?.version ?? version.error ?? 'pending' },
  ];

  return (
    <section>
      <PageHeader title="System health" subtitle="Spanreed process and readiness probes." />
      <AsyncState
        loading={health.loading || ready.loading || version.loading}
        error={health.error ?? ready.error ?? version.error}
        empty={rows.length === 0}
        emptyLabel="No system health data."
      >
        <Table
          rows={rows}
          columns={[
            { key: 'name', header: 'Check', cell: (row) => row.name },
            { key: 'status', header: 'Status', cell: (row) => <StatusBadge value={row.ok ? 'ok' : 'error'} /> },
            { key: 'detail', header: 'Detail', cell: (row) => row.detail },
          ]}
        />
      </AsyncState>
    </section>
  );
}
