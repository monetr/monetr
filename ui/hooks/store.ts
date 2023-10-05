import create from 'zustand';

export interface Store {
  mobileSidebarOpen: boolean;
  setMobileSidebarOpen: (_isOpen: boolean) => void;
}

const useStore = create<Store>((set): Store => ({
  mobileSidebarOpen: false,
  setMobileSidebarOpen: (isOpen: boolean) => set(() => ({ mobileSidebarOpen: isOpen })),
}));


export default useStore;
