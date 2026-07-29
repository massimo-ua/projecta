import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Link } from 'react-router-dom';
import { FolderKanban, ArrowRight, Share2, Check, Users } from 'lucide-react';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { toast } from 'sonner';

export function ProjectCard({ project }) {
  const content = useIntlayer('projects');
  const { locale } = useLocale();
  const [copied, setCopied] = useState(false);
  const projectUrl = getLocalizedUrl(`/projects/${project.id}`, locale);

  const handleShare = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (!project.shareToken) {
      toast.error(String(content?.shareTokenNotAvailable || 'Share token not available'));
      return;
    }
    const shareUrl = `${window.location.origin}${getLocalizedUrl(`/projects/share/${project.shareToken}`, locale)}`;
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopied(true);
      toast.success(String(content?.shareLinkCopied || 'Share link copied to clipboard!'));
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {
      toast.error(String(content?.failedToCopyShareLink || 'Failed to copy share link'));
    });
  };

  return (
    <Card className="group h-full flex flex-col justify-between transition-all duration-200 hover:shadow-md hover:border-primary/50 relative">
      <CardHeader className="p-5 pb-3">
        <div className="flex items-center justify-between gap-2 mb-2">
          <div className="flex items-center gap-2 min-w-0">
            <div className="p-2 rounded-lg bg-primary/10 text-primary shrink-0">
              <FolderKanban className="h-5 w-5" />
            </div>
            <CardTitle className="text-lg font-semibold line-clamp-1 group-hover:text-primary transition-colors">
              <Link to={projectUrl} title={project.name}>
                {project.name}
              </Link>
            </CardTitle>
          </div>
          {project.shareToken && (
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-muted-foreground hover:text-primary shrink-0"
              onClick={handleShare}
              title={String(content?.copyShareLinkTooltip || 'Copy project share link')}
            >
              {copied ? <Check className="h-4 w-4 text-green-500" /> : <Share2 className="h-4 w-4" />}
            </Button>
          )}
        </div>
        {project.isShared && (
          <div className="mb-2">
            <Badge variant="secondary" className="gap-1 text-xs font-normal bg-secondary/60">
              <Users className="h-3 w-3" />
              {String(content?.sharedTag || 'Shared')} {project.owner?.name ? `${String(content?.sharedBy || 'by')} ${project.owner.name}` : ''}
            </Badge>
          </div>
        )}
        <CardDescription className="line-clamp-2 text-sm text-muted-foreground leading-relaxed">
          {project.description || String(content?.noDescription || 'No description provided.')}
        </CardDescription>
      </CardHeader>
      <CardContent className="p-5 pt-0 flex justify-end">
        <Link
          to={projectUrl}
          className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline"
        >
          {String(content?.viewDetailsLink || 'View details')}
          <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" />
        </Link>
      </CardContent>
    </Card>
  );
}

ProjectCard.propTypes = {
  project: PropTypes.shape({
    id: PropTypes.string.isRequired,
    name: PropTypes.string.isRequired,
    description: PropTypes.string,
    shareToken: PropTypes.string,
    isShared: PropTypes.bool,
    owner: PropTypes.shape({
      id: PropTypes.string,
      name: PropTypes.string,
    }),
  }).isRequired,
};
