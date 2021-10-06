import Spending from "data/Spending";
import { Map } from 'immutable';

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
