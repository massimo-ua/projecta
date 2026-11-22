/* eslint-disable import/extensions */
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import listViewContent from './ListView.content.js';

describe('ListView pagination localization', () => {
  it('defines English and Ukrainian labels for "To the beginning" navigation', () => {
    assert.deepEqual(listViewContent.content.toBeginning?.translation, {
      en: 'To the beginning',
      uk: 'На початок',
    });
  });

  it('defines English and Ukrainian labels for "To the end" navigation', () => {
    assert.deepEqual(listViewContent.content.toEnd?.translation, {
      en: 'To the end',
      uk: 'В кінець',
    });
  });
});
