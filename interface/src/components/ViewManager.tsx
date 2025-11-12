import type React from 'react';
import { type ComponentType, createContext, type ReactNode, useCallback, useContext, useMemo, useState } from 'react';

import { usePrevious } from '@monetr/interface/hooks/usePrevious';

type ViewComponent = ComponentType<unknown>;

export interface ViewContextType<T extends string, Metadata extends {}, Form extends {}> {
  currentView: T;
  formData: Form;
  metadata: Metadata;
  updateFormData: (newData: Partial<Form>) => void;
  updateMetadata: (updates: Partial<Metadata>) => void;
  prevView: () => void;
  goToView: (view: T) => void;
  isInitialView: boolean;
  isFinalView: boolean;
  reset: () => void;
}

const ViewContext = createContext<ViewContextType<string, unknown, unknown> | undefined>(undefined);

function useViewContext<T extends string, Metadata extends {}, Form extends {}>() {
  const context = useContext(ViewContext);
  if (!context) {
    throw new Error('useViewContext must be used within a ViewManager');
  }
  return context as ViewContextType<T, Metadata, Form>;
}

interface ViewManagerProps<T extends string, Metadata> {
  viewComponents: Record<T, ViewComponent>;
  initialView: T;
  initialMetadata?: Metadata;
  // Layout is an optional wrapper component that can accept the view context as a property.
  layout?: React.FC<{
    children?: ReactNode;
  }>;
}

function ViewManager<T extends string, Metadata, Form>({
  viewComponents,
  initialView,
  initialMetadata = {} as Metadata,
  layout = null,
}: ViewManagerProps<T, Metadata>) {
  const [currentView, setCurrentView] = useState<T>(initialView);
  const [formData, setFormData] = useState<Form>({} as Form);
  const [metadata, setMetadata] = useState<Metadata>(initialMetadata);
  const viewOrder = useMemo(() => Object.keys(viewComponents) as T[], [viewComponents]);

  const previousView = usePrevious(currentView);

  // biome-ignore lint/correctness/useExhaustiveDependencies: Needs current view to trigger properly
  const prevView = useCallback(() => {
    if (previousView) {
      setCurrentView(previousView);
      return;
    }
  }, [currentView, previousView]);

  const goToView = useCallback(
    (view: T) => {
      if (viewOrder.includes(view)) {
        setCurrentView(view);
      }
    },
    [viewOrder],
  );

  const updateFormData = useCallback((newData: Partial<Form>) => {
    setFormData(prevData => ({ ...prevData, ...newData }));
  }, []);

  const updateMetadata = useCallback((updates: Partial<Metadata>) => {
    setMetadata(prevMetadata => ({ ...prevMetadata, ...updates }));
  }, []);

  const reset = useCallback(() => {
    setCurrentView(initialView);
    setFormData({} as Form);
    setMetadata(initialMetadata);
  }, [initialView, initialMetadata]);

  const value: ViewContextType<T, Metadata, Form> = useMemo(
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
