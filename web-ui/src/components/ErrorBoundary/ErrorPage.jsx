import React from 'react';
import PropTypes from 'prop-types';
import { useNavigate, useRouteError } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  AlertTriangle, RefreshCw, Home, RotateCcw,
} from 'lucide-react';
import { Button } from '@/components/ui/button';

export function ErrorPage({ error: errorProp, resetErrorBoundary }) {
  let routeError = null;
  try {
    routeError = useRouteError();
  } catch (e) {
    // Component used outside Router context
  }

  const error = errorProp || routeError;
  const content = useIntlayer('error-boundary');
  const { locale } = useLocale();

  let navigate = null;
  try {
    navigate = useNavigate();
  } catch (e) {
    // Component used outside Router context
  }

  const errorMessage = error?.message || (typeof error === 'string' ? error : null);

  const handleReload = () => {
    window.location.reload();
  };

  const handleGoHome = () => {
    const homeUrl = getLocalizedUrl('/', locale);
    if (resetErrorBoundary) {
      resetErrorBoundary();
    }
    if (navigate) {
      navigate(homeUrl);
    } else {
      window.location.href = homeUrl;
    }
  };

  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center p-6 text-center">
      <div className="w-full max-w-md space-y-6">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10 text-destructive">
          <AlertTriangle className="h-8 w-8" />
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {String(content?.title || 'Something went wrong')}
          </h1>
          <p className="text-sm text-muted-foreground">
            {String(content?.description || 'An unexpected error has occurred.')}
          </p>
        </div>

        {errorMessage && (
          <div className="rounded-lg border bg-muted/50 p-3 text-left font-mono text-xs text-muted-foreground overflow-auto max-h-32">
            <p className="font-semibold text-foreground/80 mb-1">
              {String(content?.detailsTitle || 'Error Details')}
              :
            </p>
            <p className="break-words">{errorMessage}</p>
          </div>
        )}

        <div className="flex flex-wrap items-center justify-center gap-3 pt-2">
          {resetErrorBoundary && (
            <Button onClick={resetErrorBoundary} className="gap-2 font-semibold shadow-sm">
              <RotateCcw className="h-4 w-4" />
              {String(content?.tryAgain || 'Try again')}
            </Button>
          )}

          <Button variant="outline" onClick={handleReload} className="gap-2 font-semibold">
            <RefreshCw className="h-4 w-4" />
            {String(content?.reloadPage || 'Reload page')}
          </Button>

          <Button variant="ghost" onClick={handleGoHome} className="gap-2 font-semibold">
            <Home className="h-4 w-4" />
            {String(content?.goHome || 'Go to Home')}
          </Button>
        </div>
      </div>
    </div>
  );
}

ErrorPage.propTypes = {
  error: PropTypes.oneOfType([
    PropTypes.instanceOf(Error),
    PropTypes.shape({
      message: PropTypes.string,
    }),
    PropTypes.string,
  ]),
  resetErrorBoundary: PropTypes.func,
};

ErrorPage.defaultProps = {
  error: null,
  resetErrorBoundary: null,
};
