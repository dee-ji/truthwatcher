import { useEffect, useState } from 'react';

type QueryState<T> = {
  loading: boolean;
  data?: T;
  error?: string;
};

export function useQuery<T>(key: string, fn: () => Promise<T>): QueryState<T> {
  const [state, setState] = useState<QueryState<T>>({ loading: true });

  useEffect(() => {
    let mounted = true;
    setState({ loading: true });
    fn()
      .then((data) => {
        if (mounted) {
          setState({ loading: false, data });
        }
      })
      .catch((error: unknown) => {
        if (mounted) {
          const message = error instanceof Error ? error.message : 'unknown error';
          setState({ loading: false, error: message });
        }
      });

    return () => {
      mounted = false;
    };
    // Query execution is keyed by key string to keep call sites simple.
  }, [key]);

  return state;
}
