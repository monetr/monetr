(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[2106],{1012:(e,n,r)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/install",function(){return r(9732)}])},9732:(e,n,r)=>{"use strict";r.r(n),r.d(n,{default:()=>p,useTOC:()=>c});var t=r(2540),o=r(7933),s=r(7170),a=r(8795),i=r(3670),l=r(5230);let d=(0,l.A)("Container",[["path",{d:"M22 7.7c0-.6-.4-1.2-.8-1.5l-6.3-3.9a1.72 1.72 0 0 0-1.7 0l-10.3 6c-.5.2-.9.8-.9 1.4v6.6c0 .5.4 1.2.8 1.5l6.3 3.9a1.72 1.72 0 0 0 1.7 0l10.3-6c.5-.3.9-1 .9-1.5Z",key:"1t2lqe"}],["path",{d:"M10 21.9V14L2.1 9.1",key:"o7czzq"}],["path",{d:"m10 14 11.9-6.9",key:"zm5e20"}],["path",{d:"M14 19.8v-8.1",key:"159ecu"}],["path",{d:"M18 17.5V9.4",key:"11uown"}]]),h=(0,l.A)("ShipWheel",[["circle",{cx:"12",cy:"12",r:"8",key:"46899m"}],["path",{d:"M12 2v7.5",key:"1e5rl5"}],["path",{d:"m19 5-5.23 5.23",key:"1ezxxf"}],["path",{d:"M22 12h-7.5",key:"le1719"}],["path",{d:"m19 19-5.23-5.23",key:"p3fmgn"}],["path",{d:"M12 14.5V22",key:"dgcmos"}],["path",{d:"M10.23 13.77 5 19",key:"qwopd4"}],["path",{d:"M9.5 12H2",key:"r7bup8"}],["path",{d:"M10.23 10.23 5 5",key:"k2y7lj"}],["circle",{cx:"12",cy:"12",r:"2.5",key:"ix0uyj"}]]);function c(e){let n={strong:"strong",...(0,a.R)()};return[{value:"Why Self-Host monetr?",id:"why-self-host-monetr",depth:2},{value:"Requirements for Self-Hosting",id:"requirements-for-self-hosting",depth:2},{value:"Installation Options",id:"installation-options",depth:2},{value:(0,t.jsx)(t.Fragment,{children:(0,t.jsx)(n.strong,{children:"Install via Docker Compose"})}),id:"install-via-docker-compose",depth:3},{value:(0,t.jsx)(t.Fragment,{children:(0,t.jsx)(n.strong,{children:"Install on Kubernetes"})}),id:"install-on-kubernetes",depth:3},{value:"What’s Next?",id:"whats-next",depth:2}]}function u(e,n){throw Error("Expected "+(n?"component":"object")+" `"+e+"` to be defined: you likely forgot to import, pass, or provide it.")}let p=(0,o.e)(function(e){let{toc:n=c(e)}=e,r={a:"a",br:"br",em:"em",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return i.C||u("Cards",!1),i.C.Card||u("Cards.Card",!0),(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(r.h1,{children:"Install monetr"}),"\n",(0,t.jsx)(r.p,{children:"monetr is completely free to self-host, giving you full control over your data and setup. This guide provides an\noverview of how to install and run monetr on your own infrastructure."}),"\n",(0,t.jsx)(r.h2,{id:n[0].id,children:n[0].value}),"\n",(0,t.jsx)(r.p,{children:"Self-hosting monetr offers several advantages:"}),"\n",(0,t.jsxs)(r.ul,{children:["\n",(0,t.jsxs)(r.li,{children:[(0,t.jsx)(r.strong,{children:"Cost-Free Usage"}),": Run monetr without paying subscription fees.",(0,t.jsx)(r.br,{}),"\n",(0,t.jsx)("small",{children:(0,t.jsxs)(r.em,{children:["Note: If you use Plaid as a data provider, you may incur additional costs. See ",(0,t.jsx)(r.a,{href:"https://plaid.com/pricing/",children:"Plaid’s\nPricing"})," for details."]})})]}),"\n",(0,t.jsxs)(r.li,{children:[(0,t.jsx)(r.strong,{children:"Full Data Ownership"}),": Keep complete control over your financial data."]}),"\n"]}),"\n",(0,t.jsx)(r.p,{children:"Whether you’re interested in hosting your own applications or prefer a private alternative to hosted services,\nself-hosting gives you more control over your data and setup."}),"\n",(0,t.jsx)(r.h2,{id:n[1].id,children:n[1].value}),"\n",(0,t.jsx)(r.p,{children:"Before you begin, ensure you have the following:"}),"\n",(0,t.jsxs)(r.ul,{children:["\n",(0,t.jsxs)(r.li,{children:[(0,t.jsx)(r.strong,{children:"Container Runtime"}),": Docker Compose is the officially supported method for running monetr."]}),"\n",(0,t.jsxs)(r.li,{children:[(0,t.jsx)(r.strong,{children:"Server or Host Machine"}),":","\n",(0,t.jsxs)(r.ul,{children:["\n",(0,t.jsx)(r.li,{children:"Recommended: A system with at least 512MB of RAM and 1GB of disk space."}),"\n"]}),"\n"]}),"\n",(0,t.jsxs)(r.li,{children:[(0,t.jsx)(r.strong,{children:"Database"}),": monetr runs with PostgreSQL, which is included in the Docker Compose setup."]}),"\n",(0,t.jsxs)(r.li,{children:[(0,t.jsx)(r.strong,{children:"Domain and SSL (Optional)"}),": For public access, consider setting up a domain with HTTPS."]}),"\n"]}),"\n",(0,t.jsx)(r.h2,{id:n[2].id,children:n[2].value}),"\n",(0,t.jsx)(r.p,{children:"Currently, monetr only officially supports container-based installations:"}),"\n",(0,t.jsx)(r.h3,{id:n[3].id,children:n[3].value}),"\n",(0,t.jsx)(r.p,{children:"The easiest way to self-host monetr is by using Docker Compose. This setup includes all necessary services, such as the\nmonetr application and PostgreSQL."}),"\n",(0,t.jsx)(i.C.Card,{icon:(0,t.jsx)(d,{}),title:"Follow the Docker Compose Installation Guide",description:"Install monetr on your own system using Docker Compose",href:"/documentation/install/docker/"}),"\n",(0,t.jsx)(r.h3,{id:n[4].id,children:n[4].value}),"\n",(0,t.jsxs)(r.p,{children:["If you are looking to deploy monetr on a Kubernetes cluster, there is a ",(0,t.jsx)(r.em,{children:"very"})," basic guide on how to do so. This is not\nthe recommended way to run monetr at this time. A helm chart will be available in the future, but this guide has been\nmade available until then."]}),"\n",(0,t.jsx)(i.C.Card,{icon:(0,t.jsx)(h,{}),title:"Follow the Kubernetes Installation Guide",description:"Install monetr on your own Kubernetes cluster",href:"/documentation/install/kubernetes/"}),"\n",(0,t.jsx)(r.h2,{id:n[5].id,children:n[5].value}),"\n",(0,t.jsx)(r.p,{children:"After installing monetr, you can:"}),"\n",(0,t.jsxs)(r.ul,{children:["\n",(0,t.jsxs)(r.li,{children:["Configure it for your specific needs. See ",(0,t.jsx)(r.a,{href:"./configure",children:"Configuration"}),"."]}),"\n",(0,t.jsx)(r.li,{children:"Explore advanced hosting options like reverse proxies or custom domains."}),"\n",(0,t.jsx)(r.li,{children:"Explore available resources for feedback and support, including documentation updates and support channels."}),"\n"]}),"\n",(0,t.jsxs)(r.p,{children:["For detailed steps, visit the full ",(0,t.jsx)(r.a,{href:"./install/docker",children:"Docker Compose Installation Guide"}),"."]})]})},"/documentation/install",{filePath:"src/pages/documentation/install.mdx",timestamp:1740527393e3,pageMap:s.O,frontMatter:{title:"Self-Hosted Installation",description:"Learn how to self-host monetr for free using Docker or Podman. Explore the benefits of self-hosting and get an overview of installation requirements and options."},title:"Self-Hosted Installation"},"undefined"==typeof RemoteContent?c:RemoteContent.useTOC)},3670:(e,n,r)=>{"use strict";r.d(n,{C:()=>i});var t=r(2540),o=r(1750),s=r(5270),a=r.n(s);let i=Object.assign(function({children:e,num:n=3,className:r,style:s,...a}){return(0,t.jsx)("div",{className:(0,o.A)("nextra-cards _mt-4 _gap-4 _grid","_not-prose",r),...a,style:{...s,"--rows":n},children:e})},{displayName:"Cards",Card:function({children:e,title:n,icon:r,arrow:s,href:i,...l}){return(0,t.jsxs)(a(),{href:i,className:(0,o.A)("nextra-focus nextra-card _group _flex _flex-col _justify-start _overflow-hidden _rounded-lg _border _border-gray-200","_text-current _no-underline dark:_shadow-none","hover:_shadow-gray-100 dark:hover:_shadow-none _shadow-gray-100","active:_shadow-sm active:_shadow-gray-200","_transition-all _duration-200 hover:_border-gray-300",e?"_bg-gray-100 _shadow dark:_border-neutral-700 dark:_bg-neutral-800 dark:_text-gray-50 hover:_shadow-lg dark:hover:_border-neutral-500 dark:hover:_bg-neutral-700":"_bg-transparent _shadow-sm dark:_border-neutral-800 hover:_bg-slate-50 hover:_shadow-md dark:hover:_border-neutral-700 dark:hover:_bg-neutral-900"),...l,children:[e,(0,t.jsxs)("span",{className:(0,o.A)("_flex _font-semibold _items-center _gap-2 _p-4 _text-gray-700 hover:_text-gray-900",s&&'after:_content-["→"] after:_transition-transform after:_duration-75 after:group-hover:_translate-x-0.5',e?"dark:_text-gray-300 dark:hover:_text-gray-100":"dark:_text-neutral-200 dark:hover:_text-neutral-50"),title:n,children:[r,(0,t.jsx)("span",{className:"_truncate",children:n})]})]})}})}},e=>{var n=n=>e(e.s=n);e.O(0,[7933,7170,636,6593,8792],()=>n(1012)),_N_E=e.O()}]);