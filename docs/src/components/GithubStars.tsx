import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';

interface GithubStarsProps {
  variant?: 'default' | 'large';
}

interface GithubRepoResponse {
  id: number;
  name: string;
  description: string;
  stargazers_count: number;
}

export default function GithubStars(props: GithubStarsProps): JSX.Element {
  const { data, isLoading } = useQuery<GithubRepoResponse>({
    queryKey: ['https://api.github.com/repos/monetr/monetr'],
  });

  const stars = typeof data?.stargazers_count === 'number' ? data.stargazers_count.toLocaleString('en-US', {
    compactDisplay: 'short',
    notation: 'compact',
  }) : undefined;

  if (!props.variant || props.variant === 'default') {
    return (
      <div className='group h-[33.5px] flex shrink-0 flex-row items-center rounded-lg border border-dark-monetr-border overflow-hidden transition-opacity'>
        <div className='py-1 px-1 bg-zinc-800'>
          <svg
            fill='#FFFFFF'
            role='img'
            viewBox='0 0 24 24'
            xmlns='http://www.w3.org/2000/svg'
            className='group-hover:opacity-80 opacity-100 h-6 w-6'
          >
            <title>GitHub</title>
            <path d='M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12' />
          </svg>
        </div>
        <div className='py-1 text-center font-medium text-sm group-hover:opacity-80 opacity-100 w-10'>
          { isLoading || !stars ?
            <span className='rounded bg-dark-monetr-background-emphasis text-dark-monetr-background-emphasis'>???</span> :
            <span>
              { stars }
            </span>
          }
        </div>
      </div>
    );
  }

  return (
    <Link
      href='https://github.com/monetr/monetr'
      target='_blank'
      className='group h-[54px] flex shrink-0 flex-row items-center rounded-lg border border-dark-monetr-border overflow-hidden transition-opacity'
    >
      <div className='py-1 px-3 bg-zinc-800 h-full flex gap-2 items-center'>
        <svg
          fill='#FFFFFF'
          role='img'
          viewBox='0 0 24 24'
          xmlns='http://www.w3.org/2000/svg'
          className='group-hover:opacity-80 opacity-100 h-10 w-10'
        >
          <title>GitHub</title>
          <path d='M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12' />
        </svg>
        <span className='text-dark-monetr-content text-xl font-semibold group-hover:opacity-80 opacity-100'>
          GitHub
        </span>
      </div>
      <div className='py-1 px-3 text-center font-medium text-xl group-hover:opacity-80 opacity-100 bg-black bg-opacity-20 backdrop-blur-sm h-full items-center flex justify-center'>
        { isLoading || !stars ?
          <span className='rounded bg-dark-monetr-background-emphasis text-dark-monetr-background-emphasis'>???</span> :
          <span>
            { stars }
          </span>
        }
      </div>
    </Link>
  );

}
