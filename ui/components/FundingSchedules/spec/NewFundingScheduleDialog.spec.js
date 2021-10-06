import { render } from '@testing-library/react'
import { NewFundingScheduleDialog } from "components/FundingSchedules/NewFundingScheduleDialog";

describe('new funding schedule dialog', () => {
  it('will render', () => {
    const onClose = jest.fn();
    const createFundingSchedule = jest.fn();
    render(<NewFundingScheduleDialog
      bankAccountId={ 1 }
      createFundingSchedule={ createFundingSchedule }
      onClose={ onClose }
      isOpen
    />)

    expect(document.querySelector('.new-funding-schedule')).not.toBeEmptyDOMElement();
  });

  it('will open date picker', () => {
    const onClose = jest.fn();
    const createFundingSchedule = jest.fn();
    const result = render(<NewFundingScheduleDialog
      bankAccountId={ 1 }
      createFundingSchedule={ createFundingSchedule }
      onClose={ onClose }
      isOpen
    />);

    const button = result.getByTestId('new-funding-schedule-next-button');
    button.click();

    const datePicker = result.getByTestId('new-funding-schedule-date-picker');
    const dialogButton = datePicker.querySelector('button');

    // Make sure the dialog can open. This will fail if the date-io library does not have the utils.getYearText
    // function.
    dialogButton.click();
  });
})
