import React, { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

export function CopyableText({ text, label, truncate }) {
  const [copied, setCopied] = useState(false);
  const displayText = truncate && text ? `${text.slice(0, 6)}...${text.slice(-6)}` : text;

  const handleCopy = () => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success('Copied to clipboard');
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="inline-flex items-center gap-1.5 font-mono text-xs">
      {label && <span className="text-muted-foreground">{label}:</span>}
      <span className="bg-muted/50 px-2 py-0.5 rounded border border-muted font-mono">{displayText}</span>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="h-6 w-6 text-muted-foreground hover:text-foreground"
        onClick={handleCopy}
        title="Copy to clipboard"
      >
        {copied ? <Check className="h-3.5 w-3.5 text-green-600" /> : <Copy className="h-3.5 w-3.5" />}
      </Button>
    </div>
  );
}
