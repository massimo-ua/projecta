import React, { useEffect, useState } from 'react';
import { useIntlayer } from 'react-intlayer';
import HomeLayout from '../../Layout';
import { useProjects } from '../../hooks/projects';
import { ProjectCard } from './ProjectCard';
import { AddProjectModal } from './AddProjectModal';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { FolderPlus, Plus, Sparkles } from 'lucide-react';
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
      <div className="space-y-8">
        {/* Header with Title, Count Badge, and Primary Action */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-border/60">
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
                {String(content.title)}
              </h1>
              {!loading && projects.length > 0 && (
                <Badge variant="secondary" className="font-mono text-xs px-2 py-0.5 rounded-full font-semibold">
                  {projects.length}
                </Badge>
              )}
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              {String(content.subtitle)}
            </p>
          </div>

          <Button
            onClick={() => setModalOpen(true)}
            className="gap-2 font-semibold shadow-sm shadow-primary/20 hover:shadow-md transition-all self-start sm:self-auto rounded-xl px-5 h-10"
          >
            <Plus className="h-4 w-4" />
            <span>{String(content.createProjectButton)}</span>
          </Button>
        </div>

        {/* Content Section */}
        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="rounded-2xl border border-border/60 p-6 space-y-4 bg-card">
                <div className="flex items-center justify-between">
                  <Skeleton className="h-10 w-10 rounded-xl" />
                  <Skeleton className="h-8 w-8 rounded-lg" />
                </div>
                <Skeleton className="h-5 w-3/4 rounded-md" />
                <Skeleton className="h-4 w-full rounded-md" />
                <Skeleton className="h-4 w-2/3 rounded-md" />
                <div className="pt-2 border-t border-border/40 flex justify-end">
                  <Skeleton className="h-4 w-24 rounded-md" />
                </div>
              </div>
            ))}
          </div>
        ) : projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 sm:p-16 text-center rounded-2xl border border-dashed border-border/80 bg-muted/20">
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 text-primary mb-4 ring-1 ring-primary/20">
              <FolderPlus className="h-8 w-8" />
            </div>
            <h3 className="text-lg font-bold text-foreground">
              {String(content.noProjectsFound)}
            </h3>
            <p className="text-sm text-muted-foreground mt-1.5 mb-6 max-w-sm">
              {String(content.noProjectsDesc)}
            </p>
            <Button
              onClick={() => setModalOpen(true)}
              className="gap-2 font-semibold shadow-sm rounded-xl px-5 h-10"
            >
              <Plus className="h-4 w-4" />
              <span>{String(content.createProjectButton)}</span>
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

export default Projects;
