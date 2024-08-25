import { useCallback, useEffect, useRef, useState } from "react";

const cacheStore = new Map<string, any>();

type UseFetchOptions = {
  enabled: boolean;
};

export const useFetch = <T = any>(
  fetchKey: any,
  fetchFn: () => Promise<T>,
  options?: Partial<UseFetchOptions>
) => {
  const key = JSON.stringify(fetchKey);
  const loadingRef = useRef(false);
  const [isLoading, setIsLoading] = useState(false);
  const [data, setData] = useState<T | undefined>(cacheStore.get(key));
  const [error, setError] = useState<Error | undefined>();

  const fetchData = useCallback(async () => {
    if (loadingRef.current) {
      return;
    }

    try {
      loadingRef.current = true;
      setIsLoading(true);

      const res = await fetchFn();
      setData(res);
      cacheStore.set(key, res);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Unknown error"));
      setData(undefined);
    } finally {
      loadingRef.current = false;
      setIsLoading(false);
    }
  }, [key]);

  useEffect(() => {
    if (options?.enabled !== false) {
      fetchData();
    }
  }, [fetchData, options?.enabled]);

  return { data, isLoading, error, refetch: fetchData };
};
