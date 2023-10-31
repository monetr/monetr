(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[6730],{8039:function(e,n,t){(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/development/documentation",function(){return t(647)}])},647:function(e,n,t){"use strict";t.r(n),t.d(n,{__toc:function(){return l}});var s=t(4246),i=t(9304),a=t(1441),o=t(2961);let l=[{depth:2,value:"Editing documentation",id:"editing-documentation"},{depth:3,value:"Building Locally",id:"building-locally"},{depth:3,value:"Editing Locally",id:"editing-locally"},{depth:2,value:"Style",id:"style"},{depth:3,value:"Language",id:"language"},{depth:3,value:"Reader / Author",id:"reader--author"},{depth:3,value:"Code Blocks",id:"code-blocks"},{depth:3,value:"Screenshots",id:"screenshots"},{depth:3,value:"Issue Tracking",id:"issue-tracking"},{depth:3,value:"Inclusivity",id:"inclusivity"}];function _createMdxContent(e){let n=Object.assign({h1:"h1",p:"p",ul:"ul",li:"li",a:"a",img:"img",h2:"h2",code:"code",h3:"h3",pre:"pre",span:"span"},(0,a.a)(),e.components);return o.mQ||_missingMdxReference("Tabs",!1),o.mQ.Tab||_missingMdxReference("Tabs.Tab",!0),(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(n.h1,{children:"Documentation"}),"\n",(0,s.jsx)(n.p,{children:"This is an overview of ways to contribute to monetr's documentation. To get started:"}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsxs)(n.li,{children:["You can find outstanding issues for documentation\nhere: ",(0,s.jsx)(n.a,{href:"https://github.com/monetr/monetr/issues?q=is%3Aopen+is%3Aissue+label%3Adocumentation",children:(0,s.jsx)(n.img,{src:"https://img.shields.io/github/issues/monetr/monetr/documentation",alt:"GitHub issues by-label"})})]}),"\n",(0,s.jsx)(n.li,{children:"If you don't find an issue that you'd be interested in working on, you can still create a pull request with your\ndesired changes."}),"\n",(0,s.jsx)(n.li,{children:"If you have found a gap in our documentation that you aren't able to, or do not wish to fill yourself; please create\nan issue so that others are aware of this gap, and it can be addressed."}),"\n"]}),"\n",(0,s.jsx)(n.h2,{id:"editing-documentation",children:"Editing documentation"}),"\n",(0,s.jsxs)(n.p,{children:["All of our documentation is in the form of Markdown files in the ",(0,s.jsx)(n.code,{children:"docs"})," directory of the monetr repository. You can\nsimply edit the existing files to make changes to the documentation. The documentation site is automatically generated\nin our GitHub Actions workflows."]}),"\n",(0,s.jsx)(n.h3,{id:"building-locally",children:"Building Locally"}),"\n",(0,s.jsxs)(n.p,{children:["You can build our documentation site locally using the following command, it requires Node.js and will create a static\nsite in the ",(0,s.jsx)(n.code,{children:"$PWD/docs/out"})," directory."]}),"\n",(0,s.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,s.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,s.jsxs)(n.span,{className:"line",children:[(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"docs"})]})})}),"\n",(0,s.jsx)(n.h3,{id:"editing-locally",children:"Editing Locally"}),"\n",(0,s.jsx)(n.p,{children:"If you want to work on the documentation in real time locally you can run the following command:"}),"\n",(0,s.jsxs)(o.mQ,{items:["GNUMake","pnpm"],children:[(0,s.jsx)(o.mQ.Tab,{children:(0,s.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,s.jsxs)(n.code,{"data-language":"shell","data-theme":"default",children:[(0,s.jsx)(n.span,{className:"line",children:(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# Will install any dependencies and prepare the local development environment for docs"})}),"\n",(0,s.jsxs)(n.span,{className:"line",children:[(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"develop-docs"})]})]})})}),(0,s.jsx)(o.mQ.Tab,{children:(0,s.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,s.jsxs)(n.code,{"data-language":"shell","data-theme":"default",children:[(0,s.jsx)(n.span,{className:"line",children:(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# Install dependencies manually"})}),"\n",(0,s.jsxs)(n.span,{className:"line",children:[(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"pnpm"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"install"})]}),"\n",(0,s.jsx)(n.span,{className:"line",children:" "}),"\n",(0,s.jsx)(n.span,{className:"line",children:(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# Start the documentation server"})}),"\n",(0,s.jsxs)(n.span,{className:"line",children:[(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"pnpm"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"-r"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"-filter"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"docs"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"run"}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,s.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"dev"})]})]})})})]}),"\n",(0,s.jsxs)(n.p,{children:["This will start a Next.js server locally serving docs. As you edit files in the ",(0,s.jsx)(n.code,{children:"docs/"})," directory you will be able to\nsee those changes automatically refresh in your browser."]}),"\n",(0,s.jsx)(n.h2,{id:"style",children:"Style"}),"\n",(0,s.jsx)(n.p,{children:"We would like our documentation to follow a general guide, this creates some consistency in how our documentation is\nboth written, presented, and maintained over time."}),"\n",(0,s.jsx)(n.h3,{id:"language",children:"Language"}),"\n",(0,s.jsx)(n.p,{children:'All documentation should be written in "American" English as much as possible. The exception to that rule are\nquotations, trademarks or terms that are better known by their own language\'s equivalent.'}),"\n",(0,s.jsx)(n.h3,{id:"reader--author",children:"Reader / Author"}),"\n",(0,s.jsx)(n.p,{children:'The documentation prefers "we" to address the author and "you" to address the reader. The gender of the reader shall be\nneutral if possible. Attempt to use "they" as a pronoun for the reader.'}),"\n",(0,s.jsx)(n.h3,{id:"code-blocks",children:"Code Blocks"}),"\n",(0,s.jsx)(n.p,{children:"Code blocks should always be accompanied by a preceding text to give context as to what that code block is, or\nrepresents. Adjacent code blocks without a paragraph of text between them should be avoided."}),"\n",(0,s.jsx)(n.h3,{id:"screenshots",children:"Screenshots"}),"\n",(0,s.jsxs)(n.p,{children:["Screenshots, if at all possible, should be no larger than ",(0,s.jsx)(n.code,{children:"1280x720"}),". This is not a strict requirement, but if a\nscreenshot can reasonably capture all the necessary details in that resolution or less; that is greatly preferred."]}),"\n",(0,s.jsx)(n.h3,{id:"issue-tracking",children:"Issue Tracking"}),"\n",(0,s.jsxs)(n.p,{children:["If documentation is missing and is planned to be added later. Please add a placeholder badge for that documentation\nusing ",(0,s.jsx)(n.a,{href:"https://shields.io/category/issue-tracking",children:"shields.io"}),", with the ",(0,s.jsx)(n.code,{children:"GitHub issue/pull request detail"})," shield."]}),"\n",(0,s.jsx)(n.p,{children:'The "override label" should use the following format.'}),"\n",(0,s.jsx)(n.pre,{"data-language":"text","data-theme":"default",filename:"Label Format",children:(0,s.jsx)(n.code,{"data-language":"text","data-theme":"default",children:(0,s.jsx)(n.span,{className:"line",children:(0,s.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"#{GitHub Issue Number} - {GitHub Issue Title}"})})})}),"\n",(0,s.jsx)(n.p,{children:"Please make the badge link back to the original issue as well."}),"\n",(0,s.jsx)(n.h3,{id:"inclusivity",children:"Inclusivity"}),"\n",(0,s.jsx)(n.p,{children:"Language that has been identified as hurtful or insensitive should be avoided."})]})}function MDXContent(){let e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{},{wrapper:n}=Object.assign({},(0,a.a)(),e.components);return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(_createMdxContent,{...e})}):_createMdxContent(e)}function _missingMdxReference(e,n){throw Error("Expected "+(n?"component":"object")+" `"+e+"` to be defined: you likely forgot to import, pass, or provide it.")}n.default=(0,i.j)({MDXContent,pageOpts:{filePath:"src/pages/documentation/development/documentation.mdx",route:"/documentation/development/documentation",timestamp:1698719236e3,title:"Documentation",headings:l},pageNextRoute:"/documentation/development/documentation"})}},function(e){e.O(0,[9304,9774,2888,179],function(){return e(e.s=8039)}),_N_E=e.O()}]);