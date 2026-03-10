import { useMemo, useState } from 'react';
import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { DetailList } from '../components/DetailList';
import { PageHeader } from '../components/PageHeader';
import { useQuery } from '../hooks/useQuery';

export function AuthStatusPage() {
  const [token, setToken] = useState('');
  const client = useMemo(() => new ApiClient(token || undefined), [token]);
  const version = useQuery(`version-${token}`, () => client.getVersion());

  return (
    <section>
      <PageHeader title="Login placeholder / auth status" subtitle="Token-only placeholder for current authn mode." />
      <div className="panel">
        <label htmlFor="token">Bearer token (optional for local dev bypass)</label>
        <input id="token" value={token} onChange={(event) => setToken(event.target.value)} placeholder="paste bearer token" />
      </div>
      <AsyncState loading={version.loading} error={version.error} empty={!version.data} emptyLabel="No auth status available.">
        <DetailList
          items={[
            { label: 'API base URL', value: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080' },
            { label: 'Version endpoint access', value: 'reachable' },
            { label: 'Auth mode', value: token ? 'Bearer token provided' : 'No token provided' },
            { label: 'Server version', value: version.data?.version ?? 'unknown' },
            { label: 'RBAC render mode', value: 'TODO(truthwatcher): map claims to UI capabilities' },
          ]}
        />
      </AsyncState>
    </section>
  );
}
