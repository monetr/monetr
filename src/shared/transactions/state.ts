import {Map, OrderedMap, Record} from "immutable";
import Transaction from 'data/Transaction';

export default class TransactionState extends Record({
    items: Map<number, OrderedMap<number, Transaction>>(),
    loaded: false,
    loading: false,
}) {
    items: Map<number, OrderedMap<number, Transaction>>;
    loaded: boolean;
    loading: boolean;
};
