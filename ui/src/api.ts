//

export const API_BASE_URL = "/api";

export const fetchAPI = async <T = any>(
  url: string,
  init: RequestInit = {}
) => {
  const res = await fetch(API_BASE_URL + url, init);
  if (!res.ok) {
    const data = await res.json().catch(() => null);
    const message = data.message ?? res.statusText;

    throw new Error(message);
  }

  return (await res.json()) as T;
};
