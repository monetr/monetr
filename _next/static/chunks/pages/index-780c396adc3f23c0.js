(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[332],{2574:(e,t,r)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/",function(){return r(3555)}])},3555:(e,t,r)=>{"use strict";r.r(t),r.d(t,{__toc:()=>x,default:()=>p});var a=r(2540),n=r(1354),s=r(1369),l=r(8209),i=r.n(l),o=r(3696);let c=(e,t,r,a,n)=>{let s=(e-t)*(n-a)/(r-t)+a;return s>0?s:0};function u(e){let{className:t="",quantity:r=40,staticity:n=50,ease:s=50,refresh:l=!1}=e,i=(0,o.useRef)(null),u=(0,o.useRef)(null),d=(0,o.useRef)(null),h=(0,o.useRef)([]),A=function(){let[e,t]=(0,o.useState)({x:0,y:0});return(0,o.useEffect)(()=>{let e=e=>{t({x:e.clientX,y:e.clientY})};return window.addEventListener("mousemove",e),()=>{window.removeEventListener("mousemove",e)}},[]),e}(),x=(0,o.useRef)({x:0,y:0}),m=(0,o.useRef)({w:0,h:0}),p=window.devicePixelRatio,f=(0,o.useCallback)(()=>{d.current&&d.current.clearRect(0,0,m.current.w,m.current.h)},[d]),w=(0,o.useCallback)(()=>{u.current&&i.current&&d.current&&(h.current.length=0,m.current.w=u.current.offsetWidth,m.current.h=u.current.offsetHeight,i.current.width=m.current.w*p,i.current.height=m.current.h*p,i.current.style.width="".concat(m.current.w,"px"),i.current.style.height="".concat(m.current.h,"px"),d.current.scale(p,p))},[p]),g=(0,o.useCallback)(()=>{let e=Math.floor(Math.random()*m.current.w),t=Math.floor(Math.random()*m.current.h);return{x:e,y:t,translateX:0,translateY:0,size:Math.floor(2*Math.random())+1,alpha:0,targetAlpha:parseFloat((.6*Math.random()+.1).toFixed(1)),dx:(Math.random()-.5)*.2,dy:(Math.random()-.5)*.2,magnetism:.1+4*Math.random()}},[]),b=(0,o.useCallback)(function(e){let t=arguments.length>1&&void 0!==arguments[1]&&arguments[1];if(d.current){let{x:r,y:a,translateX:n,translateY:s,size:l,alpha:i}=e;d.current.translate(n,s),d.current.beginPath(),d.current.arc(r,a,l,0,2*Math.PI),d.current.fillStyle="rgba(255, 255, 255, ".concat(i,")"),d.current.fill(),d.current.setTransform(p,0,0,p,0,0),t||h.current.push(e)}},[p]),y=(0,o.useCallback)(()=>{f();for(let e=0;e<r;e++)b(g())},[g,f,b,r]),v=(0,o.useCallback)(()=>{w(),y()},[y,w]),j=(0,o.useCallback)(()=>{f(),h.current.forEach((e,t)=>{let r=parseFloat(c([e.x+e.translateX-e.size,m.current.w-e.x-e.translateX-e.size,e.y+e.translateY-e.size,m.current.h-e.y-e.translateY-e.size].reduce((e,t)=>Math.min(e,t)),0,20,0,1).toFixed(2));r>1?(e.alpha+=.02,e.alpha>e.targetAlpha&&(e.alpha=e.targetAlpha)):e.alpha=e.targetAlpha*r,e.x+=e.dx,e.y+=e.dy,e.translateX+=(x.current.x/(n/e.magnetism)-e.translateX)/s,e.translateY+=(x.current.y/(n/e.magnetism)-e.translateY)/s,e.x<-e.size||e.x>m.current.w+e.size||e.y<-e.size||e.y>m.current.h+e.size?(h.current.splice(t,1),b(g())):b({...e,x:e.x,y:e.y,translateX:e.translateX,translateY:e.translateY,alpha:e.alpha},!0)}),window.requestAnimationFrame(j)},[g,f,b,s,n]);(0,o.useEffect)(()=>(i.current&&(d.current=i.current.getContext("2d")),v(),j(),window.addEventListener("resize",v),()=>{window.removeEventListener("resize",v)}),[j,v]);let E=(0,o.useCallback)(()=>{if(i.current){let e=i.current.getBoundingClientRect(),{w:t,h:r}=m.current,a=A.x-e.left-t/2,n=A.y-e.top-r/2;a<t/2&&a>-t/2&&n<r/2&&n>-r/2&&(x.current.x=a,x.current.y=n)}},[A.x,A.y]);return(0,o.useEffect)(()=>{E()},[A.x,A.y,E]),(0,o.useEffect)(()=>{v()},[v,l]),(0,a.jsx)("div",{className:t,ref:u,"aria-hidden":"true",children:(0,a.jsx)("canvas",{ref:i})})}let d={src:"/_next/static/media/mobile_transactions_example.2ed271bc.png",height:2796,width:1290,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAICAMAAADp7a43AAAACVBMVEUcFyAlIyszMDlwCVVXAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAIElEQVR4nGNgYmBgYGBiZGBgYAQTIC6EBSZAMkxMDAwAAZoAFYhPUbAAAAAASUVORK5CYII=",blurWidth:4,blurHeight:8},h={src:"/_next/static/media/transactions_example.b971d11d.png",height:1440,width:2560,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAFCAMAAABPT11nAAAABlBMVEUcGSEsKDTps6McAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAFUlEQVR4nGNgYGRkAAMYjWAgAYgQAAEAAAiDvnwdAAAAAElFTkSuQmCC",blurWidth:8,blurHeight:5};function A(){return(0,a.jsxs)("div",{className:"w-full relative",children:[(0,a.jsx)("div",{className:"absolute inset-0 overflow-hidden pointer-events-none -z-10","aria-hidden":"true",children:(0,a.jsx)("div",{className:"absolute flex items-center justify-center top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 w-full sm:w-1/2 aspect-square",children:(0,a.jsx)("div",{className:"absolute inset-0 translate-z-0 bg-purple-500 rounded-full blur-[120px] opacity-50 min-h-[10vh]"})})}),(0,a.jsx)("div",{className:"max-md:hidden absolute bottom-0 -mb-20 left-2/3 -translate-x-1/2 blur-2xl opacity-70 pointer-events-none","aria-hidden":"true",children:(0,a.jsxs)("svg",{xmlns:"http://www.w3.org/2000/svg",width:"434",height:"427",children:[(0,a.jsx)("defs",{children:(0,a.jsxs)("linearGradient",{id:"bs5-a",x1:"19.609%",x2:"50%",y1:"14.544%",y2:"100%",children:[(0,a.jsx)("stop",{offset:"0%",stopColor:"#A855F7"}),(0,a.jsx)("stop",{offset:"100%",stopColor:"#6366F1",stopOpacity:"0"})]})}),(0,a.jsx)("path",{fill:"url(#bs5-a)",fillRule:"evenodd",d:"m661 736 461 369-284 58z",transform:"matrix(1 0 0 -1 -661 1163)"})]})}),(0,a.jsx)(u,{className:"absolute inset-0 -z-10"}),(0,a.jsxs)("div",{className:"m-view-height m-view-width flex flex-col py-16 mx-auto items-center gap-8",children:[(0,a.jsxs)("div",{className:"max-w-3xl flex flex-col gap-8 text-center items-center",children:[(0,a.jsxs)("div",{className:"flex items-center justify-center ml-3 p-4",children:[(0,a.jsx)("span",{className:"absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-none",children:"Coming Soon"}),(0,a.jsx)("h1",{className:"h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-auto",children:"Coming Soon"})]}),(0,a.jsx)("h1",{className:"text-4xl sm:text-5xl font-bold",children:"Always know what you can spend"}),(0,a.jsx)("h2",{className:"text-lg sm:text-xl font-medium",children:"Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and know what you have left-over to save or spend on whatever you'd like."})]}),(0,a.jsx)(i(),{src:h,alt:"Easily keep track of transactions",className:"hidden sm:block rounded-md z-10 shadow-lg"}),(0,a.jsx)(i(),{src:d,alt:"Easily keep track of transactions",className:"block sm:hidden rounded-md z-10 shadow-lg"})]})]})}let x=[];function m(e){return(0,a.jsx)(A,{})}let p=(0,n.n)({MDXContent:function(){let e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{},{wrapper:t}=Object.assign({},(0,s.R)(),e.components);return t?(0,a.jsx)(t,{...e,children:(0,a.jsx)(m,{...e})}):m(e)},pageOpts:{filePath:"src/pages/index.mdx",route:"/",frontMatter:{title:"monetr",description:"Always know what you can spend. Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and know what you have left-over to save or spend on whatever you'd like."},timestamp:173276846e4,title:"monetr",headings:x},pageNextRoute:"/"})}},e=>{var t=t=>e(e.s=t);e.O(0,[354,636,593,792],()=>t(2574)),_N_E=e.O()}]);