import React, { useEffect, useState } from 'react';
import { useIntlayer } from 'react-intlayer';
import HomeLayout from '../../Layout';
import { useProjects } from '../../hooks/projects';
import { ProjectCard } from './ProjectCard';
import { AddProjectModal } from './AddProjectModal';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { FolderPlus, Plus } from 'lucide-react';
import './Projects.css';

export function Projects() {
  const content = useIntlayer('projects');
  const [loading, projects, setPagination] = useProjects();
  const [modalOpen, setModalOpen] = useState(false);

  const refreshProjects = () => {
    setPagination({ limit: 10, offset: 0 });
  };

  useEffect(() => {
    refreshProjects();
  }, []);

  const handleSuccess = () => {
    setModalOpen(false);
    refreshProjects();
  };

  return (
    <HomeLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between border-b pb-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{String(content.title)}</h1>
            <p className="text-sm text-muted-foreground">{String(content.subtitle)}</p>
          </div>
          <Button onClick={() => setModalOpen(true)} className="gap-2 font-semibold shadow-sm">
            <Plus className="h-4 w-4" />
            {String(content.createProjectButton)}
          </Button>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            <Skeleton className="h-44 w-full rounded-xl" />
            <Skeleton className="h-44 w-full rounded-xl" />
            <Skeleton className="h-44 w-full rounded-xl" />
          </div>
        ) : projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center rounded-xl border border-dashed bg-muted/20">
            <FolderPlus className="h-12 w-12 text-muted-foreground/50 mb-3" />
            <h3 className="text-lg font-semibold">{String(content.noProjectsFound)}</h3>
            <p className="text-sm text-muted-foreground mt-1 mb-4">{String(content.noProjectsDesc)}</p>
            <Button onClick={() => setModalOpen(true)} className="gap-2 font-semibold shadow-sm">
              <Plus className="h-4 w-4" />
              {String(content.createProjectButton)}
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((project) => (
              <ProjectCard key={project.id} project={project} />
            ))}
          </div>
        )}

        <AddProjectModal
          open={modalOpen}
          onSuccess={handleSuccess}
          onCancel={() => setModalOpen(false)}
        />
      </div>
    </HomeLayout>
  );
}
