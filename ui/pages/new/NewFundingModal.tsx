import NiceModal, { useModal } from "@ebay/nice-modal-react";
import { MBaseButton } from "components/MButton";
import MModal, { MModalRef } from "components/MModal";
import MSpan from "components/MSpan";
import FundingSchedule from "models/FundingSchedule";
import moment from "moment";
import React, { useRef } from "react";

function NewFundingModal(): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);

  function close() {
    modal.resolve(new FundingSchedule({
      fundingScheduleId: 123,
      bankAccountId: 1234,
      name: 'test',
      nextOccurrence: moment(),
    }));
    modal.remove();
  }

  return (
    <MModal open={ modal.visible } ref={ ref } className='sm:max-w-2xl py-4'>
      <MSpan className='font-bold text-xl p-2'>
        Create A New Funding Schedule
      </MSpan>
      <div className='flex flex-col gap-2 p-2'>
        <MBaseButton onClick={ close }>
          close
        </MBaseButton>
      </div>
    </MModal>
  );
}

const newFundingModal = NiceModal.create(NewFundingModal);

export default newFundingModal;

export function showNewFundingModal(): Promise<FundingSchedule | null> {
  return NiceModal.show<FundingSchedule | null, {}>(newFundingModal);
}
