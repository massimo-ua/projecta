import React from 'react';
import { RemoveButton } from './ListView/RemoveButton';

export default function RemovePaymentButton({ onRemove }) {
  return <RemoveButton onRemove={onRemove} />;
}
