import { Plus } from 'lucide-react';

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

  it('will run an action when it is clicked', async () => {
    const user = userEvent.setup();
    const onChange = rs.fn();
    const onSelect = rs.fn();
    render(
      <SelectCombobox<string>
        actions={[
          {
            label: 'Add A Currency',
            icon: Plus,
            onSelect,
          },
        ]}
        onChange={onChange}
        options={buildOptions()}
        value={buildOptions().find(option => option.value === 'EUR')}
      />,
    );

    const input = screen.getByRole<HTMLInputElement>('combobox');
    await user.click(input);
    await user.click(screen.getByRole('option', { name: 'Add A Currency' }));
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onChange).not.toHaveBeenCalled();
    // The action should never take the place of the actual value
    expect(input.value).toBe('Euro (EUR)');
    expect(screen.queryAllByRole('option')).toHaveLength(0);
  });

  it('will run an action when enter is pressed on it', async () => {
    const user = userEvent.setup();
    const onChange = rs.fn();
    const onSelect = rs.fn();
    render(
      <SelectCombobox<string>
        actions={[
          {
            label: 'Add A Currency',
            icon: Plus,
            onSelect,
          },
        ]}
        onChange={onChange}
        options={buildOptions()}
      />,
    );

    // Nothing is selected so the action at the top is what gets highlighted
    await user.click(screen.getByRole('combobox'));
    await waitFor(() => expect(getHighlighted()).toStrictEqual(['Add A Currency']));
    await user.keyboard('{Enter}');
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onChange).not.toHaveBeenCalled();
  });

  it('will not run an action when tabbing away from it', async () => {
    const user = userEvent.setup();
    const onChange = rs.fn();
    const onSelect = rs.fn();
    render(
      <SelectCombobox<string>
        actions={[
          {
            label: 'Add A Currency',
            icon: Plus,
            onSelect,
          },
        ]}
        onChange={onChange}
        options={buildOptions()}
      />,
    );

    const input = screen.getByRole<HTMLInputElement>('combobox');
    await user.click(input);
    await waitFor(() => expect(getHighlighted()).toStrictEqual(['Add A Currency']));
    await user.tab();
    expect(onSelect).not.toHaveBeenCalled();
    expect(onChange).not.toHaveBeenCalled();
    expect(input.value).toBe('');
  });

  it('will highlight an action that matches the filter', async () => {
    const user = userEvent.setup();
    const onChange = rs.fn();
    const onSelect = rs.fn();
    render(
      <SelectCombobox<string>
        actions={[
          {
            label: 'Add A Currency',
            icon: Plus,
            onSelect,
          },
        ]}
        onChange={onChange}
        options={buildOptions()}
        value={buildOptions().find(option => option.value === 'EUR')}
      />,
    );

    // Opening selects the text in the input, so typing replaces it
    const input = screen.getByRole<HTMLInputElement>('combobox');
    await user.click(input);
    await waitFor(() => expect(input.selectionStart).toBe(0));
    await user.keyboard('add');
    await waitFor(() => expect(getHighlighted()).toStrictEqual(['Add A Currency']));
    await user.keyboard('{Enter}');
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onChange).not.toHaveBeenCalled();
    // What they typed should be replaced with the actual value again
    await waitFor(() => expect(input.value).toBe('Euro (EUR)'));
  });

  it('will match options on their group when filtering', async () => {
    const user = userEvent.setup();
    render(
      <SelectCombobox<string>
        onChange={() => {}}
        options={[
          {
            label: 'US Dollar (USD)',
            value: 'USD',
            group: 'Suggested',
          },
          ...buildOptions().filter(option => option.value !== 'USD'),
        ]}
      />,
    );

    await user.click(screen.getByRole('combobox'));
    await user.keyboard('suggested');
    await waitFor(() => expect(getHighlighted()).toStrictEqual(['US Dollar (USD)']));
    expect(screen.getAllByRole('option')).toHaveLength(1);
  });

  it('will show a header above a group of options', async () => {
    const user = userEvent.setup();
    render(
      <SelectCombobox<string>
        onChange={() => {}}
        options={[
          {
            label: 'US Dollar (USD)',
            value: 'USD',
            group: 'Suggested',
          },
          ...buildOptions().filter(option => option.value !== 'USD'),
        ]}
      />,
    );

    await user.click(screen.getByRole('combobox'));
    await waitFor(() => expect(screen.getByText('Suggested')).toBeVisible());
    expect(screen.getAllByRole('option')).toHaveLength(5);
  });
});
