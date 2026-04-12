export type PreviewProps = {
  children: string;
};

const MAX_LENGTH = 150;

export function Preview({ children }: PreviewProps) {
  const text = children.substring(0, MAX_LENGTH);
  // Pad with zero-width characters so email clients don't show body text in preview
  const padding = '\u200C\u00A0'.repeat(Math.max(0, MAX_LENGTH - text.length));

  return (
    // Must be inline -- clients that strip <style> tags would otherwise
    // reveal the preview text in the email body.
    <div
      data-skip-in-text='true'
      style={{
        display: 'none',
        overflow: 'hidden',
        lineHeight: '1px',
        opacity: 0,
        maxHeight: 0,
        maxWidth: 0,
      }}
    >
      {text}
      <div style={{ display: 'none', overflow: 'hidden', maxHeight: 0 }}>{padding}</div>
    </div>
  );
}
