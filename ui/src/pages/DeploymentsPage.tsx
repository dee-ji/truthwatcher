import { useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ApiClient } from '../api/client';
import { AsyncState } from '../components/AsyncState';
import { DetailList } from '../components/DetailList';
import { PageHeader } from '../components/PageHeader';
import { PlaceholderState } from '../components/PlaceholderState';
import { StatusBadge } from '../components/StatusBadge';
import { useQuery } from '../hooks/useQuery';

const client = new ApiClient();

export function DeploymentsListPage() {
  return (
    <section>
      <PageHeader title="Deployments" subtitle="Deployment list flow." />
      <PlaceholderState
        title="Endpoint placeholder"
        message="TODO(truthwatcher): Spanreed currently supports deployment create and deployment get by ID, but no deployment list endpoint yet."
      />
      <p>
        If you have a deployment ID, open its detail page by editing the URL to <code>/deployments/&lt;deployment_id&gt;</code>.
      </p>
    </section>
  );
}

export function DeploymentDetailPage() {
  const { deploymentId = '' } = useParams();
  const queryKey = useMemo(() => `deployment-${deploymentId}`, [deploymentId]);
  const deployment = useQuery(queryKey, () => client.getDeploymentByID(deploymentId));

  return (
    <section>
      <PageHeader
        title={`Deployment detail: ${deploymentId}`}
        subtitle="Execution metadata for intent deployment."
        actions={
          <Link className="button-link" to="/deployments">
            Back to list
          </Link>
        }
      />
      <AsyncState loading={deployment.loading} error={deployment.error} empty={!deployment.data} emptyLabel="Deployment not found.">
        <div className="panel">
          <h3>Deployment status</h3>
          <StatusBadge value={deployment.data?.status ?? 'unknown'} />
        </div>
        <DetailList
          items={[
            { label: 'Deployment ID', value: deployment.data?.id ?? '' },
            { label: 'Intent ID', value: deployment.data?.intent_id ?? '' },
            { label: 'Mode', value: deployment.data?.mode ?? '' },
            { label: 'Idempotency key', value: deployment.data?.idempotency_key ?? '' },
            { label: 'Created at', value: deployment.data ? new Date(deployment.data.created_at).toLocaleString() : '' },
          ]}
        />
      </AsyncState>
    </section>
  );
}
