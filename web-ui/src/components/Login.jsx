import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import HomeLayout from '../Layout';
import { authProvider } from '../api';
import { GoogleLoginBtn } from './GoogleLoginBtn';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { toast } from 'sonner';
import { Loader2, Lock, User } from 'lucide-react';
import './Login.css';

export function Login() {
  const content = useIntlayer('login');
  const { locale } = useLocale();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!username || !password) {
      toast.error(String(content.inputRequiredError));
      return;
    }

    setLoading(true);
    try {
      await authProvider.login(username, password);
      toast.success(String(content.loginSuccess));
      const homeUrl = getLocalizedUrl('/', locale);
      navigate(homeUrl);
    } catch (error) {
      toast.error(`${String(content.loginFailed)}: ${error.message || 'Invalid credentials'}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <HomeLayout>
      <div className="flex min-h-[80vh] items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md shadow-lg border-muted">
          <CardHeader className="space-y-1 text-center">
            <CardTitle className="text-2xl font-bold tracking-tight">{String(content.title)}</CardTitle>
            <CardDescription className="text-sm text-muted-foreground">
              {String(content.subtitle)}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-4">
              <GoogleLoginBtn />
            </div>

            <div className="relative flex items-center justify-center">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-muted" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground font-medium">{String(content.orContinueWith)}</span>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="username">{String(content.usernameLabel)}</Label>
                <div className="relative">
                  <User className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="username"
                    type="text"
                    placeholder={String(content.usernamePlaceholder)}
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="pl-9"
                    disabled={loading}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">{String(content.passwordLabel)}</Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="pl-9"
                    disabled={loading}
                    required
                  />
                </div>
              </div>

              <Button type="submit" className="w-full h-10 font-semibold" disabled={loading}>
                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : String(content.signInButton)}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </HomeLayout>
  );
}
