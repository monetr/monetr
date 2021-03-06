import { Map } from 'immutable';
import Spending from 'models/Spending';

export default class SpendingState {
  items: Map<number, Map<number, Spending>>;
  loaded: boolean;
  loading: boolean;

  selectedExpenseId: number | null;
  selectedGoalId: number | null;

  constructor() {
    this.items = Map<number, Map<number, Spending>>();
  }
}
