import request from "shared/util/request";
import BillingPlan from "models/BillingPlan";


export default function fetchBillingPlans(): Promise<BillingPlan[]> {
  return request().get('/billing/plans')
    .then(result => {
      return result.data.map((item: object) => new BillingPlan(item));
    })
}
