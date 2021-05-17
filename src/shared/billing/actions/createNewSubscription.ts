import request from "shared/util/request";

export default function createNewSubscription(priceId: string, paymentMethodId: string): Promise<any> {
  return request().post(`/billing/subscribe`, {
    priceId,
    paymentMethodId,
  })
}
