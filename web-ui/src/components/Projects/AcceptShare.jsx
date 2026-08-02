import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { projectsRepository } from '../../api';
import { toast } from 'sonner';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';

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
    <div className="min-h-[60vh] flex items-center justify-center p-4">
      <Card className="max-w-md w-full text-center shadow-lg border">
        <CardHeader className="pb-4">
          <div className="mx-auto mb-3 p-3 rounded-full bg-primary/10 text-primary w-fit">
            {status === 'loading' && <Loader2 className="h-8 w-8 animate-spin" />}
            {status === 'success' && <CheckCircle2 className="h-8 w-8 text-green-500" />}
            {status === 'error' && <AlertCircle className="h-8 w-8 text-red-500" />}
          </div>
          <CardTitle className="text-xl font-bold">
            {status === 'loading' && String(content.joiningProject)}
            {status === 'success' && String(content.projectShared)}
            {status === 'error' && String(content.unableToJoinProject)}
          </CardTitle>
          <CardDescription>
            {status === 'loading' && String(content.processingInvitation)}
            {status === 'success' && (isOwner ? String(content.alreadyOwner) : String(content.accessGrantedRedirecting))}
            {status === 'error' && (errorMsg || String(content.invalidOrExpiredLink))}
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-2">
          {status === 'error' && (
            <Button
              onClick={() => navigate(getLocalizedUrl('/projects', locale))}
              className="w-full"
            >
              {String(content.backToProjectsButton)}
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AcceptShare;
