import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class TransactionClusterMember {
  transactionId: string;
  bankAccountId: string;
  transactionClusterId: string;
  createdAt: Date;
  updatedAt: Date;

  constructor(data: WithJsonValues<TransactionClusterMember>) {
    this.transactionId = data.transactionId;
    this.bankAccountId = data.bankAccountId;
    this.transactionClusterId = data.transactionClusterId;
    this.createdAt = parseDate(data.createdAt);
    this.updatedAt = parseDate(data.updatedAt);
  }
}
