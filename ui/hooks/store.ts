import create from 'zustand';

export interface Store {
  selectedBankAccountId: number | null;
  setCurrentBankAccount: (_bankAccountId: number) => void;
}

const useStore = create<Store>((set): Store => ({
  selectedBankAccountId: null,
  setCurrentBankAccount: (bankAccountId: number) => set(() => {
    window.localStorage.setItem('selectedBankAccountId', bankAccountId.toString(10));
    return { selectedBankAccountId: bankAccountId };
  }),
}));

const bankAccountId = +window.localStorage.getItem('selectedBankAccountId');
if (bankAccountId) {
  useStore.setState({ selectedBankAccountId: bankAccountId });
}

export default useStore;
