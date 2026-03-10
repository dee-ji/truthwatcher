import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { PageHeader } from '../components/PageHeader';
import { Table } from '../components/Table';
import { useQuery } from '../hooks/useQuery';

const client = new ApiClient();

export function TopologyDevicesPage() {
  const devices = useQuery('topology-devices', () => client.listTopologyDevices());

  return (
    <section>
      <PageHeader title="Topology devices" subtitle="Current topology device inventory." />
      <AsyncState loading={devices.loading} error={devices.error} empty={!devices.data || devices.data.length === 0} emptyLabel="No topology devices found.">
        <Table
          rows={devices.data ?? []}
          columns={[
            { key: 'id', header: 'ID', cell: (row) => row.id },
            { key: 'hostname', header: 'Hostname', cell: (row) => row.hostname },
            { key: 'vendor', header: 'Vendor', cell: (row) => row.vendor },
            { key: 'platform', header: 'Platform', cell: (row) => row.platform ?? '-' },
            { key: 'site', header: 'Site', cell: (row) => row.site ?? '-' },
          ]}
        />
      </AsyncState>
    </section>
  );
}
