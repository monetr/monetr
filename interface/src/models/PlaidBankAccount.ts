import parseDate from '@monetr/interface/util/parseDate';

export default class PlaidBankAccount {
  name: string;
  officialName?: string;
  mask?: string;
  availableBalance: number;
  currentBalance: number;
  limitBalance?: number;
  createdAt: Date;
  createdByUserId: number;

  constructor(data?: Partial<PlaidBankAccount>) {
    if (data) {
      Object.assign(this, {
        ...data,
        createdAt: parseDate(data?.createdAt),
      });
    }
  }
}
