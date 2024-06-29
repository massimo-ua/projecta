import { HomeLayout } from '../../Layout.jsx';
import { useProjects } from '../../hooks/projects.js';
import { useEffect } from 'react';
import { ProjectCard } from './ProjectCard.jsx';
import './Projects.css';

export function Projects() {
  const [loading, projects, setPagination] = useProjects();

  useEffect(() => {
    setPagination({ limit: 10, offset: 0 });
  }, []);

  return (
    <HomeLayout>
      {loading
        ? <div>Loading...</div>
        : <div className={'Projects_container'}>{projects.map((project) => <ProjectCard key={project.id} project={project} />)}</div>
      }
    </HomeLayout>
  );
}
