import type { ReactNode } from 'react';

type Column<T> = {
  key: string;
  header: string;
  cell: (row: T) => ReactNode;
};

type Props<T> = {
  columns: Array<Column<T>>;
  rows: T[];
};

export function Table<T>({ columns, rows }: Props<T>) {
  return (
    <div className="panel">
      <table>
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.key}>{column.header}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, index) => (
            <tr key={index}>
              {columns.map((column) => (
                <td key={column.key}>{column.cell(row)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
