import {Record} from "immutable";
import moment from "moment";


export default class Transaction extends Record({
  transactionId: 0,
  bankAccountId: 0,
  amount: 0,
  expenseId: null,
  categories: [],
  originalCategories: [],
  date: moment(),
  authorizedDate: moment(),
  name: null,
  originalName: '',
  merchantName: null,
  originalMerchantName: null,
  isPending: false,
  createdAt: moment(),
}) {

}
