// eslint-disable-next-line max-len
import React, { type ComponentType, createContext, ReactNode, useCallback, useContext, useMemo, useState } from 'react';

import { usePrevious } from '@monetr/interface/hooks/usePrevious';

type ViewComponent = ComponentType<any>;

export interface ViewContextType<T extends string, M extends Record<string, any>> {
  currentView: T;
  formData: Record<string, any>;
  metadata: M;
  updateFormData: (newData: Record<string, any>) => void;
  updateMetadata: (updates: Partial<M>) => void;
  prevView: () => void;
  goToView: (view: T) => void;
  isInitialView: boolean;
  isFinalView: boolean;
  reset: () => void;
}

const ViewContext = createContext<ViewContextType<any, any> | undefined>(undefined);

function useViewContext<T extends string, M extends Record<string, any>>() {
  const context = useContext(ViewContext);
  if (!context) {
    throw new Error('useViewContext must be used within a ViewManager');
  }
  return context as ViewContextType<T, M>;
}

interface ViewManagerProps<T extends string, M extends Record<string, any>> {
  viewComponents: Record<T, ViewComponent>;
  initialView: T;
  initialMetadata?: M;
  // Layout is an optional wrapper component that can accept the view context as a property.
  layout?: React.FC<{
    children: ReactNode | undefined;
  }>;
}

function ViewManager<T extends string, M extends Record<string, any>>({
  viewComponents,
  initialView,
  initialMetadata = {} as M,
  layout = null,
}: ViewManagerProps<T, M>) {
  const [currentView, setCurrentView] = useState<T>(initialView);
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [metadata, setMetadata] = useState<M>(initialMetadata);

  const viewOrder = useMemo(() => Object.keys(viewComponents) as T[], [viewComponents]);

  const previousView = usePrevious(currentView);

  const prevView = useCallback(() => {
    if (previousView) {
      setCurrentView(previousView);
      return;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentView, previousView]);

  const goToView = useCallback(
    (view: T) => {
      if (viewOrder.includes(view)) {
        setCurrentView(view);
      }
    },
    [viewOrder],
  );

  const updateFormData = useCallback((newData: Record<string, any>) => {
    setFormData(prevData => ({ ...prevData, ...newData }));
  }, []);

  const updateMetadata = useCallback((updates: Partial<M>) => {
    setMetadata(prevMetadata => ({ ...prevMetadata, ...updates }));
  }, []);

  const reset = useCallback(() => {
    setCurrentView(initialView);
    setFormData({});
    setMetadata(initialMetadata);
  }, [initialView, initialMetadata]);

  const value: ViewContextType<T, M> = useMemo(
    () => ({
      currentView,
      formData,
      metadata,
      updateFormData,
      updateMetadata,
      prevView,
      goToView,
      isInitialView: currentView === viewOrder[0],
      isFinalView: currentView === viewOrder[viewOrder.length - 1],
      reset,
    }),
    [currentView, formData, metadata, updateFormData, updateMetadata, prevView, goToView, viewOrder, reset],
  );

  const CurrentViewComponent: ViewComponent = viewComponents[currentView];

  if (layout) {
    const Layout = layout;
    return (
      <ViewContext.Provider value={value}>
        <Layout>
          <CurrentViewComponent />
        </Layout>
      </ViewContext.Provider>
    );
  }

  return (
    <ViewContext.Provider value={value}>
      <CurrentViewComponent />
    </ViewContext.Provider>
  );
}

export { useViewContext, ViewManager };
