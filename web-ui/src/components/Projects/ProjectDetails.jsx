import React, { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation, useParams, Link } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import HomeLayout from '../../Layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { projectsRepository } from '../../api';
import {
  PieChart,
  Boxes,
  FileText,
  DollarSign,
  Package,
  Settings,
  ArrowLeft,
  Share2,
  Check,
  Users,
  Coins,
} from 'lucide-react';
import { cn } from '@/lib/utils';

export function ProjectDetails() {
  const content = useIntlayer('project-details');
  const navigate = useNavigate();
  const location = useLocation();
  const { locale } = useLocale();
  const { projectId } = useParams();

  const [project, setProject] = useState(null);
  const [loadingProject, setLoadingProject] = useState(true);
  const [copied, setCopied] = useState(false);

  // Fetch project details for header context
  useEffect(() => {
    let isMounted = true;
    setLoadingProject(true);
    projectsRepository
      .getProject(projectId)
      .then((data) => {
        if (isMounted) setProject(data);
      })
      .catch((err) => {
        console.error('Failed to load project details', err);
      })
      .finally(() => {
        if (isMounted) setLoadingProject(false);
      });

    return () => {
      isMounted = false;
    };
  }, [projectId]);

  const navGroups = [
    {
      title: String(content.operations),
      items: [
        { key: 'payments', label: String(content.payments), icon: DollarSign },
        { key: 'total', label: String(content.total), icon: FileText },
        { key: 'assets', label: String(content.assets), icon: Package },
      ],
    },
    {
      title: String(content.taxonomy),
      items: [
        { key: 'categories', label: String(content.categories), icon: PieChart },
        { key: 'types', label: String(content.types), icon: Boxes },
        { key: 'settings', label: String(content.settings), icon: Settings },
      ],
    },
  ];

  // Fix UX issue: determine active tab properly
  const validKeys = ['payments', 'assets', 'categories', 'types', 'settings', 'total'];
  const pathSegments = location.pathname.split('/').filter(Boolean);
  const lastSegment = pathSegments[pathSegments.length - 1];
  const currentTab = validKeys.includes(lastSegment) ? lastSegment : 'payments';

  const handleSelect = (key) => {
    const targetUrl = getLocalizedUrl(`/projects/${projectId}/${key}`, locale);
    navigate(targetUrl);
  };

  const handleShare = () => {
    if (!project?.shareToken) {
      toast.error('Share link is not available');
      return;
    }
    const shareUrl = `${window.location.origin}${getLocalizedUrl(`/projects/share/${project.shareToken}`, locale)}`;
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopied(true);
      toast.success('Share link copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {
      toast.error('Failed to copy share link');
    });
  };

  const projectsListUrl = getLocalizedUrl('/projects', locale);

  return (
    <HomeLayout>
      <div className="space-y-6">
        {/* Project Context Header & Breadcrumbs */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-border/60">
          <div className="space-y-1.5">
            {/* Breadcrumb / Back button */}
            <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
              <Link
                to={projectsListUrl}
                className="inline-flex items-center gap-1.5 hover:text-foreground transition-colors group"
              >
                <ArrowLeft className="h-3.5 w-3.5 transition-transform group-hover:-translate-x-1" />
                <span>{String(content.backToProjects || 'Projects')}</span>
              </Link>
              <span>/</span>
              <span className="text-foreground truncate max-w-[200px]">
                {loadingProject ? '...' : (project?.name || String(content.projectNotFound || 'Project'))}
              </span>
            </div>

            {/* Project Title & Badges */}
            <div className="flex items-center gap-3 flex-wrap">
              {loadingProject ? (
                <Skeleton className="h-8 w-48 rounded-lg" />
              ) : (
                <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
                  {project?.name}
                </h1>
              )}

              {project?.mainCurrency && (
                <Badge variant="outline" className="font-mono text-xs gap-1 border-border/70 bg-card">
                  <Coins className="h-3 w-3 text-muted-foreground" />
                  <span>{project.mainCurrency}</span>
                </Badge>
              )}

              {project?.isShared && (
                <Badge variant="secondary" className="gap-1 text-xs">
                  <Users className="h-3 w-3" />
                  <span>Shared</span>
                </Badge>
              )}
            </div>
          </div>

          {/* Quick Share Action */}
          {project?.shareToken && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
              className="gap-2 self-start sm:self-auto rounded-xl border-border/70 hover:border-primary/40 hover:text-primary transition-all text-xs font-semibold h-9 px-3.5"
            >
              {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Share2 className="h-3.5 w-3.5" />}
              <span>{copied ? 'Copied!' : String(content.shareProject || 'Share')}</span>
            </Button>
          )}
        </div>

        {/* Mobile Navigation Tabs (Single-tap horizontal scrolling pills) */}
        <div className="md:hidden overflow-x-auto pb-1 -mx-4 px-4 sm:mx-0 sm:px-0">
          <div className="flex items-center gap-1.5 p-1 bg-muted/40 rounded-xl border border-border/50 min-w-max">
            {navGroups.flatMap((g) => g.items).map((item) => {
              const Icon = item.icon;
              const isActive = currentTab === item.key;
              return (
                <button
                  key={item.key}
                  type="button"
                  onClick={() => handleSelect(item.key)}
                  className={cn(
                    "flex items-center gap-2 px-3 py-1.5 text-xs font-medium rounded-lg transition-all whitespace-nowrap",
                    isActive
                      ? "bg-primary text-primary-foreground shadow-xs font-semibold"
                      : "text-muted-foreground hover:text-foreground hover:bg-background/80"
                  )}
                >
                  <Icon className="h-3.5 w-3.5" />
                  <span>{item.label}</span>
                </button>
              );
            })}
          </div>
        </div>

        {/* Main Content Layout with Desktop Sidebar */}
        <div className="flex flex-col md:flex-row gap-8 items-start">
          {/* Desktop Sidebar Navigation */}
          <aside className="hidden md:flex flex-col w-52 shrink-0 space-y-6 sticky top-24">
            {navGroups.map((group) => (
              <div key={group.title} className="space-y-1.5">
                <h4 className="text-[11px] font-semibold text-muted-foreground/80 uppercase tracking-wider px-3">
                  {group.title}
                </h4>
                <nav className="space-y-1">
                  {group.items.map((item) => {
                    const Icon = item.icon;
                    const isActive = currentTab === item.key;
                    return (
                      <button
                        key={item.key}
                        type="button"
                        onClick={() => handleSelect(item.key)}
                        className={cn(
                          "w-full flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-xl transition-all text-left",
                          isActive
                            ? "bg-primary text-primary-foreground shadow-sm shadow-primary/20 font-semibold"
                            : "text-muted-foreground hover:bg-accent/80 hover:text-foreground"
                        )}
                      >
                        <Icon className={cn("h-4 w-4 shrink-0", isActive ? "text-primary-foreground" : "text-muted-foreground")} />
                        <span>{item.label}</span>
                      </button>
                    );
                  })}
                </nav>
              </div>
            ))}
          </aside>

          {/* Main Outlet Area */}
          <div className="flex-1 min-w-0 w-full">
            <Outlet />
          </div>
        </div>
      </div>
    </HomeLayout>
  );
}

export default ProjectDetails;
