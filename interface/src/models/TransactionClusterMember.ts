import parseDate from '@monetr/interface/util/parseDate';

export default class TransactionClusterMember {
  transactionId: string;
  bankAccountId: string;
  transactionClusterId: string;
  createdAt: Date;
  updatedAt: Date;

  constructor(data?: Partial<TransactionClusterMember>) {
    if (data) {
      Object.assign(this, {
        ...data,
        createdAt: Boolean(data?.createdAt) && parseDate(data.createdAt),
        updatedAt: Boolean(data?.updatedAt) && parseDate(data.updatedAt),
      });
    }
  }
}
