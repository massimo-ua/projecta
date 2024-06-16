'use client';
import { useProjectCollection } from '@/app/lib/projects';
import ProjectCard from '@/app/ui/project/ProjectCard';
import { useState } from 'react';

export default function Page() {
  const [pagination, setPagination] = useState();
  const [projects, loading] = useProjectCollection(pagination);
  return (
    <main>
      <h1 className={'mb-4 text-xl md:text-2xl'}>
        Dashboard
      </h1>
      <div className="w-full">
        {loading ? (
          <p>Loading...</p>
        ) : (
          projects.map((project) => (
            <ProjectCard key={project.id} {...project} />
          ))
        )}
      </div>
    </main>
  );
}
