import type Spending from '@monetr/interface/models/Spending';

export default function getFundingScheduleContribution(fundingScheduleId: string, spending: Array<Spending>): number {
  return spending.reduce((total: number, item: Spending) => {
    return total + (item.fundingScheduleId === fundingScheduleId && !item.isPaused ? item.nextContributionAmount : 0);
  }, 0);
}
