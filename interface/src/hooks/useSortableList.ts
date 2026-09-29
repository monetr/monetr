import type React from 'react';
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';

import type { DragController } from '@monetr/interface/util/sortableDrag';

// Everything to do with dragging lives in its own chunk, see sortableDrag.ts. Sortable items have to be direct children
// of the container, that's how it finds and measures them. It's shared across every sortable list on the page so it
// only ever gets fetched once.
let dragModule: Promise<typeof import('@monetr/interface/util/sortableDrag')> | null = null;
export function preloadSortableList(): void {
  loadDragModule().catch(() => undefined);
}

function loadDragModule(): Promise<typeof import('@monetr/interface/util/sortableDrag')> {
  dragModule ??= import('@monetr/interface/util/sortableDrag').catch(error => {
    // Let the next hover try again instead of caching the failure forever.
    dragModule = null;
    throw error;
  });
  return dragModule;
}

export interface UseSortableListOptions<T, Id extends string> {
  // The items in their default order. Anything missing from `order` shows up after the ordered items, in this order.
  items: Array<T>;
  // The saved order, if there is one. Ids in here that don't match an item are ignored.
  order?: Array<Id> | null;
  getId: (item: T) => Id;
  onReorder: (ids: Array<Id>) => void;
}

export interface SortableItemProps {
  'data-sortable-id': string;
  'data-dragging'?: 'true';
  onPointerEnter: () => void;
  onPointerDown: (event: React.PointerEvent<HTMLElement>) => void;
  onDragStart: (event: React.DragEvent<HTMLElement>) => void;
}

export interface UseSortableListResult<T, Id extends string> {
  items: Array<T>;
  containerRef: React.RefObject<HTMLDivElement | null>;
  draggingId: Id | null;
  getItemProps: (id: Id) => SortableItemProps;
}

export function useSortableList<T, Id extends string>(
  options: UseSortableListOptions<T, Id>,
): UseSortableListResult<T, Id> {
  const { items, order, getId } = options;
  const [draggingId, setDraggingId] = useState<Id | null>(null);
  // The order we just committed locally. It sticks around until the order passed in changes, that way a drop shows the
  // new order in the same render instead of waiting on whoever owns `order` to catch up.
  const [committed, setCommitted] = useState<{ base: Array<Id> | null | undefined; ids: Array<Id> } | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const controllerRef = useRef<DragController | null>(null);
  const mountedRef = useRef(true);

  const onReorderRef = useRef(options.onReorder);
  onReorderRef.current = options.onReorder;
  const orderRef = useRef(order);
  orderRef.current = order;

  const currentOrder = committed && committed.base === order ? committed.ids : order;

  const ordered = useMemo(() => {
    if (!currentOrder || currentOrder.length === 0) {
      return items;
    }
    const byId = new Map(items.map(item => [getId(item), item]));
    const result: Array<T> = [];
    for (const id of currentOrder) {
      const item = byId.get(id);
      if (item !== undefined) {
        result.push(item);
        byId.delete(id);
      }
    }
    return result.concat(Array.from(byId.values()));
  }, [items, currentOrder, getId]);

  const commit = useCallback((ids: Array<Id>) => {
    setCommitted({ base: orderRef.current, ids });
    onReorderRef.current(ids);
  }, []);

  const getController = useCallback(
    (module: typeof import('@monetr/interface/util/sortableDrag')): DragController => {
      controllerRef.current ??= module.createDragController({
        getContainer: () => containerRef.current,
        onStart: id => setDraggingId(id as Id),
        onEnd: ids => {
          if (ids) {
            commit(ids as Array<Id>);
          }
          setDraggingId(null);
        },
      });
      return controllerRef.current;
    },
    [commit],
  );

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      controllerRef.current?.dispose();
    };
  }, []);

  // A drop leaves the items with transforms from the drag, the controller animates them back once the new order has
  // rendered. Nothing to do if nobody has dragged yet, the controller isn't even loaded then.
  useLayoutEffect(() => {
    controllerRef.current?.settle();
  });

  const onPointerDown = useCallback(
    (event: React.PointerEvent<HTMLElement>, id: Id) => {
      if (!event.isPrimary || event.button !== 0) {
        return;
      }
      const start = {
        id,
        pointerId: event.pointerId,
        pointerType: event.pointerType,
        element: event.currentTarget,
        clientX: event.clientX,
        clientY: event.clientY,
      };
      if (controllerRef.current) {
        controllerRef.current.start(start);
        return;
      }

      // The drag code hasn't finished loading yet, usually because this is a tap that came right on the heels of the
      // pointerenter that started the load. Hand the press over once it's here, unless it already ended, in which case
      // it was just a click and there's nothing to do.
      let released = false;
      const onRelease = (release: PointerEvent) => {
        if (release.pointerId === start.pointerId) {
          released = true;
        }
      };
      window.addEventListener('pointerup', onRelease);
      window.addEventListener('pointercancel', onRelease);
      loadDragModule()
        .then(module => {
          if (!released && mountedRef.current) {
            getController(module).start(start);
          }
        })
        .catch(() => {
          // Couldn't load the chunk, the item still works as a plain link.
        })
        .finally(() => {
          window.removeEventListener('pointerup', onRelease);
          window.removeEventListener('pointercancel', onRelease);
        });
    },
    [getController],
  );

  const getItemProps = useCallback(
    (id: Id): SortableItemProps => ({
      'data-sortable-id': id,
      'data-dragging': draggingId === id ? 'true' : undefined,
      // Start fetching the drag code as soon as someone points at an item, it's almost always here before they press.
      onPointerEnter: () => void loadDragModule().then(getController, () => undefined),
      onPointerDown: event => onPointerDown(event, id),
      // The native HTML drag and drop would fight with ours, especially on the links inside each item.
      onDragStart: event => event.preventDefault(),
    }),
    [draggingId, getController, onPointerDown],
  );

  return { items: ordered, containerRef, draggingId, getItemProps };
}
