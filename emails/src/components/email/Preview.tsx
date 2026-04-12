import styles from './Preview.module.scss';

export type PreviewProps = {
  children: string;
};

const MAX_LENGTH = 150;

export function Preview({ children }: PreviewProps) {
  const text = children.substring(0, MAX_LENGTH);
  // Pad with zero-width characters so email clients don't show body text in preview
  const padding = '\u200C\u00A0'.repeat(Math.max(0, MAX_LENGTH - text.length));

  return (
    <div className={styles.preview} data-skip-in-text='true'>
      {text}
      <div className={styles.padding}>{padding}</div>
    </div>
  );
}
