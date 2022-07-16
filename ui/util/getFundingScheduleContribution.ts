import Spending from 'models/Spending';

export default function getFundingScheduleContribution(
  fundingScheduleId: number,
  spending: Map<number, Spending>,
): number {
  return (Array.from(spending.values())).reduce((total: number, item: Spending) => {
    return total + (item.fundingScheduleId === fundingScheduleId && !item.isPaused ? item.nextContributionAmount : 0);
  }, 0);
}
