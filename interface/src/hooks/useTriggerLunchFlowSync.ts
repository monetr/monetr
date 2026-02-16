import request from '@monetr/interface/util/request';

export default function useTriggerLunchFlowSync(): (linkId: string) => Promise<void> {
  return async (linkId: string): Promise<void> => {
    return request().post(`/lunch_flow/link/sync`, {
      linkId,
    });
  };
}
