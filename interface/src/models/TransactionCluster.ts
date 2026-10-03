import { idPrefix } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class TransactionCluster {
  readonly [idPrefix] = 'tcl';

  readonly transactionClusterId: string;
  readonly bankAccountId: string;
  name: string;
  members: Array<string>;
  createdAt: Date;

  constructor(data: WithJsonValues<TransactionCluster>) {
    this.transactionClusterId = data.transactionClusterId;
    this.bankAccountId = data.bankAccountId;
    this.name = data.name;
    this.members = data.members;
    this.createdAt = parseDate(data.createdAt);
  }
}
