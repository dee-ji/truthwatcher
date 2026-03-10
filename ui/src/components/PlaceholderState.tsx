type Props = {
  title: string;
  message: string;
};

export function PlaceholderState({ title, message }: Props) {
  return (
    <div className="panel placeholder">
      <strong>{title}</strong>
      <p>{message}</p>
    </div>
  );
}
