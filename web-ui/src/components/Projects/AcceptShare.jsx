import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { projectsRepository } from '../../api';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import HomeLayout from '../../Layout';

export function AcceptShare() {
  const { shareToken } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();
  const content = useIntlayer('accept-share');
  const [status, setStatus] = useState('loading'); // 'loading' | 'success' | 'error'
  const [isOwner, setIsOwner] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    let isMounted = true;

    async function handleAcceptShare() {
      if (!shareToken) {
        setStatus('error');
        setErrorMsg(String(content.noTokenProvided));
        return;
      }

      try {
        const project = await projectsRepository.acceptShare(shareToken);
        if (isMounted) {
          const ownerState = !project.isShared;
          setIsOwner(ownerState);
          setStatus('success');
          if (ownerState) {
            toast.info(`${String(content.projectWord)} "${project.name}": ${String(content.alreadyOwnerToast)}`);
          } else {
            toast.success(`${String(content.projectWord)} "${project.name}" ${String(content.addedToList)}`);
          }
          setTimeout(() => {
            navigate(getLocalizedUrl(`/projects/${project.id}`, locale), { replace: true });
          }, 1200);
        }
      } catch (err) {
        if (isMounted) {
          setStatus('error');
          setErrorMsg(err.message || String(content.failedToAcceptShare));
        }
      }
    }

    handleAcceptShare();

    return () => {
      isMounted = false;
    };
  }, [shareToken, navigate, locale, content]);

  return (
    <HomeLayout>
      <div className="min-h-[60vh] flex items-center justify-center p-4">
        <div className="max-w-md w-full text-center rounded-3xl border border-border/80 bg-card p-8 sm:p-10 shadow-lg shadow-black/5 dark:shadow-none space-y-5">
          <div className="mx-auto p-4 rounded-2xl bg-primary/10 text-primary w-fit ring-1 ring-primary/20">
            {status === 'loading' && <Loader2 className="h-8 w-8 animate-spin" />}
            {status === 'success' && <CheckCircle2 className="h-8 w-8 text-emerald-500" />}
            {status === 'error' && <AlertCircle className="h-8 w-8 text-rose-500" />}
          </div>

          <div className="space-y-1.5">
            <h2 className="text-xl font-bold tracking-tight text-foreground">
              {status === 'loading' && String(content.joiningProject)}
              {status === 'success' && String(content.projectShared)}
              {status === 'error' && String(content.unableToJoinProject)}
            </h2>
            <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed">
              {status === 'loading' && String(content.processingInvitation)}
              {status === 'success' && (isOwner ? String(content.alreadyOwner) : String(content.accessGrantedRedirecting))}
              {status === 'error' && (errorMsg || String(content.invalidOrExpiredLink))}
            </p>
          </div>

          {status === 'error' && (
            <div className="pt-2">
              <Button
                onClick={() => navigate(getLocalizedUrl('/projects', locale))}
                className="w-full rounded-xl h-11 font-semibold"
              >
                {String(content.backToProjectsButton)}
              </Button>
            </div>
          )}
        </div>
      </div>
    </HomeLayout>
  );
}

export default AcceptShare;
