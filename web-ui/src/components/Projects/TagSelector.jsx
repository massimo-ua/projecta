import React, { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { Plus, X, Tag as TagIcon } from 'lucide-react';
import { investmentRepository } from '../../api';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';

export default function TagSelector({
  tags = [],
  onChange,
  projectId: propProjectId,
  availableTags: propAvailableTags,
  label,
  placeholder,
  id = 'tag-selector',
}) {
  const { projectId: routeProjectId } = useParams();
  const projectId = propProjectId || routeProjectId;
  const content = useIntlayer('payments');

  const [tagInput, setTagInput] = useState('');
  const [fetchedTags, setFetchedTags] = useState([]);

  useEffect(() => {
    if (propAvailableTags !== undefined || !projectId) {
      return undefined;
    }

    let isMounted = true;
    investmentRepository
      .getTags(projectId)
      .then((tagsList) => {
        if (isMounted && Array.isArray(tagsList)) {
          setFetchedTags(tagsList);
        }
      })
      .catch(() => {
        // Ignore tag fetch failures
      });

    return () => {
      isMounted = false;
    };
  }, [projectId, propAvailableTags]);

  const poolOfTags = propAvailableTags !== undefined ? propAvailableTags : fetchedTags;

  const handleAdd = (tagToAdd) => {
    const raw = typeof tagToAdd === 'string' ? tagToAdd : tagInput;
    const cleanTag = raw.trim().replace(/^#/, '');
    if (!cleanTag) return;

    if (!tags.includes(cleanTag)) {
      onChange([...tags, cleanTag]);
    }
    setTagInput('');
  };

  const handleRemove = (tagToRemove) => {
    onChange(tags.filter((t) => t !== tagToRemove));
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      handleAdd();
    }
  };

  // Filter available suggestions: must not already be selected
  const suggestedTags = useMemo(() => {
    if (!Array.isArray(poolOfTags)) return [];
    const query = tagInput.trim().toLowerCase().replace(/^#/, '');

    return poolOfTags.filter((t) => {
      if (!t || tags.includes(t)) return false;
      if (!query) return true;
      return t.toLowerCase().includes(query);
    });
  }, [poolOfTags, tags, tagInput]);

  return (
    <div className="space-y-2">
      {label !== false && (
        <Label htmlFor={id} className="text-xs font-semibold flex items-center gap-1.5">
          <TagIcon className="h-3.5 w-3.5 text-muted-foreground" />
          <span>{label || String(content?.tagsLabel || 'Tags')}</span>
        </Label>
      )}

      {/* Input row */}
      <div className="flex gap-2">
        <Input
          id={id}
          placeholder={placeholder || String(content?.tagsPlaceholder || 'Type tag and press Enter')}
          value={tagInput}
          onChange={(e) => setTagInput(e.target.value)}
          onKeyDown={handleKeyDown}
          className="rounded-xl"
        />
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => handleAdd()}
          className="rounded-xl shrink-0 px-3"
          disabled={!tagInput.trim()}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      {/* Selected tags badges */}
      {tags.length > 0 && (
        <div className="flex gap-1.5 flex-wrap pt-0.5">
          {tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="gap-1 rounded-md text-xs py-0.5">
              <span>{`#${tag}`}</span>
              <button
                type="button"
                onClick={() => handleRemove(tag)}
                className="text-muted-foreground hover:text-foreground inline-flex items-center"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}

      {/* Available existing tags suggestions */}
      {suggestedTags.length > 0 && (
        <div className="flex flex-wrap items-center gap-1.5 pt-1">
          <span className="text-[11px] text-muted-foreground font-medium">
            {String(content?.suggestedTags || 'Existing tags:')}
          </span>
          {suggestedTags.slice(0, 12).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => handleAdd(t)}
              className="inline-flex items-center gap-0.5 px-2 py-0.5 rounded-md text-xs font-medium bg-muted/60 hover:bg-primary/10 hover:text-primary transition-colors text-muted-foreground border border-dashed border-border"
              title={`Add #${t}`}
            >
              <Plus className="h-3 w-3" />
              <span>{`#${t}`}</span>
            </button>
          ))}
          {suggestedTags.length > 12 && (
            <span className="text-[10px] text-muted-foreground">
              {`+${suggestedTags.length - 12} more`}
            </span>
          )}
        </div>
      )}
    </div>
  );
}

TagSelector.propTypes = {
  tags: PropTypes.arrayOf(PropTypes.string),
  onChange: PropTypes.func.isRequired,
  projectId: PropTypes.string,
  availableTags: PropTypes.arrayOf(PropTypes.string),
  label: PropTypes.node,
  placeholder: PropTypes.string,
  id: PropTypes.string,
};

TagSelector.defaultProps = {
  tags: [],
  projectId: undefined,
  availableTags: undefined,
  label: undefined,
  placeholder: undefined,
  id: 'tag-selector',
};
