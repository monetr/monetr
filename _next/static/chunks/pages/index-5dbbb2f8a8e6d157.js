(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[5405],{7803:function(e,t,r){(window.__NEXT_P=window.__NEXT_P||[]).push(["/",function(){return r(8536)}])},8536:function(e,t,r){"use strict";r.r(t),r.d(t,{__toc:function(){return x},default:function(){return p}});var n=r(4246),a=r(9304),s=r(1441),l=r(8579),i=r.n(l),c=r(7378);function u(){let e=(0,c.useRef)(null),t=(0,c.useCallback)(()=>{var t,r,n,a;null===(t=e.current)||void 0===t||t.classList.remove("opacity-0"),null===(r=e.current)||void 0===r||r.classList.remove("scale-90"),null===(n=e.current)||void 0===n||n.classList.add("opacity-90"),null===(a=e.current)||void 0===a||a.classList.add("scale-100")},[]);return(0,c.useEffect)(()=>{if(!e.current)return;let r=e.current;return e.current.addEventListener("load",t),()=>{r.removeEventListener("load",t)}}),(0,n.jsx)("iframe",{ref:e,title:"monetr interface",loading:"lazy",className:"w-full h-full translate-x-0 translate-y-0 scale-90 delay-150 duration-500 ease-in-out rounded-2xl mt-8 shadow-2xl z-10 backdrop-blur-md bg-black/90 transition-all opacity-0 pointer-events-none select-none max-w-[1280px] max-h-[720px] aspect-video-vertical md:aspect-video",src:"/_storybook/iframe.html?viewMode=story&id=new-ui--transactions&shortcuts=false&singleStory=true&args="})}let o=(e,t,r,n,a)=>{let s=(e-t)*(a-n)/(r-t)+n;return s>0?s:0};function d(e){let{className:t="",quantity:r=40,staticity:a=50,ease:s=50,refresh:l=!1}=e,i=(0,c.useRef)(null),u=(0,c.useRef)(null),d=(0,c.useRef)(null),h=(0,c.useRef)([]),f=function(){let[e,t]=(0,c.useState)({x:0,y:0});return(0,c.useEffect)(()=>{let e=e=>{t({x:e.clientX,y:e.clientY})};return window.addEventListener("mousemove",e),()=>{window.removeEventListener("mousemove",e)}},[]),e}(),x=(0,c.useRef)({x:0,y:0}),m=(0,c.useRef)({w:0,h:0}),p=window.devicePixelRatio,v=(0,c.useCallback)(()=>{d.current&&d.current.clearRect(0,0,m.current.w,m.current.h)},[d]),w=(0,c.useCallback)(()=>{u.current&&i.current&&d.current&&(h.current.length=0,m.current.w=u.current.offsetWidth,m.current.h=u.current.offsetHeight,i.current.width=m.current.w*p,i.current.height=m.current.h*p,i.current.style.width="".concat(m.current.w,"px"),i.current.style.height="".concat(m.current.h,"px"),d.current.scale(p,p))},[p]),g=(0,c.useCallback)(()=>{let e=Math.floor(Math.random()*m.current.w);return{x:e,y:Math.floor(Math.random()*m.current.h),translateX:0,translateY:0,size:Math.floor(2*Math.random())+1,alpha:0,targetAlpha:parseFloat((.6*Math.random()+.1).toFixed(1)),dx:(Math.random()-.5)*.2,dy:(Math.random()-.5)*.2,magnetism:.1+4*Math.random()}},[]),y=(0,c.useCallback)(function(e){let t=arguments.length>1&&void 0!==arguments[1]&&arguments[1];if(d.current){let{x:r,y:n,translateX:a,translateY:s,size:l,alpha:i}=e;d.current.translate(a,s),d.current.beginPath(),d.current.arc(r,n,l,0,2*Math.PI),d.current.fillStyle="rgba(255, 255, 255, ".concat(i,")"),d.current.fill(),d.current.setTransform(p,0,0,p,0,0),t||h.current.push(e)}},[p]),b=(0,c.useCallback)(()=>{v();for(let e=0;e<r;e++)y(g())},[g,v,y,r]),j=(0,c.useCallback)(()=>{w(),b()},[b,w]),N=(0,c.useCallback)(()=>{v(),h.current.forEach((e,t)=>{let r=parseFloat(o([e.x+e.translateX-e.size,m.current.w-e.x-e.translateX-e.size,e.y+e.translateY-e.size,m.current.h-e.y-e.translateY-e.size].reduce((e,t)=>Math.min(e,t)),0,20,0,1).toFixed(2));r>1?(e.alpha+=.02,e.alpha>e.targetAlpha&&(e.alpha=e.targetAlpha)):e.alpha=e.targetAlpha*r,e.x+=e.dx,e.y+=e.dy,e.translateX+=(x.current.x/(a/e.magnetism)-e.translateX)/s,e.translateY+=(x.current.y/(a/e.magnetism)-e.translateY)/s,e.x<-e.size||e.x>m.current.w+e.size||e.y<-e.size||e.y>m.current.h+e.size?(h.current.splice(t,1),y(g())):y({...e,x:e.x,y:e.y,translateX:e.translateX,translateY:e.translateY,alpha:e.alpha},!0)}),window.requestAnimationFrame(N)},[g,v,y,s,a]);(0,c.useEffect)(()=>(i.current&&(d.current=i.current.getContext("2d")),j(),N(),window.addEventListener("resize",j),()=>{window.removeEventListener("resize",j)}),[N,j]);let z=(0,c.useCallback)(()=>{if(i.current){let e=i.current.getBoundingClientRect(),{w:t,h:r}=m.current,n=f.x-e.left-t/2,a=f.y-e.top-r/2;n<t/2&&n>-t/2&&a<r/2&&a>-r/2&&(x.current.x=n,x.current.y=a)}},[f.x,f.y]);return(0,c.useEffect)(()=>{z()},[f.x,f.y,z]),(0,c.useEffect)(()=>{j()},[j,l]),(0,n.jsx)("div",{className:t,ref:u,"aria-hidden":"true",children:(0,n.jsx)("canvas",{ref:i})})}var h=r(1714);function f(){return(0,n.jsxs)("div",{className:"w-full",children:[(0,n.jsx)("div",{className:"absolute inset-0 overflow-hidden pointer-events-none -z-10","aria-hidden":"true",children:(0,n.jsx)("div",{className:"absolute flex items-center justify-center top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 w-1/3 aspect-square",children:(0,n.jsx)("div",{className:"absolute inset-0 translate-z-0 bg-purple-500 rounded-full blur-[120px] opacity-50 min-h-[10vh]"})})}),(0,n.jsx)("div",{className:"max-md:hidden absolute bottom-0 -mb-20 left-2/3 -translate-x-1/2 blur-2xl opacity-70 pointer-events-none","aria-hidden":"true",children:(0,n.jsxs)("svg",{xmlns:"http://www.w3.org/2000/svg",width:"434",height:"427",children:[(0,n.jsx)("defs",{children:(0,n.jsxs)("linearGradient",{id:"bs5-a",x1:"19.609%",x2:"50%",y1:"14.544%",y2:"100%",children:[(0,n.jsx)("stop",{offset:"0%",stopColor:"#A855F7"}),(0,n.jsx)("stop",{offset:"100%",stopColor:"#6366F1",stopOpacity:"0"})]})}),(0,n.jsx)("path",{fill:"url(#bs5-a)",fillRule:"evenodd",d:"m661 736 461 369-284 58z",transform:"matrix(1 0 0 -1 -661 1163)"})]})}),(0,n.jsx)(d,{className:"absolute inset-0 -z-10"}),(0,n.jsxs)("div",{className:"m-view-height m-view-width flex flex-col py-8 mx-auto items-center justify-center",children:[(0,n.jsxs)("div",{className:"max-w-3xl flex flex-col",children:[(0,n.jsx)(i(),{src:h.Z,alt:"monetr logo",width:75,height:75}),(0,n.jsx)("h1",{className:"text-5xl font-bold",children:"monetr"}),(0,n.jsxs)("h2",{className:"text-xl font-medium",children:["monetr is currently in a ",(0,n.jsx)("b",{children:"closed beta"}),"! We are building a source-visible financial planning application focused on helping you plan and budget for recurring expenses, or future goals."]})]}),(0,n.jsx)(u,{})]})]})}let x=[];function m(e){return(0,n.jsx)(f,{})}var p=(0,a.j)({MDXContent:function(){let e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{},{wrapper:t}=Object.assign({},(0,s.a)(),e.components);return t?(0,n.jsx)(t,{...e,children:(0,n.jsx)(m,{...e})}):m(e)},pageOpts:{filePath:"src/pages/index.mdx",route:"/",frontMatter:{title:"monetr"},timestamp:1699159937e3,title:"monetr",headings:x},pageNextRoute:"/"})}},function(e){e.O(0,[9304,2888,9774,179],function(){return e(e.s=7803)}),_N_E=e.O()}]);