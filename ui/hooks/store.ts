import create from 'zustand';

export interface Store {
  selectedBankAccountId: number | null;
  setCurrentBankAccount: (_bankAccountId: number) => void;
}

const useStore = create<Store>((set): Store => ({
  selectedBankAccountId: null,
  setCurrentBankAccount: (bankAccountId: number) => set(() => ({ selectedBankAccountId: bankAccountId })),
}));

export default useStore;
