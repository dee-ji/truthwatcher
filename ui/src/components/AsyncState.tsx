import type { ReactNode } from 'react';

type Props = {
  loading: boolean;
  error?: string;
  empty: boolean;
  emptyLabel: string;
  children: ReactNode;
};

export function AsyncState({ loading, error, empty, emptyLabel, children }: Props) {
  if (loading) {
    return <div className="panel muted">Loading…</div>;
  }
  if (error) {
    return <div className="panel error">Error: {error}</div>;
  }
  if (empty) {
    return <div className="panel muted">{emptyLabel}</div>;
  }
  return <>{children}</>;
}
