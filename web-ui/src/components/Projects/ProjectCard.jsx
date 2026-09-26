import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { FolderKanban, ArrowRight, Share2, Check, Users, Coins } from 'lucide-react';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { toast } from 'sonner';

export function ProjectCard({ project }) {
  const content = useIntlayer('projects');
  const { locale } = useLocale();
  const navigate = useNavigate();
  const [copied, setCopied] = useState(false);
  const projectUrl = getLocalizedUrl(`/projects/${project.id}`, locale);

  const handleCardClick = () => {
    navigate(projectUrl);
  };

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
    <div
      role="button"
      tabIndex={0}
      onClick={handleCardClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleCardClick();
        }
      }}
      className="group relative flex flex-col justify-between rounded-2xl border border-border/70 bg-card p-6 shadow-xs transition-all duration-300 hover:shadow-md hover:shadow-primary/5 hover:border-primary/50 hover:-translate-y-0.5 cursor-pointer text-left focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
    >
      <div>
        {/* Top bar with icon, title, and share button */}
        <div className="flex items-start justify-between gap-3 mb-3">
          <div className="flex items-center gap-3 min-w-0">
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/15 transition-all duration-300 group-hover:bg-primary group-hover:text-primary-foreground group-hover:scale-105">
              <FolderKanban className="h-5 w-5" />
            </div>
            <div className="min-w-0">
              <h3 className="text-base font-semibold tracking-tight text-foreground truncate transition-colors group-hover:text-primary">
                {project.name}
              </h3>
              {project.mainCurrency && (
                <div className="flex items-center gap-1 text-[11px] font-mono text-muted-foreground mt-0.5">
                  <Coins className="h-3 w-3" />
                  <span>{project.mainCurrency}</span>
                </div>
              )}
            </div>
          </div>

          {project.shareToken && (
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8 rounded-lg shrink-0 border-border/60 text-muted-foreground hover:text-primary hover:border-primary/40 hover:bg-primary/5 transition-all"
              onClick={handleShare}
              title={String(content?.copyShareLinkTooltip || 'Copy project share link')}
            >
              {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Share2 className="h-3.5 w-3.5" />}
            </Button>
          )}
        </div>

        {/* Badges */}
        {project.isShared && (
          <div className="mb-3">
            <Badge variant="secondary" className="gap-1.5 text-xs font-normal bg-secondary/80 text-secondary-foreground border border-border/40">
              <Users className="h-3 w-3" />
              <span>{String(content?.sharedTag || 'Shared')}</span>
              {project.owner?.name && (
                <span className="text-muted-foreground">({project.owner.name})</span>
              )}
            </Badge>
          </div>
        )}

        {/* Description */}
        <p className="text-xs text-muted-foreground leading-relaxed line-clamp-2 mb-4 min-h-[2rem]">
          {project.description || String(content?.noDescription || 'No description provided.')}
        </p>
      </div>

      {/* Card Footer */}
      <div className="flex items-center justify-end pt-3 border-t border-border/50 text-xs font-semibold text-primary">
        <span className="inline-flex items-center gap-1.5 transition-transform group-hover:translate-x-0.5">
          {String(content?.viewDetailsLink || 'View details')}
          <ArrowRight className="h-3.5 w-3.5 transition-transform duration-200 group-hover:translate-x-1" />
        </span>
      </div>
    </div>
  );
}

ProjectCard.propTypes = {
  project: PropTypes.shape({
    id: PropTypes.string.isRequired,
    name: PropTypes.string.isRequired,
    description: PropTypes.string,
    mainCurrency: PropTypes.string,
    shareToken: PropTypes.string,
    isShared: PropTypes.bool,
    owner: PropTypes.shape({
      id: PropTypes.string,
      name: PropTypes.string,
    }),
  }).isRequired,
};

export default ProjectCard;
