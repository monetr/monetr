import styles from './MergeDiagram.module.scss';

// Tone drives the accent on a value or a badge. Per the docs palette rule cyan
// means positive/yes and pink means warning/negative, so a write goes cyan and
// a rejection goes pink. Everything else stays in the purple family or muted.
type Tone = 'updated' | 'cleared' | 'untouched' | 'rejected';

// Kind only exists so the schema sheet can render prose values instead of the
// monospace literals the other two sheets show.
type Kind = 'literal' | 'prose';

interface Field {
  name: string;
  // An undefined value means the client never sent the key at all, which
  // renders as an empty slot rather than a value. That distinction between
  // absent and null is the entire point of the figure.
  value?: string;
  tone?: Tone;
}

interface Sheet {
  title: string;
  subtitle: string;
  kind: Kind;
  fields: Array<Field>;
}

const SHEETS: Array<Sheet> = [
  {
    title: 'The existing row',
    subtitle: 'read from the database first',
    kind: 'literal',
    fields: [
      { name: 'name', value: '"Checking"' },
      { name: 'mask', value: '"4242"' },
      { name: 'currentBalance', value: '125000' },
      { name: 'limitBalance', value: 'null' },
      { name: 'accountSubType', value: '"checking"' },
      { name: 'status', value: '"active"' },
    ],
  },
  {
    title: 'The schema',
    subtitle: 'PatchManualBankAccount',
    kind: 'prose',
    fields: [
      { name: 'name', value: 'writable, null rejected' },
      { name: 'mask', value: 'writable, null allowed' },
      { name: 'currentBalance', value: 'writable, null rejected' },
      { name: 'limitBalance', value: 'writable, null allowed' },
      { name: 'accountSubType', value: 'writable, null rejected' },
      { name: 'status', value: 'not writable at all', tone: 'rejected' },
    ],
  },
  {
    title: 'The request body',
    subtitle: 'only what the client actually sent',
    kind: 'literal',
    fields: [
      { name: 'name' },
      { name: 'mask', value: 'null', tone: 'cleared' },
      { name: 'currentBalance', value: '98000', tone: 'updated' },
      { name: 'limitBalance' },
      { name: 'accountSubType' },
      { name: 'status' },
    ],
  },
];

interface Outcome {
  name: string;
  value: string;
  badge: string;
  tone: Tone;
  note: string;
}

const OUTCOMES: Array<Outcome> = [
  {
    name: 'name',
    value: '"Checking"',
    badge: 'untouched',
    tone: 'untouched',
    note: 'the key was never visited',
  },
  {
    name: 'mask',
    value: 'null',
    badge: 'cleared',
    tone: 'cleared',
    note: 'nullable, so an explicit null unsets it',
  },
  {
    name: 'currentBalance',
    value: '98000',
    badge: 'updated',
    tone: 'updated',
    note: 'the only field the client meant to change',
  },
  {
    name: 'limitBalance',
    value: 'null',
    badge: 'untouched',
    tone: 'untouched',
    note: 'absent is not the same as null',
  },
  {
    name: 'accountSubType',
    value: '"checking"',
    badge: 'untouched',
    tone: 'untouched',
    note: 'writable, just not written to',
  },
  {
    name: 'status',
    value: '"active"',
    badge: 'unreachable',
    tone: 'rejected',
    note: 'sending it at all is a 400',
  },
];

export default function MergeDiagram(): React.JSX.Element {
  return (
    <figure className={styles.root}>
      <div className={styles.sheets}>
        {SHEETS.map(sheet => (
          <section className={styles.sheet} data-kind={sheet.kind} key={sheet.title}>
            <h4 className={styles.title}>{sheet.title}</h4>
            <p className={styles.subtitle}>{sheet.subtitle}</p>
            <dl className={styles.fields}>
              {sheet.fields.map(field => (
                <div className={styles.field} key={field.name}>
                  <dt className={styles.fieldName}>{field.name}</dt>
                  {field.value ? (
                    <dd className={styles.fieldValue} data-tone={field.tone}>{field.value}</dd>
                  ) : (
                    <dd className={styles.absent}>key absent</dd>
                  )}
                </div>
              ))}
            </dl>
          </section>
        ))}
      </div>

      <div aria-hidden='true' className={styles.chevrons}>
        <span />
        <span />
        <span />
      </div>

      <p className={styles.band}>merge, key by key, right into left</p>

      <section className={styles.result}>
        <h4 className={styles.title}>What actually gets persisted</h4>
        <p className={styles.subtitle}>what you see looking straight down through the stack</p>
        <table className={styles.outcomes}>
          <tbody>
            {OUTCOMES.map(outcome => (
              <tr data-tone={outcome.tone} key={outcome.name}>
                <th className={styles.outcomeName} scope='row' title={outcome.name}>{outcome.name}</th>
                <td className={styles.outcomeValue} title={outcome.value}>{outcome.value}</td>
                <td className={styles.outcomeBadgeCell}>
                  <span className={styles.badge}>{outcome.badge}</span>
                </td>
                <td className={styles.outcomeNote}>{outcome.note}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </figure>
  );
}
