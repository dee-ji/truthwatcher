import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { PageHeader } from '../components/PageHeader';
import { StatusBadge } from '../components/StatusBadge';
import { Table } from '../components/Table';
import { useQuery } from '../hooks/useQuery';

const client = new ApiClient();

export function DriftFindingsPage() {
  const findings = useQuery('drift-findings', () => client.listDriftFindings());

  return (
    <section>
      <PageHeader title="Drift findings" subtitle="Detected drift between intent artifact and observed state." />
      <AsyncState loading={findings.loading} error={findings.error} empty={!findings.data || findings.data.length === 0} emptyLabel="No drift findings.">
        <Table
          rows={findings.data ?? []}
          columns={[
            { key: 'id', header: 'ID', cell: (row) => row.id },
            { key: 'device_id', header: 'Device ID', cell: (row) => row.device_id },
            { key: 'severity', header: 'Severity', cell: (row) => <StatusBadge value={row.severity} /> },
            { key: 'kind', header: 'Kind', cell: (row) => row.kind },
            { key: 'summary', header: 'Summary', cell: (row) => row.summary },
          ]}
        />
      </AsyncState>
    </section>
  );
}
