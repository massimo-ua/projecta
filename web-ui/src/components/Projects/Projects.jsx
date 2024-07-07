import { useEffect } from 'react';
import HomeLayout from '../../Layout';
import { useProjects } from '../../hooks/projects';
import { ProjectCard } from './ProjectCard';
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
        : <div className="Projects_container">{projects.map((project) => <ProjectCard key={project.id} project={project} />)}</div>}
    </HomeLayout>
  );
}
