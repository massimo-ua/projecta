import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { projectsRepository } from '../../api';
import { toast } from 'sonner';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, Share2, CheckCircle2, AlertCircle } from 'lucide-react';
import { useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';

export function AcceptShare() {
  const { shareToken } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();
  const [status, setStatus] = useState('loading'); // 'loading' | 'success' | 'error'
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    let isMounted = true;

    async function handleAcceptShare() {
      if (!shareToken) {
        setStatus('error');
        setErrorMsg('No share token provided.');
        return;
      }

      try {
        const project = await projectsRepository.acceptShare(shareToken);
        if (isMounted) {
          setStatus('success');
          toast.success(`Project "${project.name}" added to your list!`);
          setTimeout(() => {
            navigate(getLocalizedUrl(`/projects/${project.id}`, locale), { replace: true });
          }, 1200);
        }
      } catch (err) {
        if (isMounted) {
          setStatus('error');
          setErrorMsg(err.message || 'Failed to accept project share.');
        }
      }
    }

    handleAcceptShare();

    return () => {
      isMounted = false;
    };
  }, [shareToken, navigate, locale]);

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
            {status === 'loading' && 'Joining Project...'}
            {status === 'success' && 'Project Shared!'}
            {status === 'error' && 'Unable to Join Project'}
          </CardTitle>
          <CardDescription>
            {status === 'loading' && 'Processing your invitation to collaborate on this project.'}
            {status === 'success' && 'You now have access to this project. Redirecting...'}
            {status === 'error' && (errorMsg || 'The share link may be invalid or expired.')}
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-2">
          {status === 'error' && (
            <Button
              onClick={() => navigate(getLocalizedUrl('/projects', locale))}
              className="w-full"
            >
              Back to Projects
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AcceptShare;
