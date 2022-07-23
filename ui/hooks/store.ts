import create from 'zustand';

export interface Store {
  selectedExpenseId: number | null;
  selectedGoalId: number | null;
  selectedBankAccountId: number | null;
  setCurrentBankAccount: (_bankAccountId: number) => void;
  setCurrentExpense: (_spendingId: number) => void;
  setCurrentGoal: (_spendingId: number) => void;
}

const useStore = create<Store>((set): Store => ({
  selectedExpenseId: null,
  selectedGoalId: null,
  selectedBankAccountId: null,
  setCurrentBankAccount: (bankAccountId: number) => set(() => {
    window.localStorage.setItem('selectedBankAccountId', bankAccountId.toString(10));
    return { selectedBankAccountId: bankAccountId };
  }),
  setCurrentExpense: (spendingId: number) => set(() => ({ selectedExpenseId: spendingId })),
  setCurrentGoal: (spendingId: number) => set(() => ({ selectedGoalId: spendingId })),
}));

const bankAccountId = +window.localStorage.getItem('selectedBankAccountId');
if (bankAccountId) {
  useStore.setState({ selectedBankAccountId: bankAccountId });
}

export default useStore;
