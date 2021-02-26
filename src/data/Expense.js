import { Record } from "immutable";
import moment from 'moment';


export default class Expense extends Record({
  expenseId: 0,
  bankAccountId: 0,
  fundingScheduleId: 0,
  name: '',
  description: null,
  targetAmount: 0,
  currentAmount: 0,
  recurrenceRule: '',
  lastRecurrence: null,
  nextRecurrence: moment(),
  nextContributionAmount: 0,
  isBehind: false,
}) {

}
