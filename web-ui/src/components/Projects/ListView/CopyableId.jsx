import React from 'react';
import { CopyableText } from './CopyableText';

export function CopyableId({ id, label = 'ID' }) {
  return <CopyableText text={id} label={label} truncate />;
}
