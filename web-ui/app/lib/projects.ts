import { Collection, Pagination } from '@/app/lib/types';
import { Project } from '@/app/ui/project/types';
import { request } from '@/app/lib/request';
import { useEffect, useState } from 'react';

const BASE_URL = '/api';

const responseMapper = async (response: Response): Promise<Project[]> => {
  if (!response.ok) {
    throw new Error('Failed to fetch projects');
  }

  const json = await response.json();
  const { projects } = json;
  return projects.map((project: Record<string, unknown>): Project => ({
    id: String(project.id),
    name: String(project.name),
    description: String(project.description),
  }));
}

const requestProjects = request.get<Collection<Project>>(responseMapper);

export function useProjectCollection(opts?: Pagination): [projects: Project[], loading: boolean] {
  const [data, setData] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async (offset: number, limit: number) => {
      try {
        setLoading(true);
        const res = await requestProjects(`${BASE_URL}/projects?offset=${offset}&limit=${limit}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('token')}`,
          },
        });
        setData(res);
      }
      finally {
        setLoading(false);
      }
    }

    const offset = opts?.offset || 0;
    const limit = opts?.limit || 10;
    fetchData(offset, limit).catch(console.error.bind('useProjectCollection'));
  }, [opts]);

  return [data, loading]
}
