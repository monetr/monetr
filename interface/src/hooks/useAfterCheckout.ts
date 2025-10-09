import { useMutation, useQueryClient } from '@tanstack/react-query';

import request from '@monetr/interface/util/request';

export interface AfterCheckoutResult {
  message: string | null;
  nextUrl: string;
  isActive: boolean;
}

// useAfterCheckout is a hook that provides a function where the caller can give a Stripe checkout session ID which is
// used to refresh the state of the currently authenticated user's subscription. This is intended to be used after a
// user has been redirected back to the application from Stripe to see if their subscription is now/still active.
// The function yielded by this hook will return the result of that "after checkout" data. But will also mutate the
// `isActive` variable from `useAuthentication` to properly represent the new subscription status.
export function useAfterCheckout(): (_checkoutSessionId: string) => Promise<AfterCheckoutResult> {
  const queryClient = useQueryClient();

  async function queryCheckoutSession(checkoutSessionId: string): Promise<AfterCheckoutResult> {
    return request()
      .get<AfterCheckoutResult>(`/billing/checkout/${checkoutSessionId}`)
      .then(result => result.data);
  }

  const mutation = useMutation({
    mutationFn: queryCheckoutSession,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['/users/me'] }),
  });

  return mutation.mutateAsync;
}
