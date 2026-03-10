type Props = {
  value: string;
};

export function StatusBadge({ value }: Props) {
  const normalized = value.toLowerCase();
  const tone = normalized.includes('error') || normalized.includes('fail')
    ? 'danger'
    : normalized.includes('planned') || normalized.includes('queued')
      ? 'warning'
      : 'ok';

  return <span className={`status-badge ${tone}`}>{value}</span>;
}
