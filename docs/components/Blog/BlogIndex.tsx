/*
  @license
  This code is adapted from https://github.com/langfuse/langfuse-docs

  MIT License

  Copyright (c) 2022 Finto Technologies

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.
  @license-end
*/

import { format, parse } from 'date-fns';
import { ArrowRight } from 'lucide-react';

import { usePages } from '@rspress/core/runtime';
import { Link } from '@rspress/core/theme-original';

export default function BlogIndex(): JSX.Element {
  const { pages } = usePages();

  const blogPages = pages
    .filter(page => page.routePath.startsWith('/blog/') && page.routePath !== '/blog/' && page.routePath !== '/blog')
    .sort((a, b) => {
      const dateA = a.frontmatter?.date ? new Date(a.frontmatter.date as string).getTime() : 0;
      const dateB = b.frontmatter?.date ? new Date(b.frontmatter.date as string).getTime() : 0;
      return dateB - dateA;
    });

  return (
    <div className='m-view-height py-8'>
      <div className='w-full flex flex-col gap-8 text-center items-center'>
        <div className='flex items-center justify-center p-4'>
          <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-5xl/tight font-extrabold text-transparent text-center select-none'>
            Blog
          </span>
          <h1 className='h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-5xl/tight font-extrabold text-transparent text-center select-auto'>
            Blog
          </h1>
        </div>
      </div>
      <div className='flex m-view-width mx-auto justify-center flex-wrap'>
        {blogPages.map(page => (
          <Link
            className='block mb-8 group flex-shrink-0 w-full lg:w-1/2 p-2 no-underline text-inherit'
            href={page.routePath}
            key={page.routePath}
          >
            {(page.frontmatter?.ogImage as string) ? (
              <div className='mt-4 rounded relative aspect-video overflow-hidden'>
                <img
                  alt={(page.frontmatter?.title as string) ?? 'Blog post image'}
                  className='object-cover w-full h-full transform group-hover:scale-105 transition-transform'
                  src={page.frontmatter.ogImage as string}
                />
              </div>
            ) : null}
            <h2 className='flex mt-8 text-3xl opacity-90 group-hover:opacity-100 items-center gap-2'>
              {(page.frontmatter?.title as string) || page.title}
              {(page.frontmatter?.tag as string) ? (
                <span className='opacity-80 text-xs py-1 px-2 ring-1 ring-gray-300 rounded group-hover:opacity-100 mt-1'>
                  {page.frontmatter.tag as string}
                </span>
              ) : null}
            </h2>
            <div className='opacity-80 mt-2 group-hover:opacity-100'>
              {page.frontmatter?.description as string}
              &nbsp;
              <span className='flex items-center'>
                Read more <ArrowRight className='h-4' />
              </span>
            </div>
            <div className='flex gap-1 flex-wrap mt-3 items-baseline'>
              {(page.frontmatter?.date as string) ? (
                <span className='opacity-60 text-sm group-hover:opacity-100'>
                  {format(parse(page.frontmatter.date as string, 'yyyy/MM/dd', new Date()), 'MMMM dd, yyyy')}
                </span>
              ) : null}
              {(page.frontmatter?.author as string) ? (
                <span className='opacity-60 text-sm group-hover:opacity-100'>
                  by {page.frontmatter.author as string}
                </span>
              ) : null}
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
