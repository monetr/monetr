(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[857],{4194:(e,s,i)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/cors",function(){return i(3706)}])},3706:(e,s,i)=>{"use strict";i.r(s),i.d(s,{default:()=>h,useTOC:()=>l});var n=i(2540),r=i(7933),t=i(907),d=i(8439);function l(e){return[]}let h=(0,r.e)(function(e){let s={code:"code",h1:"h1",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,d.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.h1,{children:"CORS (Cross Origin Resource Sharing)"}),"\n",(0,n.jsx)(s.p,{children:"monetr generally is hosted on a single domain name and thus does not require CORS, however if your self hosting setup\nrequires that your monetr instance be accessible from another domain name then you must configure CORS."}),"\n",(0,n.jsx)(s.p,{children:"Below is an example of the CORS configuration block:"}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(s.code,{children:[(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"cors"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # allowedOrigins determines the value of the `Access-Control-Allow-Origin` response header. In monetr this defaults to"})}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # an empty list. This default forbids all cross origin access."})}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  allowedOrigins"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"    - "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"https://your.monetr.local"})]}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # Enable debug logging to help diagnose CORS issues."})}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  debug"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"})]})]})}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Name"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Type"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Default"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Description"})})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"allowedOrigins"})}),(0,n.jsx)(s.td,{children:"Array"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"[]"})}),(0,n.jsx)(s.td,{children:"Other origins that are allowed to access your monetr server."})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"debug"})}),(0,n.jsx)(s.td,{children:"Boolean"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"false"})}),(0,n.jsx)(s.td,{children:"Debug logging for helping diagnose CORS issues."})]})]})]}),"\n",(0,n.jsx)(s.p,{children:"The following environment variables can be used to configure CORS options:"}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:"Variable"}),(0,n.jsx)(s.th,{children:"Config File Field"})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_CORS_ALLOWED_ORIGINS"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"cors.allowedOrigins"})})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_CORS_DEBUG"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"cors.debug"})})]})]})]})]})},"/documentation/configure/cors",{filePath:"src/pages/documentation/configure/cors.mdx",timestamp:1732468966e3,pageMap:t.O,frontMatter:{title:"CORS"},title:"CORS"},"undefined"==typeof RemoteContent?l:RemoteContent.useTOC)}},e=>{var s=s=>e(e.s=s);e.O(0,[2615,636,6593,8792],()=>s(4194)),_N_E=e.O()}]);