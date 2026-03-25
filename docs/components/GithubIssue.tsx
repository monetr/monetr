import { Link } from '@rspress/core/theme-original';

interface GithubIssueProps {
  prefix: string;
  issueNumber: number;
}

export default function GithubIssue(props: GithubIssueProps): JSX.Element {
  return (
    <Link href={`https://github.com/monetr/monetr/issues/${props.issueNumber}`} rel='noreferrer' target='_blank'>
      <img
        alt='GitHub issue/pull request detail'
        src={`https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fmonetr%2Fmonetr%2Fissues%2F${props.issueNumber}&query=%24.title&logo=github&label=${props.prefix}`}
      />
    </Link>
  );
}
