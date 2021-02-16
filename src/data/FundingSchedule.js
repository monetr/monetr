import {Record} from "immutable";
import moment from "moment";

export default class FundingSchedule extends Record({
  fundingScheduleId: 0,
  bankAccountId: 0,
  name: '',
  description: null,
  rule: '',
  lastOccurrence: null,
  nextOccurrence: moment(),
}) {

}
