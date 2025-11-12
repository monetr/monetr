interface GithubIssueProps {
  prefix: string;
  issueNumber: number;
}

export default function GithubIssue(props: GithubIssueProps): JSX.Element {
  return (
    <a
      className='nx-text-primary-600 nx-underline nx-decoration-from-font [text-underline-position:from-font]'
      href={`https://github.com/monetr/monetr/issues/${props.issueNumber}`}
      rel='noreferrer'
      target='_blank'
    >
      <img
        alt='GitHub issue/pull request detail'
        src={`https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fmonetr%2Fmonetr%2Fissues%2F${props.issueNumber}&query=%24.title&logo=github&label=${props.prefix}`}
      />
      <span className='nx-sr-only nx-select-none'>(opens in a new tab)</span>
    </a>
  );
}
