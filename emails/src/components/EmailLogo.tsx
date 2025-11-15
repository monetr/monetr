import { Img, Section } from '@react-email/components';

export interface EmailLogoProps {
  baseUrl: string;
}

export default function EmailLogo(props: EmailLogoProps): JSX.Element {
  return (
    <Section className='mt-8 border-0'>
      <Img
        alt='monetr'
        className='my-0 mx-auto'
        height='64'
        src={`${props.baseUrl}/assets/resources/transparent-128.png `}
        width='64'
      />
    </Section>
  );
}
