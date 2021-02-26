import { Map, OrderedMap } from "immutable";
import Transaction from 'data/Transaction';

export default class TransactionState {
  items: Map<number, OrderedMap<number, Transaction>>;
  loaded: boolean;
  loading: boolean;
}
