import { Navigate, Route, Routes } from 'react-router-dom';
import { AppLayout } from './layout/AppLayout';
import { AuthStatusPage } from './pages/AuthStatusPage';
import { DashboardPage } from './pages/DashboardPage';
import { DeploymentDetailPage, DeploymentsListPage } from './pages/DeploymentsPage';
import { DriftFindingsPage } from './pages/DriftFindingsPage';
import { IntentDetailPage, IntentsListPage } from './pages/IntentsPage';
import { SystemHealthPage } from './pages/SystemHealthPage';
import { TopologyDevicesPage } from './pages/TopologyDevicesPage';

export function App() {
  return (
    <Routes>
      <Route path="/" element={<AppLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="auth" element={<AuthStatusPage />} />
        <Route path="intents" element={<IntentsListPage />} />
        <Route path="intents/:intentId" element={<IntentDetailPage />} />
        <Route path="topology/devices" element={<TopologyDevicesPage />} />
        <Route path="deployments" element={<DeploymentsListPage />} />
        <Route path="deployments/:deploymentId" element={<DeploymentDetailPage />} />
        <Route path="drift/findings" element={<DriftFindingsPage />} />
        <Route path="system/health" element={<SystemHealthPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
