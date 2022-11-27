
export default class SpendingFunding {
  spendingId: number;
  bankAccountId: number;
  fundingScheduleId: number;
  nextContributionAmount: number;

  constructor(data?: Partial<SpendingFunding>) {
    if (data) {
      Object.assign(this, {
        ...data,
      });
    }
  }
}
