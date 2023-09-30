import create from 'zustand';

export interface Store {
  selectedExpenseId: number | null;
  selectedGoalId: number | null;
  selectedBankAccountId: number | null;
  mobileSidebarOpen: boolean;
  setCurrentBankAccount: (_bankAccountId: number) => void;
  setCurrentExpense: (_spendingId: number) => void;
  setCurrentGoal: (_spendingId: number) => void;
  setMobileSidebarOpen: (_isOpen: boolean) => void;
}

const useStore = create<Store>((set): Store => ({
  selectedExpenseId: null,
  selectedGoalId: null,
  selectedBankAccountId: null,
  mobileSidebarOpen: false,
  setCurrentBankAccount: (bankAccountId: number) => set(() => {
    window.localStorage.setItem('selectedBankAccountId', bankAccountId.toString(10));
    return { selectedBankAccountId: bankAccountId };
  }),
  setCurrentExpense: (spendingId: number) => set(() => ({ selectedExpenseId: spendingId })),
  setCurrentGoal: (spendingId: number) => set(() => ({ selectedGoalId: spendingId })),
  setMobileSidebarOpen: (isOpen: boolean) => set(() => ({ mobileSidebarOpen: isOpen })),
}));

const bankAccountId = +window.localStorage.getItem('selectedBankAccountId');
if (bankAccountId) {
  useStore.setState({ selectedBankAccountId: bankAccountId });
}

export default useStore;
