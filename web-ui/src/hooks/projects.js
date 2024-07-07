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
      .then((data) => {
        setProjects(data);
      })
      .finally(() => setLoading(false));
  }, [pagination]);

  return [loading, projects, setPagination];
}

export function useProjectTotals(projectId) {
  const [loading, setLoading] = useState(false);
  const [totals, setTotals] = useState([]);
  const [updatedAt, setUpdatedAt] = useState(null);

  const reload = () => setUpdatedAt(Date.now());

  useEffect(() => {
    if (!updatedAt) {
      return;
    }

    setLoading(true);
    projectsRepository.getTotals(projectId)
      .then((data) => {
        setTotals(data);
      })
      .finally(() => setLoading(false));
  }, [updatedAt, projectId]);

  return [loading, totals, reload];
}
