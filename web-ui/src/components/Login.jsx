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
      <div className="flex min-h-[75vh] items-center justify-center py-10 px-4 sm:px-6">
        <div className="w-full max-w-md">
          {/* Card Container with subtle elevation */}
          <div className="rounded-3xl border border-border/80 bg-card p-8 sm:p-10 shadow-lg shadow-black/5 dark:shadow-none space-y-6">
            <div className="space-y-2 text-center">
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
                {String(content.title)}
              </h1>
              <p className="text-sm text-muted-foreground leading-relaxed">
                {String(content.subtitle)}
              </p>
            </div>

            <div className="space-y-4">
              <GoogleLoginBtn />
            </div>

            <div className="relative flex items-center justify-center">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border/60" />
              </div>
              <div className="relative flex justify-center text-xs uppercase tracking-wider">
                <span className="bg-card px-3 text-muted-foreground font-semibold">
                  {String(content.orContinueWith)}
                </span>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="username" className="text-xs font-semibold text-foreground">
                  {String(content.usernameLabel)}
                </Label>
                <div className="relative">
                  <User className="absolute left-3.5 top-3.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="username"
                    type="text"
                    placeholder={String(content.usernamePlaceholder)}
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="pl-10 h-11 rounded-xl"
                    disabled={loading}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-xs font-semibold text-foreground">
                  {String(content.passwordLabel)}
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3.5 top-3.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="pl-10 h-11 rounded-xl"
                    disabled={loading}
                    required
                  />
                </div>
              </div>

              <Button
                type="submit"
                className="w-full h-11 rounded-xl font-semibold shadow-sm shadow-primary/20 hover:shadow-md transition-all mt-2"
                disabled={loading}
              >
                {loading ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  String(content.signInButton)
                )}
              </Button>
            </form>
          </div>
        </div>
      </div>
    </HomeLayout>
  );
}

export default Login;
