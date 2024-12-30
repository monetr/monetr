import Image from 'next/image';
import Link from 'next/link';
// import type { Page } from 'nextra';

interface BlogListItemProps {
  // page: Page;
}

export default function BlogListItem(props: BlogListItemProps): JSX.Element {
  const page = {
    name: '2024-12-30-introduction',
    route: '/blog/2024-12-30-introduction',
    frontMatter: {
      title: 'Introducing monetr test',
      date: '2024/12/30',
      description: 'Announcing monetr\'s launch, why it was built, and how it works.',
      tag: 'Announcement',
      ogImage: '/blog/2024-12-30-introduction/preview.png',
      author: 'Elliot Courant',
    },
  };
  // const page = props.page as Page & { frontMatter: FrontMatter };
  // console.log(page);
  // https://api.github.com/users/elliotcourant
  //
  //
  return (
    <Link key={ page.route } href={ page.route } className='block mb-8 group'>
      { page.frontMatter?.ogImage ? (
        <div className='mt-4 rounded relative aspect-video overflow-hidden'>
          <Image
            src={ page.frontMatter.ogImage }
            className='object-cover transform group-hover:scale-105 transition-transform'
            alt={ page.frontMatter?.title ?? 'Blog post image' }
            fill={ true }
            sizes='(min-width: 1024px) 33vw, 100vw'
            priority
          />
        </div>
      ) : null }
      <h2 className='block mt-8 text-4xl opacity-90 group-hover:opacity-100'>
        { page.meta?.title || page.frontMatter?.title || page.name }
      </h2>
      <div className='opacity-80 mt-2 group-hover:opacity-100'>
        { page.frontMatter?.description } <span>Read more →</span>
      </div>
      <div className='flex gap-2 flex-wrap mt-3 items-baseline'>
        { page.frontMatter?.tag ? (
          <span className='opacity-80 text-xs py-1 px-2 ring-1 ring-gray-300 rounded group-hover:opacity-100'>
            { page.frontMatter.tag }
          </span>
        ) : null }
        { page.frontMatter?.date ? (
          <span className='opacity-60 text-sm group-hover:opacity-100'>
            { page.frontMatter.date }
          </span>
        ) : null }
        { page.frontMatter?.author ? (
          <span className='opacity-60 text-sm group-hover:opacity-100'>
            by { page.frontMatter.author }
          </span>
        ) : null }
      </div>
    </Link>
  );
}
