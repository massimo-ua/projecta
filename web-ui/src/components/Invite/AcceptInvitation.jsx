import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { useGoogleLogin } from '@react-oauth/google';
import { toast } from 'sonner';
import { Loader2, Mail, AlertCircle, Sparkles } from 'lucide-react';
import HomeLayout from '../../Layout';
import { Button } from '@/components/ui/button';
import { invitationsRepository, authProvider } from '../../api';

function GoogleIcon(props) {
  return (
    <svg className="h-4 w-4 shrink-0" viewBox="0 0 24 24" {...props}>
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
        fill="#EA4335"
      />
    </svg>
  );
}

export function AcceptInvitation() {
  const { code } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();
  const content = useIntlayer('accept-invitation');

  const [verifying, setVerifying] = useState(true);
  const [invitation, setInvitation] = useState(null);
  const [error, setError] = useState(null);
  const [loginLoading, setLoginLoading] = useState(false);

  useEffect(() => {
    let isMounted = true;

    async function checkInvitation() {
      if (!code) {
        setVerifying(false);
        setError('No invitation code provided');
        return;
      }

      try {
        const data = await invitationsRepository.validate(code);
        if (isMounted) {
          setInvitation(data);
          setVerifying(false);
        }
      } catch (err) {
        if (isMounted) {
          setError(err.message || String(content.invalidSubtitle));
          setVerifying(false);
        }
      }
    }

    checkInvitation();

    return () => {
      isMounted = false;
    };
  }, [code, content.invalidSubtitle]);

  const handleGoogleSuccess = async (response) => {
    setLoginLoading(true);
    try {
      await authProvider.loginSocial(response.code, 'GOOGLE', code);
      toast.success(String(content.loginSuccess));
      navigate(getLocalizedUrl('/', locale));
    } catch (err) {
      toast.error(`Login failed: ${err.message}`);
    } finally {
      setLoginLoading(false);
    }
  };

  const handleGoogleError = (err) => {
    setLoginLoading(false);
    toast.error(`Google login failed: ${err.message || 'Unknown error'}`);
  };

  const triggerGoogleLogin = useGoogleLogin({
    flow: 'auth-code',
    onSuccess: handleGoogleSuccess,
    onError: handleGoogleError,
  });

  return (
    <HomeLayout>
      <div className="flex min-h-[70vh] items-center justify-center p-4">
        <div className="max-w-md w-full text-center rounded-3xl border border-border/80 bg-card p-8 sm:p-10 shadow-lg shadow-black/5 dark:shadow-none space-y-6">
          {verifying ? (
            <div className="space-y-4 py-8">
              <Loader2 className="h-10 w-10 animate-spin mx-auto text-primary" />
              <p className="text-sm text-muted-foreground">{String(content.loading)}</p>
            </div>
          ) : error || !invitation ? (
            <div className="space-y-5">
              <div className="mx-auto p-4 rounded-2xl bg-rose-500/10 text-rose-500 w-fit ring-1 ring-rose-500/20">
                <AlertCircle className="h-8 w-8" />
              </div>
              <div className="space-y-2">
                <h2 className="text-xl font-bold tracking-tight text-foreground">
                  {String(content.invalidTitle)}
                </h2>
                <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed">
                  {error || String(content.invalidSubtitle)}
                </p>
              </div>
              <div className="pt-2">
                <Button
                  onClick={() => navigate(getLocalizedUrl('/login', locale))}
                  variant="outline"
                  className="w-full rounded-xl h-11 font-semibold"
                >
                  {String(content.loginInstead)}
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              <div className="mx-auto p-4 rounded-2xl bg-primary/10 text-primary w-fit ring-1 ring-primary/20">
                <Sparkles className="h-8 w-8" />
              </div>

              <div className="space-y-2">
                <h2 className="text-2xl font-bold tracking-tight text-foreground">
                  {String(content.invitedTitle)}
                </h2>
                <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed">
                  {String(content.invitedSubtitle)}
                </p>
              </div>

              {/* Invited Email badge */}
              <div className="p-3.5 rounded-2xl border border-border/70 bg-muted/30 text-left flex items-center gap-3">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                  <Mail className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <p className="text-[11px] uppercase tracking-wider text-muted-foreground font-semibold">
                    {String(content.invitationFor)}
                  </p>
                  <p className="font-semibold text-sm text-foreground truncate">
                    {invitation.email}
                  </p>
                </div>
              </div>

              <p className="text-xs text-muted-foreground leading-relaxed">
                {String(content.signInInstructions)}
              </p>

              <Button
                type="button"
                variant="outline"
                className="w-full flex items-center justify-center gap-2.5 h-11 rounded-xl font-semibold border-border/80 hover:bg-accent/80 hover:border-border transition-all shadow-2xs"
                onClick={() => {
                  setLoginLoading(true);
                  triggerGoogleLogin();
                }}
                disabled={loginLoading}
              >
                {loginLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                ) : (
                  <GoogleIcon />
                )}
                <span>Sign in with Google</span>
              </Button>
            </div>
          )}
        </div>
      </div>
    </HomeLayout>
  );
}

export default AcceptInvitation;
