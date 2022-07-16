import { Map } from 'immutable';
import FundingSchedule from 'models/FundingSchedule';


export default class FundingScheduleState {
  items: Map<number, Map<number, FundingSchedule>>;
  loaded: boolean;
  loading: boolean;

  constructor() {
    this.items = Map<number, Map<number, FundingSchedule>>();
  }
}
