import { Map } from 'immutable';
import Balance from 'models/Balance';

export default class BalancesState {
  items: Map<number, Balance>;
  loaded: boolean;
  loading: boolean;

  constructor() {
    this.items = Map<number, Balance>();
  }
}
