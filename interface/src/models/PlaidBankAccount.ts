import { parseJSON } from 'date-fns';

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
        createdAt: data?.createdAt && parseJSON(data.createdAt),
      });
    }
  }
}
