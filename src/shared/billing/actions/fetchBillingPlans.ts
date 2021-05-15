import request from "shared/util/request";
import BillingPlan from "data/BillingPlan";


export default function fetchBillingPlans() {
  return request().get('/billing/plans')
    .then(result => {
      return result.data.map(item => new BillingPlan(item));
    })
}
