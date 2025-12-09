import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import { selectAllIds, deselectAllIds } from './useRecipientSelection';

/**
 * **Feature: wechat-notification, Property 11: 全选/取消全选**
 * 
 * *对于任意* 接收者列表，全选操作后选中数量应等于总数，取消全选后选中数量应为 0
 * **验证: 需求 4.3, 4.4**
 */
describe('Property 11: 全选/取消全选', () => {
  it('selectAll should select all recipients - selected count equals total count', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 10000 }), { minLength: 0, maxLength: 100 }),
        (recipientIds) => {
          // Remove duplicates to simulate real recipient IDs
          const uniqueIds = [...new Set(recipientIds)];
          
          const selected = selectAllIds(uniqueIds);
          
          // After selectAll, selected count should equal total count
          expect(selected.size).toBe(uniqueIds.length);
          
          // All IDs should be in the selected set
          for (const id of uniqueIds) {
            expect(selected.has(id)).toBe(true);
          }
        }
      ),
      { numRuns: 100 }
    );
  });

  it('deselectAll should deselect all recipients - selected count equals 0', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 10000 }), { minLength: 0, maxLength: 100 }),
        (recipientIds) => {
          // Remove duplicates to simulate real recipient IDs
          const uniqueIds = [...new Set(recipientIds)];
          
          // First select all to have some selection
          selectAllIds(uniqueIds);
          
          // Then deselect all
          const deselected = deselectAllIds();
          
          // After deselectAll, selected count should be 0
          expect(deselected.size).toBe(0);
          
          // No IDs should be in the selected set
          for (const id of uniqueIds) {
            expect(deselected.has(id)).toBe(false);
          }
        }
      ),
      { numRuns: 100 }
    );
  });

  it('selectAll then deselectAll should result in empty selection', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 10000 }), { minLength: 1, maxLength: 100 }),
        (recipientIds) => {
          const uniqueIds = [...new Set(recipientIds)];
          
          // Select all
          const afterSelectAll = selectAllIds(uniqueIds);
          expect(afterSelectAll.size).toBe(uniqueIds.length);
          
          // Deselect all
          const afterDeselectAll = deselectAllIds();
          expect(afterDeselectAll.size).toBe(0);
        }
      ),
      { numRuns: 100 }
    );
  });
});
