import { useState, useCallback } from 'react';

/**
 * Hook for managing recipient selection state
 * Extracted for testability of select all/deselect all logic
 */
export function useRecipientSelection(recipientIds: number[]) {
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  const selectAll = useCallback(() => {
    setSelectedIds(new Set(recipientIds));
  }, [recipientIds]);

  const deselectAll = useCallback(() => {
    setSelectedIds(new Set());
  }, []);

  const toggle = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const isSelected = useCallback((id: number) => selectedIds.has(id), [selectedIds]);

  return {
    selectedIds,
    selectedCount: selectedIds.size,
    selectAll,
    deselectAll,
    toggle,
    isSelected,
  };
}

/**
 * Pure functions for selection logic - easier to test with property-based testing
 */
export function selectAllIds(recipientIds: number[]): Set<number> {
  return new Set(recipientIds);
}

export function deselectAllIds(): Set<number> {
  return new Set();
}

export function toggleId(selectedIds: Set<number>, id: number): Set<number> {
  const next = new Set(selectedIds);
  if (next.has(id)) {
    next.delete(id);
  } else {
    next.add(id);
  }
  return next;
}
