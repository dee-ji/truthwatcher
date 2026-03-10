import { NavLink, Outlet } from 'react-router-dom';
import { ToastHost } from './ToastHost';

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/auth', label: 'Auth' },
  { to: '/intents', label: 'Intents' },
  { to: '/topology/devices', label: 'Topology Devices' },
  { to: '/deployments', label: 'Deployments' },
  { to: '/drift/findings', label: 'Drift Findings' },
  { to: '/system/health', label: 'System Health' },
];

export function AppLayout() {
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <h2>Truthwatcher</h2>
        <p className="muted-text">UI foundation</p>
        <nav>
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} className={({ isActive }) => (isActive ? 'active' : undefined)} end={item.to === '/'}>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Outlet />
      </main>
      <ToastHost />
    </div>
  );
}
