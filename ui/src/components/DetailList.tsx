type DetailItem = {
  label: string;
  value: string;
};

type Props = {
  items: DetailItem[];
};

export function DetailList({ items }: Props) {
  return (
    <div className="panel">
      <dl className="detail-list">
        {items.map((item) => (
          <div key={item.label}>
            <dt>{item.label}</dt>
            <dd>{item.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
