(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[9443],{2512:(e,r,s)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/redis",function(){return s(728)}])},728:(e,r,s)=>{"use strict";s.r(r),s.d(r,{default:()=>h,useTOC:()=>a});var i=s(2540),d=s(7933),n=s(7170),t=s(8795),l=s(1785);function a(e){return[]}let h=(0,d.e)(function(e){let r={a:"a",code:"code",h1:"h1",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,t.R)(),...e.components};return(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(r.h1,{children:"Redis Configuration"}),"\n",(0,i.jsxs)(r.p,{children:["monetr can use any cache that is compatible with Redis’s wire protocol. In the provided Docker compose file, local\ndevelopment environment as well as in production ",(0,i.jsx)(r.a,{href:"https://github.com/valkey-io/valkey",children:"valkey"})," is used."]}),"\n",(0,i.jsxs)(r.p,{children:["monetr only caches a few things, and for self-hosting it may not even be necessary to run a dedicated cache at this time\nas monetr also leverages ",(0,i.jsx)(r.a,{href:"https://github.com/alicebob/miniredis",children:"miniredis"})," when a cache server has not been configured.\nFor a single monetr server this embedded “Redis” is sufficient."]}),"\n",(0,i.jsx)(r.p,{children:"To configure a dedicated cache server though:"}),"\n",(0,i.jsx)(r.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,i.jsxs)(r.code,{children:[(0,i.jsxs)(r.span,{children:[(0,i.jsx)(r.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"redis"}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,i.jsxs)(r.span,{children:[(0,i.jsx)(r.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  enabled"}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"<true|false>"}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # If this is set to false then miniredis is used."})]}),"\n",(0,i.jsxs)(r.span,{children:[(0,i.jsx)(r.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  address"}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"localhost"'})]}),"\n",(0,i.jsxs)(r.span,{children:[(0,i.jsx)(r.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  port"}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,i.jsx)(r.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"6379"})]})]})}),"\n",(0,i.jsxs)(r.table,{children:[(0,i.jsx)(r.thead,{children:(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.th,{children:(0,i.jsx)(r.strong,{children:"Name"})}),(0,i.jsx)(r.th,{children:(0,i.jsx)(r.strong,{children:"Type"})}),(0,i.jsx)(r.th,{children:(0,i.jsx)(r.strong,{children:"Default"})}),(0,i.jsx)(r.th,{children:(0,i.jsx)(r.strong,{children:"Description"})})]})}),(0,i.jsxs)(r.tbody,{children:[(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"enabled"})}),(0,i.jsx)(r.td,{children:"Boolean"}),(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"false"})}),(0,i.jsxs)(r.td,{children:["Enable a dedicated cache server, if this is set to ",(0,i.jsx)(r.code,{children:"false"})," then an embedded miniredis instance is used instead."]})]}),(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"address"})}),(0,i.jsx)(r.td,{children:"String"}),(0,i.jsx)(r.td,{}),(0,i.jsx)(r.td,{children:"The IP, or DNS resolvable address of your Redis-compatible cache server."})]}),(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"port"})}),(0,i.jsx)(r.td,{children:"Number"}),(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"6379"})}),(0,i.jsx)(r.td,{children:"Port that the Redis-compatible cache server can be reached at."})]})]})]}),"\n",(0,i.jsx)(l.P,{type:"info",children:(0,i.jsx)(r.p,{children:"monetr does not support credentials or TLS for this cache server at this time. Sensitive information is never cached\non this server and the use of the cache is purely for performance."})}),"\n",(0,i.jsx)(r.p,{children:"The following environment variables map to the following configuration file fields. Each field is documented below."}),"\n",(0,i.jsxs)(r.table,{children:[(0,i.jsx)(r.thead,{children:(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.th,{children:"Variable"}),(0,i.jsx)(r.th,{children:"Config File Field"})]})}),(0,i.jsxs)(r.tbody,{children:[(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"MONETR_REDIS_ENABLED"})}),(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"redis.enabled"})})]}),(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"MONETR_REDIS_ADDRESS"})}),(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"redis.address"})})]}),(0,i.jsxs)(r.tr,{children:[(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"MONETR_REDIS_PORT"})}),(0,i.jsx)(r.td,{children:(0,i.jsx)(r.code,{children:"redis.port"})})]})]})]})]})},"/documentation/configure/redis",{filePath:"src/pages/documentation/configure/redis.mdx",timestamp:173561954e4,pageMap:n.O,frontMatter:{title:"Redis"},title:"Redis"},"undefined"==typeof RemoteContent?a:RemoteContent.useTOC)},1785:(e,r,s)=>{"use strict";s.d(r,{P:()=>a});var i=s(2540),d=s(1750),n=s(6877);let t={default:"\uD83D\uDCA1",error:"\uD83D\uDEAB",info:(0,i.jsx)(n.KS,{className:"_mt-1"}),warning:"⚠️"},l={default:(0,d.A)("_border-orange-100 _bg-orange-50 _text-orange-800 dark:_border-orange-400/30 dark:_bg-orange-400/20 dark:_text-orange-300"),error:(0,d.A)("_border-red-200 _bg-red-100 _text-red-900 dark:_border-red-200/30 dark:_bg-red-900/30 dark:_text-red-200"),info:(0,d.A)("_border-blue-200 _bg-blue-100 _text-blue-900 dark:_border-blue-200/30 dark:_bg-blue-900/30 dark:_text-blue-200"),warning:(0,d.A)("_border-yellow-100 _bg-yellow-50 _text-yellow-900 dark:_border-yellow-200/30 dark:_bg-yellow-700/30 dark:_text-yellow-200")};function a({children:e,type:r="default",emoji:s=t[r]}){return(0,i.jsxs)("div",{className:(0,d.A)("nextra-callout _overflow-x-auto _mt-6 _flex _rounded-lg _border _py-2 ltr:_pr-4 rtl:_pl-4","contrast-more:_border-current contrast-more:dark:_border-current",l[r]),children:[(0,i.jsx)("div",{className:"_select-none _text-xl ltr:_pl-3 ltr:_pr-2 rtl:_pr-3 rtl:_pl-2",style:{fontFamily:'"Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"'},children:s}),(0,i.jsx)("div",{className:"_w-full _min-w-0 _leading-7",children:e})]})}}},e=>{var r=r=>e(e.s=r);e.O(0,[7933,7170,636,6593,8792],()=>r(2512)),_N_E=e.O()}]);