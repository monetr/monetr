import logoData from './logo.svg';

interface LogoProps {
  className?: string;
}

export default function Logo(props: LogoProps): JSX.Element {
  return <img alt='monetr' className={props.className} src={logoData} />;
}
