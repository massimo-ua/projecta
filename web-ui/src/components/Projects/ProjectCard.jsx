import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Link } from 'react-router-dom';
import { FolderKanban, ArrowRight } from 'lucide-react';

export function ProjectCard({ project }) {
  return (
    <Card className="group h-full flex flex-col justify-between transition-all duration-200 hover:shadow-md hover:border-primary/50">
      <CardHeader className="p-5 pb-3">
        <div className="flex items-center gap-2 mb-2">
          <div className="p-2 rounded-lg bg-primary/10 text-primary">
            <FolderKanban className="h-5 w-5" />
          </div>
          <CardTitle className="text-lg font-semibold line-clamp-1 group-hover:text-primary transition-colors">
            <Link to={`/projects/${project.id}`} title={project.name}>
              {project.name}
            </Link>
          </CardTitle>
        </div>
        <CardDescription className="line-clamp-2 text-sm text-muted-foreground leading-relaxed">
          {project.description || 'No description provided.'}
        </CardDescription>
      </CardHeader>
      <CardContent className="p-5 pt-0 flex justify-end">
        <Link
          to={`/projects/${project.id}`}
          className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline"
        >
          View details
          <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" />
        </Link>
      </CardContent>
    </Card>
  );
}
