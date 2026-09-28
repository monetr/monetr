import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { SelectCombobox, type SelectOption } from '@monetr/interface/components/Select';

// Build the options fresh each time since a lot of callers do that every render, so the selected value is never the same
// object as the option in the list.
function buildOptions(): Array<SelectOption<string>> {
  return [
    { label: 'Australian Dollar (AUD)', value: 'AUD' },
    { label: 'Euro (EUR)', value: 'EUR' },
    { label: 'Japanese Yen (JPY)', value: 'JPY' },
    { label: 'Swiss Franc (CHF)', value: 'CHF' },
    { label: 'US Dollar (USD)', value: 'USD' },
  ];
}

function getHighlighted(): Array<string> {
  return screen
    .getAllByRole('option')
    .filter(option => option.getAttribute('aria-selected') === 'true')
    .map(option => option.textContent ?? '');
}

describe('select combobox', () => {
  it('will highlight the selected item when opened', async () => {
    const user = userEvent.setup();
    render(
      <SelectCombobox<string>
        onChange={() => {}}
        options={buildOptions()}
        value={buildOptions().find(option => option.value === 'JPY')}
      />,
    );

    // Clicking focuses the input and the click bubbles up to the wrapper, which all try to open the menu. The highlight
    // should still end up on the selected item.
    await user.click(screen.getByRole('combobox'));
    await waitFor(() => expect(screen.getAllByRole('option')).toHaveLength(5));
    expect(getHighlighted()).toStrictEqual(['Japanese Yen (JPY)']);
  });

  it('will highlight the first item when nothing is selected', async () => {
    const user = userEvent.setup();
    render(<SelectCombobox<string> onChange={() => {}} options={buildOptions()} />);

    await user.click(screen.getByRole('combobox'));
    await waitFor(() => expect(screen.getAllByRole('option')).toHaveLength(5));
    expect(getHighlighted()).toStrictEqual(['Australian Dollar (AUD)']);
  });

  it('will select the input text when opened', async () => {
    const user = userEvent.setup();
    render(
      <SelectCombobox<string>
        onChange={() => {}}
        options={buildOptions()}
        value={buildOptions().find(option => option.value === 'EUR')}
      />,
    );

    const input = screen.getByRole<HTMLInputElement>('combobox');
    expect(input.value).toBe('Euro (EUR)');
    await user.click(input);
    await waitFor(() => expect(input.selectionStart).toBe(0));
    expect(input.selectionEnd).toBe(input.value.length);
  });

  it('will call on change when an item is picked', async () => {
    const user = userEvent.setup();
    const onChange = rs.fn();
    render(<SelectCombobox<string> onChange={onChange} options={buildOptions()} />);

    await user.click(screen.getByRole('combobox'));
    await user.click(screen.getByRole('option', { name: 'Swiss Franc (CHF)' }));
    expect(onChange).toHaveBeenCalledWith({ label: 'Swiss Franc (CHF)', value: 'CHF' });
  });
});
