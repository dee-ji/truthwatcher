import { useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { DetailList } from '../components/DetailList';
import { PageHeader } from '../components/PageHeader';
import { Table } from '../components/Table';
import { useQuery } from '../hooks/useQuery';

const client = new ApiClient();

export function IntentsListPage() {
  const intents = useQuery('intents-list', () => client.listIntents());

  return (
    <section>
      <PageHeader title="Intents" subtitle="Desired network behavior and policy definitions." />
      <AsyncState loading={intents.loading} error={intents.error} empty={!intents.data || intents.data.length === 0} emptyLabel="No intent found yet.">
        <Table
          rows={intents.data ?? []}
          columns={[
            { key: 'id', header: 'ID', cell: (row) => <Link to={`/intents/${row.id}`}>{row.id}</Link> },
            { key: 'name', header: 'Name', cell: (row) => row.name },
            { key: 'revision', header: 'Revision', cell: (row) => row.revision },
            { key: 'created', header: 'Created at', cell: (row) => new Date(row.created_at).toLocaleString() },
          ]}
        />
      </AsyncState>
    </section>
  );
}

export function IntentDetailPage() {
  const { intentId = '' } = useParams();
  const queryKey = useMemo(() => `intent-${intentId}`, [intentId]);
  const intent = useQuery(queryKey, () => client.getIntentByID(intentId));

  return (
    <section>
      <PageHeader title={`Intent detail: ${intentId}`} subtitle="Intent revision, metadata, and spec preview." />
      <AsyncState loading={intent.loading} error={intent.error} empty={!intent.data} emptyLabel="Intent not available.">
        <DetailList
          items={[
            { label: 'ID', value: intent.data?.id ?? '' },
            { label: 'Name', value: intent.data?.name ?? '' },
            { label: 'Revision', value: String(intent.data?.revision ?? '') },
            { label: 'Created at', value: intent.data ? new Date(intent.data.created_at).toLocaleString() : '' },
          ]}
        />
        <div className="panel">
          <h3>Spec</h3>
          <pre>{JSON.stringify(intent.data?.spec ?? {}, null, 2)}</pre>
        </div>
      </AsyncState>
    </section>
  );
}
