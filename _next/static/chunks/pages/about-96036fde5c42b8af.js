(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[239],{7396:(e,t,o)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/about",function(){return o(8612)}])},8612:(e,t,o)=>{"use strict";o.r(t),o.d(t,{default:()=>l,useTOC:()=>u});var n=o(2540),r=o(7933),i=o(931),a=o(753);function u(e){return[]}let l=(0,r.e)(function(e){return(0,n.jsx)("div",{className:"m-view-height w-full py-8",children:(0,n.jsx)(a.z,{})})},"/about",{filePath:"src/pages/about.mdx",timestamp:173275525e4,pageMap:i.O,frontMatter:{title:"About"},title:"About"},"undefined"==typeof RemoteContent?u:RemoteContent.useTOC)},753:(e,t,o)=>{"use strict";o.d(t,{z:()=>r});var n=o(2540);function r(){return(0,n.jsx)("div",{className:"w-full h-full flex items-center justify-center",children:(0,n.jsxs)("div",{className:"flex items-center justify-center ml-3 p-4",children:[(0,n.jsx)("span",{className:"absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl lg:text-6xl font-extrabold text-transparent text-center select-none",children:"Coming Soon"}),(0,n.jsx)("h1",{className:"h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl lg:text-6xl font-extrabold text-transparent text-center select-auto",children:"Coming Soon"})]})})}o(3696)},8439:(e,t,o)=>{"use strict";o.d(t,{R:()=>l});var n=o(3023),r=o(8209),i=o.n(r),a=o(3696);let u={img:e=>(0,a.createElement)("object"==typeof e.src?i():"img",e)},l=e=>(0,n.R)({...u,...e})},7933:(e,t,o)=>{"use strict";o.d(t,{e:()=>s});var n=o(2540),r=o(2922),i=o(8808);let a=(0,o(3696).createContext)({}),u=a.Provider;a.displayName="SSG";var l=o(8439);function s(e,t,o,n){let i=globalThis[r.VZ];return i.route=t,i.pageMap=o.pageMap,i.context[t]={Content:e,pageOpts:o,useTOC:n},c}function c({__nextra_pageMap:e=[],__nextra_dynamic_opts:t,...o}){let a=globalThis[r.VZ],{Layout:l,themeConfig:s}=a,{route:c,locale:m}=(0,i.r)(),p=a.context[c];if(!p)throw Error(`No content found for the "${c}" route. Please report it as a bug.`);let{pageOpts:f,useTOC:g,Content:h}=p;if(c.startsWith("/["))f.pageMap=e;else for(let{route:t,children:o}of e){let e=t.split("/").slice(m?2:1);(function e(t,[o,...n]){for(let r of t)if("children"in r&&o===r.name)return n.length?e(r.children,n):r})(f.pageMap,e).children=o}if(t){let{title:e,frontMatter:o}=t;f={...f,title:e,frontMatter:o}}return(0,n.jsx)(l,{themeConfig:s,pageOpts:f,pageProps:o,children:(0,n.jsx)(u,{value:o,children:(0,n.jsx)(d,{useTOC:g,children:(0,n.jsx)(h,{...o})})})})}function d({children:e,useTOC:t}){let{wrapper:o}=(0,l.R)();return(0,n.jsx)(m,{useTOC:t,wrapper:o,children:e})}function m({children:e,useTOC:t,wrapper:o,...r}){let i=t(r);return o?(0,n.jsx)(o,{toc:i,children:e}):e}},931:(e,t,o)=>{"use strict";o.d(t,{O:()=>n});let n=[{data:{index:{type:"page",title:"monetr",display:"hidden",theme:{layout:"raw"}},about:{type:"page",title:"About",theme:{layout:"raw"}},pricing:{type:"page",title:"Pricing",theme:{layout:"raw"}},blog:{type:"page",title:"Blog",theme:{layout:"raw"}},documentation:{type:"page",title:"Documentation"},contact:{type:"page",title:"Contact",display:"hidden"},policy:{type:"page",title:"Policies",display:"hidden"}}},{name:"about",route:"/about",frontMatter:{title:"About"}},{name:"blog",route:"/blog",frontMatter:{title:"Blog"}},{name:"contact",route:"/contact",frontMatter:{sidebarTitle:"Contact"}},{name:"documentation",route:"/documentation",children:[{data:{index:"Introduction","-- Help":{type:"separator",title:"Help"},use:"Using monetr","-- Installation":{type:"separator",title:"Installation"},install:"",configure:"","-- Contributing":{type:"separator",title:"Contributing"},development:""}},{name:"configure",route:"/documentation/configure",children:[{name:"captcha",route:"/documentation/configure/captcha",frontMatter:{title:"ReCAPTCHA"}},{name:"cors",route:"/documentation/configure/cors",frontMatter:{title:"CORS"}},{name:"email",route:"/documentation/configure/email",frontMatter:{sidebarTitle:"Email"}},{name:"kms",route:"/documentation/configure/kms",frontMatter:{title:"Key Management"}},{name:"links",route:"/documentation/configure/links",frontMatter:{sidebarTitle:"Links"}},{name:"logging",route:"/documentation/configure/logging",frontMatter:{sidebarTitle:"Logging"}},{name:"plaid",route:"/documentation/configure/plaid",frontMatter:{sidebarTitle:"Plaid"}},{name:"postgres",route:"/documentation/configure/postgres",frontMatter:{sidebarTitle:"Postgres"}},{name:"redis",route:"/documentation/configure/redis",frontMatter:{sidebarTitle:"Redis"}},{name:"security",route:"/documentation/configure/security",frontMatter:{sidebarTitle:"Security"}},{name:"sentry",route:"/documentation/configure/sentry",frontMatter:{sidebarTitle:"Sentry"}},{name:"server",route:"/documentation/configure/server",frontMatter:{sidebarTitle:"Server"}},{name:"storage",route:"/documentation/configure/storage",frontMatter:{sidebarTitle:"Storage"}}]},{name:"configure",route:"/documentation/configure",frontMatter:{title:"Configuration",description:"Configure self-hosted monetr servers"}},{name:"development",route:"/documentation/development",children:[{data:{documentation:"",code_of_conduct:"",build:"",local_development:"",credentials:""}},{name:"build",route:"/documentation/development/build",frontMatter:{sidebarTitle:"Build"}},{name:"code_of_conduct",route:"/documentation/development/code_of_conduct",frontMatter:{sidebarTitle:"Code of Conduct"}},{name:"credentials",route:"/documentation/development/credentials",frontMatter:{sidebarTitle:"Credentials"}},{name:"documentation",route:"/documentation/development/documentation",frontMatter:{sidebarTitle:"Documentation"}},{name:"local_development",route:"/documentation/development/local_development",frontMatter:{sidebarTitle:"Local Development"}}]},{name:"development",route:"/documentation/development",frontMatter:{title:"Contributing",description:"Guides on how to contribute to monetr, make changes to the application's code."}},{name:"index",route:"/documentation",frontMatter:{title:"Documentation",description:"Guides on how to use, self-host, or develop against monetr."}},{name:"install",route:"/documentation/install",children:[{name:"docker",route:"/documentation/install/docker",frontMatter:{title:"Self-Host via Docker",description:"Self-host monetr via Docker containers"}}]},{name:"install",route:"/documentation/install",frontMatter:{title:"Self-Host Installation",description:"Options on how to run monetr yourself for free."}},{name:"use",route:"/documentation/use",children:[{data:{starting_fresh:"Starting Fresh",funding_schedule:"Funding Schedules",expense:"Expenses",goal:"Goals",free_to_use:"Free-To-Use",security:"Security"}},{name:"expense",route:"/documentation/use/expense",frontMatter:{title:"Expenses",description:"Keep track of your regular or planned spending easily using expenses."}},{name:"free_to_use",route:"/documentation/use/free_to_use",frontMatter:{sidebarTitle:"Free to Use"}},{name:"funding_schedule",route:"/documentation/use/funding_schedule",frontMatter:{title:"Funding Schedules",description:"Contribute to your budgets on a regular basis, like every time you get paid."}},{name:"goal",route:"/documentation/use/goal",frontMatter:{sidebarTitle:"Goal"}},{name:"security",route:"/documentation/use/security",children:[{name:"user_password",route:"/documentation/use/security/user_password",frontMatter:{sidebarTitle:"User Password"}}]},{name:"starting_fresh",route:"/documentation/use/starting_fresh",frontMatter:{sidebarTitle:"Starting Fresh"}}]},{name:"use",route:"/documentation/use",frontMatter:{title:"Using monetr",description:"How to use and get the most out of monetr"}}]},{name:"index",route:"/",frontMatter:{title:"monetr",description:"Always know what you can spend. Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and know what you have left-over to save or spend on whatever you'd like."}},{name:"policy",route:"/policy",children:[{data:{terms:{title:"Terms & Conditions",theme:{sidebar:!1}},privacy:{title:"Privacy Policy",theme:{sidebar:!1}}}},{name:"privacy",route:"/policy/privacy",frontMatter:{sidebarTitle:"Privacy"}},{name:"terms",route:"/policy/terms",frontMatter:{sidebarTitle:"Terms"}}]},{name:"pricing",route:"/pricing",frontMatter:{title:"Pricing"}}]}},e=>{var t=t=>e(e.s=t);e.O(0,[636,593,792],()=>t(7396)),_N_E=e.O()}]);