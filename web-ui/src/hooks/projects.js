import { useEffect, useState } from 'react';
import { projectsRepository } from '../api';

export function useProjects() {
  const [loading, setLoading] = useState(false);
  const [projects, setProjects] = useState([]);
  const [pagination, setPagination] = useState();

  useEffect(() => {
    if (!pagination) {
      return;
    }

    setLoading(true);
    projectsRepository.getProjects()
      .then((projects) => {
        setProjects(projects);
      })
      .finally(() => setLoading(false));
  }, [pagination]);

  return [loading, projects, setPagination];
}
