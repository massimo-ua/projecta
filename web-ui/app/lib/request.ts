type responseMapper<T> = (res: Response) => Promise<T>;
export const request = {
  get: <T>(mapper: responseMapper<T>) => async <T>(url: string, opts?: RequestInit) => {
    const res = await fetch(url, opts);
    return mapper(res);
  },
  post: async <T>(url: string, body: any, mapper: responseMapper<T>) => {
    const res = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });
    return mapper(res);
  },
};
