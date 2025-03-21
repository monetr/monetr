"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[705],{1868:(t,e,n)=>{n.d(e,{DX:()=>u});var r=n(5834);function i(t,e){if("function"==typeof t)return t(e);null!=t&&(t.current=e)}var o=n(6514),u=r.forwardRef((t,e)=>{let{children:n,...i}=t,u=r.Children.toArray(n),c=u.find(a);if(c){let t=c.props.children,n=u.map(e=>e!==c?e:r.Children.count(t)>1?r.Children.only(null):r.isValidElement(t)?t.props.children:null);return(0,o.jsx)(l,{...i,ref:e,children:r.isValidElement(t)?r.cloneElement(t,void 0,n):null})}return(0,o.jsx)(l,{...i,ref:e,children:n})});u.displayName="Slot";var l=r.forwardRef((t,e)=>{let{children:n,...o}=t;if(r.isValidElement(n)){let t=function(t){let e=Object.getOwnPropertyDescriptor(t.props,"ref")?.get,n=e&&"isReactWarning"in e&&e.isReactWarning;return n?t.ref:(n=(e=Object.getOwnPropertyDescriptor(t,"ref")?.get)&&"isReactWarning"in e&&e.isReactWarning)?t.props.ref:t.props.ref||t.ref}(n),u=function(t,e){let n={...e};for(let r in e){let i=t[r],o=e[r];/^on[A-Z]/.test(r)?i&&o?n[r]=(...t)=>{o(...t),i(...t)}:i&&(n[r]=i):"style"===r?n[r]={...i,...o}:"className"===r&&(n[r]=[i,o].filter(Boolean).join(" "))}return{...t,...n}}(o,n.props);return n.type!==r.Fragment&&(u.ref=e?function(...t){return e=>{let n=!1,r=t.map(t=>{let r=i(t,e);return n||"function"!=typeof r||(n=!0),r});if(n)return()=>{for(let e=0;e<r.length;e++){let n=r[e];"function"==typeof n?n():i(t[e],null)}}}}(e,t):t),r.cloneElement(n,u)}return r.Children.count(n)>1?r.Children.only(null):null});l.displayName="SlotClone";var c=({children:t})=>(0,o.jsx)(o.Fragment,{children:t});function a(t){return r.isValidElement(t)&&t.type===c}},4200:(t,e,n)=>{n.d(e,{A:()=>r});let r=(0,n(5606).A)("ArrowLeft",[["path",{d:"m12 19-7-7 7-7",key:"1l729n"}],["path",{d:"M19 12H5",key:"x3x0zl"}]])},4779:(t,e,n)=>{n.d(e,{F:()=>u});var r=n(3367);let i=t=>"boolean"==typeof t?`${t}`:0===t?"0":t,o=r.$,u=(t,e)=>n=>{var r;if((null==e?void 0:e.variants)==null)return o(t,null==n?void 0:n.class,null==n?void 0:n.className);let{variants:u,defaultVariants:l}=e,c=Object.keys(u).map(t=>{let e=null==n?void 0:n[t],r=null==l?void 0:l[t];if(null===e)return null;let o=i(e)||i(r);return u[t][o]}),a=n&&Object.entries(n).reduce((t,e)=>{let[n,r]=e;return void 0===r||(t[n]=r),t},{});return o(t,c,null==e?void 0:null===(r=e.compoundVariants)||void 0===r?void 0:r.reduce((t,e)=>{let{class:n,className:r,...i}=e;return Object.entries(i).every(t=>{let[e,n]=t;return Array.isArray(n)?n.includes({...l,...a}[e]):({...l,...a})[e]===n})?[...t,n,r]:t},[]),null==n?void 0:n.class,null==n?void 0:n.className)}},6854:(t,e,n)=>{n.d(e,{A:()=>D});var r=n(5834);function i(t){return"[object Object]"===Object.prototype.toString.call(t)||Array.isArray(t)}function o(t,e){let n=Object.keys(t),r=Object.keys(e);return n.length===r.length&&JSON.stringify(Object.keys(t.breakpoints||{}))===JSON.stringify(Object.keys(e.breakpoints||{}))&&n.every(n=>{let r=t[n],u=e[n];return"function"==typeof r?`${r}`==`${u}`:i(r)&&i(u)?o(r,u):r===u})}function u(t){return t.concat().sort((t,e)=>t.name>e.name?1:-1).map(t=>t.options)}function l(t){return"number"==typeof t}function c(t){return"string"==typeof t}function a(t){return"boolean"==typeof t}function s(t){return"[object Object]"===Object.prototype.toString.call(t)}function f(t){return Math.abs(t)}function d(t){return Math.sign(t)}function p(t){return h(t).map(Number)}function m(t){return t[g(t)]}function g(t){return Math.max(0,t.length-1)}function y(t,e=0){return Array.from(Array(t),(t,n)=>e+n)}function h(t){return Object.keys(t)}function v(t,e){return void 0!==e.MouseEvent&&t instanceof e.MouseEvent}function b(){let t=[],e={add:function(n,r,i,o={passive:!0}){let u;return"addEventListener"in n?(n.addEventListener(r,i,o),u=()=>n.removeEventListener(r,i,o)):(n.addListener(i),u=()=>n.removeListener(i)),t.push(u),e},clear:function(){t=t.filter(t=>t())}};return e}function x(t=0,e=0){let n=f(t-e);function r(n){return n<t||n>e}return{length:n,max:e,min:t,constrain:function(n){return r(n)?n<t?t:e:n},reachedAny:r,reachedMax:function(t){return t>e},reachedMin:function(e){return e<t},removeOffset:function(t){return n?t-n*Math.ceil((t-e)/n):t}}}function O(t){let e=t;function n(t){return l(t)?t:t.get()}return{get:function(){return e},set:function(t){e=n(t)},add:function(t){e+=n(t)},subtract:function(t){e-=n(t)}}}function S(t,e){let n="x"===t.scroll?function(t){return`translate3d(${t}px,0px,0px)`}:function(t){return`translate3d(0px,${t}px,0px)`},r=e.style,i=null,o=!1;return{clear:function(){o||(r.transform="",e.getAttribute("style")||e.removeAttribute("style"))},to:function(e){if(o)return;let u=Math.round(100*t.direction(e))/100;u!==i&&(r.transform=n(u),i=u)},toggleActive:function(t){o=!t}}}let w={align:"center",axis:"x",container:null,slides:null,containScroll:"trimSnaps",direction:"ltr",slidesToScroll:1,inViewThreshold:0,breakpoints:{},dragFree:!1,dragThreshold:10,loop:!1,skipSnaps:!1,duration:25,startIndex:0,active:!0,watchDrag:!0,watchResize:!0,watchSlides:!0,watchFocus:!0};function E(t,e,n){let r,i,o,u,D;let I=t.ownerDocument,k=I.defaultView,A=function(t){function e(t,e){return function t(e,n){return[e,n].reduce((e,n)=>(h(n).forEach(r=>{let i=e[r],o=n[r],u=s(i)&&s(o);e[r]=u?t(i,o):o}),e),{})}(t,e||{})}return{mergeOptions:e,optionsAtMedia:function(n){let r=n.breakpoints||{},i=h(r).filter(e=>t.matchMedia(e).matches).map(t=>r[t]).reduce((t,n)=>e(t,n),{});return e(n,i)},optionsMediaQueries:function(e){return e.map(t=>h(t.breakpoints||{})).reduce((t,e)=>t.concat(e),[]).map(t.matchMedia)}}}(k),F=(D=[],{init:function(t,e){return(D=e.filter(({options:t})=>!1!==A.optionsAtMedia(t).active)).forEach(e=>e.init(t,A)),e.reduce((t,e)=>Object.assign(t,{[e.name]:e}),{})},destroy:function(){D=D.filter(t=>t.destroy())}}),L=b(),N=function(){let t,e={},n={init:function(e){t=e},emit:function(r){return(e[r]||[]).forEach(e=>e(t,r)),n},off:function(t,r){return e[t]=(e[t]||[]).filter(t=>t!==r),n},on:function(t,r){return e[t]=(e[t]||[]).concat([r]),n},clear:function(){e={}}};return n}(),{mergeOptions:j,optionsAtMedia:M,optionsMediaQueries:T}=A,{on:P,off:V,emit:C}=N,R=!1,H=j(w,E.globalOptions),z=j(H),$=[];function B(e,n){!R&&(z=M(H=j(H,e)),$=n||$,function(){let{container:e,slides:n}=z;o=(c(e)?t.querySelector(e):e)||t.children[0];let r=c(n)?o.querySelectorAll(n):n;u=[].slice.call(r||o.children)}(),r=function e(n){let r=function(t,e,n,r,i,o,u){let s,w;let{align:E,axis:D,direction:I,startIndex:k,loop:A,duration:F,dragFree:L,dragThreshold:N,inViewThreshold:j,slidesToScroll:M,skipSnaps:T,containScroll:P,watchResize:V,watchSlides:C,watchDrag:R,watchFocus:H}=o,z={measure:function(t){let{offsetTop:e,offsetLeft:n,offsetWidth:r,offsetHeight:i}=t;return{top:e,right:n+r,bottom:e+i,left:n,width:r,height:i}}},$=z.measure(e),B=n.map(z.measure),U=function(t,e){let n="rtl"===e,r="y"===t,i=!r&&n?-1:1;return{scroll:r?"y":"x",cross:r?"x":"y",startEdge:r?"top":n?"right":"left",endEdge:r?"bottom":n?"left":"right",measureSize:function(t){let{height:e,width:n}=t;return r?e:n},direction:function(t){return t*i}}}(D,I),q=U.measureSize($),W={measure:function(t){return t/100*q}},_=function(t,e){let n={start:function(){return 0},center:function(t){return(e-t)/2},end:function(t){return e-t}};return{measure:function(r,i){return c(t)?n[t](r):t(e,r,i)}}}(E,q),X=!A&&!!P,{slideSizes:J,slideSizesWithGaps:Q,startGap:Y,endGap:Z}=function(t,e,n,r,i,o){let{measureSize:u,startEdge:l,endEdge:c}=t,a=n[0]&&i,s=function(){if(!a)return 0;let t=n[0];return f(e[l]-t[l])}(),d=a?parseFloat(o.getComputedStyle(m(r)).getPropertyValue(`margin-${c}`)):0,p=n.map(u),y=n.map((t,e,n)=>{let r=e===g(n);return e?r?p[e]+d:n[e+1][l]-t[l]:p[e]+s}).map(f);return{slideSizes:p,slideSizesWithGaps:y,startGap:s,endGap:d}}(U,$,B,n,A||!!P,i),G=function(t,e,n,r,i,o,u,c,a){let{startEdge:s,endEdge:d,direction:y}=t,h=l(n);return{groupSlides:function(t){return h?p(t).filter(t=>t%n==0).map(e=>t.slice(e,e+n)):t.length?p(t).reduce((n,l,a)=>{let p=m(n)||0,h=l===g(t),v=i[s]-o[p][s],b=i[s]-o[l][d],x=r||0!==p?0:y(u),O=f(b-(!r&&h?y(c):0)-(v+x));return a&&O>e+2&&n.push(l),h&&n.push(t.length),n},[]).map((e,n,r)=>{let i=Math.max(r[n-1]||0);return t.slice(i,e)}):[]}}}(U,q,M,A,$,B,Y,Z,0),{snaps:K,snapsAligned:tt}=function(t,e,n,r,i){let{startEdge:o,endEdge:u}=t,{groupSlides:l}=i,c=l(r).map(t=>m(t)[u]-t[0][o]).map(f).map(e.measure),a=r.map(t=>n[o]-t[o]).map(t=>-f(t)),s=l(a).map(t=>t[0]).map((t,e)=>t+c[e]);return{snaps:a,snapsAligned:s}}(U,_,$,B,G),te=-m(K)+m(Q),{snapsContained:tn,scrollContainLimit:tr}=function(t,e,n,r,i){let o=x(-e+t,0),u=n.map((t,e)=>{let{min:r,max:i}=o,u=o.constrain(t),l=e===g(n);return e?l||function(t,e){return 1>=f(t-e)}(r,u)?r:function(t,e){return 1>=f(t-e)}(i,u)?i:u:i}).map(t=>parseFloat(t.toFixed(3))),l=function(){let t=u[0],e=m(u);return x(u.lastIndexOf(t),u.indexOf(e)+1)}();function c(t,e){return 1>=f(t-e)}return{snapsContained:function(){if(e<=t+2)return[o.max];if("keepSnaps"===r)return u;let{min:n,max:i}=l;return u.slice(n,i)}(),scrollContainLimit:l}}(q,te,tt,P,0),ti=X?tn:tt,{limit:to}=function(t,e,n){let r=e[0];return{limit:x(n?r-t:m(e),r)}}(te,ti,A),tu=function t(e,n,r){let{constrain:i}=x(0,e),o=e+1,u=l(n);function l(t){return r?f((o+t)%o):i(t)}function c(){return t(e,u,r)}let a={get:function(){return u},set:function(t){return u=l(t),a},add:function(t){return c().set(u+t)},clone:c};return a}(g(ti),k,A),tl=tu.clone(),tc=p(n),ta=({dragHandler:t,scrollBody:e,scrollBounds:n,options:{loop:r}})=>{r||n.constrain(t.pointerDown()),e.seek()},ts=({scrollBody:t,translate:e,location:n,offsetLocation:r,previousLocation:i,scrollLooper:o,slideLooper:u,dragHandler:l,animation:c,eventHandler:a,scrollBounds:s,options:{loop:f}},d)=>{let p=t.settled(),m=!s.shouldConstrain(),g=f?p:p&&m;g&&!l.pointerDown()&&(c.stop(),a.emit("settle")),g||a.emit("scroll");let y=n.get()*d+i.get()*(1-d);r.set(y),f&&(o.loop(t.direction()),u.loop()),e.to(r.get())},tf=function(t,e,n,r){let i=b(),o=1e3/60,u=null,l=0,c=0;function a(t){if(!c)return;u||(u=t,n(),n());let i=t-u;for(u=t,l+=i;l>=o;)n(),l-=o;r(l/o),c&&(c=e.requestAnimationFrame(a))}function s(){e.cancelAnimationFrame(c),u=null,l=0,c=0}return{init:function(){i.add(t,"visibilitychange",()=>{t.hidden&&(u=null,l=0)})},destroy:function(){s(),i.clear()},start:function(){c||(c=e.requestAnimationFrame(a))},stop:s,update:n,render:r}}(r,i,()=>ta(tD),t=>ts(tD,t)),td=ti[tu.get()],tp=O(td),tm=O(td),tg=O(td),ty=O(td),th=function(t,e,n,r,i,o){let u=0,l=0,c=i,a=.68,s=t.get(),p=0;function m(t){return c=t,y}function g(t){return a=t,y}let y={direction:function(){return l},duration:function(){return c},velocity:function(){return u},seek:function(){let e=r.get()-t.get(),i=0;return c?(n.set(t),u+=e/c,u*=a,s+=u,t.add(u),i=s-p):(u=0,n.set(r),t.set(r),i=e),l=d(i),p=s,y},settled:function(){return .001>f(r.get()-e.get())},useBaseFriction:function(){return g(.68)},useBaseDuration:function(){return m(i)},useFriction:g,useDuration:m};return y}(tp,tg,tm,ty,F,.68),tv=function(t,e,n,r,i){let{reachedAny:o,removeOffset:u,constrain:l}=r;function c(t){return t.concat().sort((t,e)=>f(t)-f(e))[0]}function a(e,r){let i=[e,e+n,e-n];if(!t)return e;if(!r)return c(i);let o=i.filter(t=>d(t)===r);return o.length?c(o):m(i)-n}return{byDistance:function(n,r){let c=i.get()+n,{index:s,distance:d}=function(n){let r=t?u(n):l(n),{index:i}=e.map((t,e)=>({diff:a(t-r,0),index:e})).sort((t,e)=>f(t.diff)-f(e.diff))[0];return{index:i,distance:r}}(c),p=!t&&o(c);if(!r||p)return{index:s,distance:n};let m=n+a(e[s]-d,0);return{index:s,distance:m}},byIndex:function(t,n){let r=a(e[t]-i.get(),n);return{index:t,distance:r}},shortcut:a}}(A,ti,te,to,ty),tb=function(t,e,n,r,i,o,u){function l(i){let l=i.distance,c=i.index!==e.get();o.add(l),l&&(r.duration()?t.start():(t.update(),t.render(1),t.update())),c&&(n.set(e.get()),e.set(i.index),u.emit("select"))}return{distance:function(t,e){l(i.byDistance(t,e))},index:function(t,n){let r=e.clone().set(t);l(i.byIndex(r.get(),n))}}}(tf,tu,tl,th,tv,ty,u),tx=function(t){let{max:e,length:n}=t;return{get:function(t){return n?-((t-e)/n):0}}}(to),tO=b(),tS=function(t,e,n,r){let i;let o={},u=null,l=null,c=!1;return{init:function(){i=new IntersectionObserver(t=>{c||(t.forEach(t=>{o[e.indexOf(t.target)]=t}),u=null,l=null,n.emit("slidesInView"))},{root:t.parentElement,threshold:r}),e.forEach(t=>i.observe(t))},destroy:function(){i&&i.disconnect(),c=!0},get:function(t=!0){if(t&&u)return u;if(!t&&l)return l;let e=h(o).reduce((e,n)=>{let r=parseInt(n),{isIntersecting:i}=o[r];return(t&&i||!t&&!i)&&e.push(r),e},[]);return t&&(u=e),t||(l=e),e}}}(e,n,u,j),{slideRegistry:tw}=function(t,e,n,r,i,o){let{groupSlides:u}=i,{min:l,max:c}=r;return{slideRegistry:function(){let r=u(o);return 1===n.length?[o]:t&&"keepSnaps"!==e?r.slice(l,c).map((t,e,n)=>{let r=e===g(n);return e?r?y(g(o)-m(n)[0]+1,m(n)[0]):t:y(m(n[0])+1)}):r}()}}(X,P,ti,tr,G,tc),tE=function(t,e,n,r,i,o,u,c){let s={passive:!0,capture:!0},f=0;function d(t){"Tab"===t.code&&(f=new Date().getTime())}return{init:function(p){c&&(o.add(document,"keydown",d,!1),e.forEach((e,d)=>{o.add(e,"focus",e=>{(a(c)||c(p,e))&&function(e){if(new Date().getTime()-f>10)return;u.emit("slideFocusStart"),t.scrollLeft=0;let o=n.findIndex(t=>t.includes(e));l(o)&&(i.useDuration(0),r.index(o,0),u.emit("slideFocus"))}(d)},s)}))}}}(t,n,tw,tb,th,tO,u,H),tD={ownerDocument:r,ownerWindow:i,eventHandler:u,containerRect:$,slideRects:B,animation:tf,axis:U,dragHandler:function(t,e,n,r,i,o,u,l,c,s,p,m,g,y,h,O,S,w,E){let{cross:D,direction:I}=t,k=["INPUT","SELECT","TEXTAREA"],A={passive:!1},F=b(),L=b(),N=x(50,225).constrain(y.measure(20)),j={mouse:300,touch:400},M={mouse:500,touch:600},T=h?43:25,P=!1,V=0,C=0,R=!1,H=!1,z=!1,$=!1;function B(t){if(!v(t,r)&&t.touches.length>=2)return U(t);let e=o.readPoint(t),n=o.readPoint(t,D),u=f(e-V),c=f(n-C);if(!H&&!$&&(!t.cancelable||!(H=u>c)))return U(t);let a=o.pointerMove(t);u>O&&(z=!0),s.useFriction(.3).useDuration(.75),l.start(),i.add(I(a)),t.preventDefault()}function U(t){let e=p.byDistance(0,!1).index!==m.get(),n=o.pointerUp(t)*(h?M:j)[$?"mouse":"touch"],r=function(t,e){let n=m.add(-1*d(t)),r=p.byDistance(t,!h).distance;return h||f(t)<N?r:S&&e?.5*r:p.byIndex(n.get(),0).distance}(I(n),e),i=function(t,e){var n,r;if(0===t||0===e||f(t)<=f(e))return 0;let i=(n=f(t),r=f(e),f(n-r));return f(i/t)}(n,r);H=!1,R=!1,L.clear(),s.useDuration(T-10*i).useFriction(.68+i/50),c.distance(r,!h),$=!1,g.emit("pointerUp")}function q(t){z&&(t.stopPropagation(),t.preventDefault(),z=!1)}return{init:function(t){E&&F.add(e,"dragstart",t=>t.preventDefault(),A).add(e,"touchmove",()=>void 0,A).add(e,"touchend",()=>void 0).add(e,"touchstart",l).add(e,"mousedown",l).add(e,"touchcancel",U).add(e,"contextmenu",U).add(e,"click",q,!0);function l(l){(a(E)||E(t,l))&&function(t){let l=v(t,r);$=l,z=h&&l&&!t.buttons&&P,P=f(i.get()-u.get())>=2,(!l||0===t.button)&&!function(t){let e=t.nodeName||"";return k.includes(e)}(t.target)&&(R=!0,o.pointerDown(t),s.useFriction(0).useDuration(0),i.set(u),function(){let t=$?n:e;L.add(t,"touchmove",B,A).add(t,"touchend",U).add(t,"mousemove",B,A).add(t,"mouseup",U)}(),V=o.readPoint(t),C=o.readPoint(t,D),g.emit("pointerDown"))}(l)}},destroy:function(){F.clear(),L.clear()},pointerDown:function(){return R}}}(U,t,r,i,ty,function(t,e){let n,r;function i(t){return t.timeStamp}function o(n,r){let i=r||t.scroll,o=`client${"x"===i?"X":"Y"}`;return(v(n,e)?n:n.touches[0])[o]}return{pointerDown:function(t){return n=t,r=t,o(t)},pointerMove:function(t){let e=o(t)-o(r),u=i(t)-i(n)>170;return r=t,u&&(n=t),e},pointerUp:function(t){if(!n||!r)return 0;let e=o(r)-o(n),u=i(t)-i(n),l=i(t)-i(r)>170,c=e/u;return u&&!l&&f(c)>.1?c:0},readPoint:o}}(U,i),tp,tf,tb,th,tv,tu,u,W,L,N,T,0,R),eventStore:tO,percentOfView:W,index:tu,indexPrevious:tl,limit:to,location:tp,offsetLocation:tg,previousLocation:tm,options:o,resizeHandler:function(t,e,n,r,i,o,u){let l,c;let s=[t].concat(r),d=[],p=!1;function m(t){return i.measureSize(u.measure(t))}return{init:function(i){o&&(c=m(t),d=r.map(m),l=new ResizeObserver(n=>{(a(o)||o(i,n))&&function(n){for(let o of n){if(p)return;let n=o.target===t,u=r.indexOf(o.target),l=n?c:d[u];if(f(m(n?t:r[u])-l)>=.5){i.reInit(),e.emit("resize");break}}}(n)}),n.requestAnimationFrame(()=>{s.forEach(t=>l.observe(t))}))},destroy:function(){p=!0,l&&l.disconnect()}}}(e,u,i,n,U,V,z),scrollBody:th,scrollBounds:function(t,e,n,r,i){let o=i.measure(10),u=i.measure(50),l=x(.1,.99),c=!1;function a(){return!!(!c&&t.reachedAny(n.get())&&t.reachedAny(e.get()))}return{shouldConstrain:a,constrain:function(i){if(!a())return;let c=t.reachedMin(e.get())?"min":"max",s=f(t[c]-e.get()),d=n.get()-e.get(),p=l.constrain(s/u);n.subtract(d*p),!i&&f(d)<o&&(n.set(t.constrain(n.get())),r.useDuration(25).useBaseFriction())},toggleActive:function(t){c=!t}}}(to,tg,ty,th,W),scrollLooper:function(t,e,n,r){let{reachedMin:i,reachedMax:o}=x(e.min+.1,e.max+.1);return{loop:function(e){if(!(1===e?o(n.get()):-1===e&&i(n.get())))return;let u=-1*e*t;r.forEach(t=>t.add(u))}}}(te,to,tg,[tp,tg,tm,ty]),scrollProgress:tx,scrollSnapList:ti.map(tx.get),scrollSnaps:ti,scrollTarget:tv,scrollTo:tb,slideLooper:function(t,e,n,r,i,o,u,l,c){let a=p(i),s=p(i).reverse(),f=g(m(s,u[0]),n,!1).concat(g(m(a,e-u[0]-1),-n,!0));function d(t,e){return t.reduce((t,e)=>t-i[e],e)}function m(t,e){return t.reduce((t,n)=>d(t,e)>0?t.concat([n]):t,[])}function g(i,u,a){let s=o.map((t,n)=>({start:t-r[n]+.5+u,end:t+e-.5+u}));return i.map(e=>{let r=a?0:-n,i=a?n:0,o=s[e][a?"end":"start"];return{index:e,loopPoint:o,slideLocation:O(-1),translate:S(t,c[e]),target:()=>l.get()>o?r:i}})}return{canLoop:function(){return f.every(({index:t})=>.1>=d(a.filter(e=>e!==t),e))},clear:function(){f.forEach(t=>t.translate.clear())},loop:function(){f.forEach(t=>{let{target:e,translate:n,slideLocation:r}=t,i=e();i!==r.get()&&(n.to(i),r.set(i))})},loopPoints:f}}(U,q,te,J,Q,K,ti,tg,n),slideFocus:tE,slidesHandler:(w=!1,{init:function(t){C&&(s=new MutationObserver(e=>{!w&&(a(C)||C(t,e))&&function(e){for(let n of e)if("childList"===n.type){t.reInit(),u.emit("slidesChanged");break}}(e)})).observe(e,{childList:!0})},destroy:function(){s&&s.disconnect(),w=!0}}),slidesInView:tS,slideIndexes:tc,slideRegistry:tw,slidesToScroll:G,target:ty,translate:S(U,e)};return tD}(t,o,u,I,k,n,N);return n.loop&&!r.slideLooper.canLoop()?e(Object.assign({},n,{loop:!1})):r}(z),T([H,...$.map(({options:t})=>t)]).forEach(t=>L.add(t,"change",U)),z.active&&(r.translate.to(r.location.get()),r.animation.init(),r.slidesInView.init(),r.slideFocus.init(X),r.eventHandler.init(X),r.resizeHandler.init(X),r.slidesHandler.init(X),r.options.loop&&r.slideLooper.loop(),o.offsetParent&&u.length&&r.dragHandler.init(X),i=F.init(X,$)))}function U(t,e){let n=_();q(),B(j({startIndex:n},t),e),N.emit("reInit")}function q(){r.dragHandler.destroy(),r.eventStore.clear(),r.translate.clear(),r.slideLooper.clear(),r.resizeHandler.destroy(),r.slidesHandler.destroy(),r.slidesInView.destroy(),r.animation.destroy(),F.destroy(),L.clear()}function W(t,e,n){z.active&&!R&&(r.scrollBody.useBaseFriction().useDuration(!0===e?0:z.duration),r.scrollTo.index(t,n||0))}function _(){return r.index.get()}let X={canScrollNext:function(){return r.index.add(1).get()!==_()},canScrollPrev:function(){return r.index.add(-1).get()!==_()},containerNode:function(){return o},internalEngine:function(){return r},destroy:function(){R||(R=!0,L.clear(),q(),N.emit("destroy"),N.clear())},off:V,on:P,emit:C,plugins:function(){return i},previousScrollSnap:function(){return r.indexPrevious.get()},reInit:U,rootNode:function(){return t},scrollNext:function(t){W(r.index.add(1).get(),t,-1)},scrollPrev:function(t){W(r.index.add(-1).get(),t,1)},scrollProgress:function(){return r.scrollProgress.get(r.location.get())},scrollSnapList:function(){return r.scrollSnapList},scrollTo:W,selectedScrollSnap:_,slideNodes:function(){return u},slidesInView:function(){return r.slidesInView.get()},slidesNotInView:function(){return r.slidesInView.get(!1)}};return B(e,n),setTimeout(()=>N.emit("init"),0),X}function D(t={},e=[]){let n=(0,r.useRef)(t),i=(0,r.useRef)(e),[l,c]=(0,r.useState)(),[a,s]=(0,r.useState)(),f=(0,r.useCallback)(()=>{l&&l.reInit(n.current,i.current)},[l]);return(0,r.useEffect)(()=>{o(n.current,t)||(n.current=t,f())},[t,f]),(0,r.useEffect)(()=>{!function(t,e){if(t.length!==e.length)return!1;let n=u(t),r=u(e);return n.every((t,e)=>o(t,r[e]))}(i.current,e)&&(i.current=e,f())},[e,f]),(0,r.useEffect)(()=>{if("undefined"!=typeof window&&window.document&&window.document.createElement&&a){E.globalOptions=D.globalOptions;let t=E(a,n.current,i.current);return c(t),()=>t.destroy()}c(void 0)},[a,c]),[s,l]}E.globalOptions=void 0,D.globalOptions=void 0},8202:(t,e,n)=>{n.d(e,{A:()=>i});let r={active:!0,breakpoints:{},delay:4e3,jump:!1,playOnInit:!0,stopOnFocusIn:!0,stopOnInteraction:!0,stopOnMouseEnter:!1,stopOnLastSnap:!1,rootNode:null};function i(t={}){let e,n,o,u;let l=null,c=0,a=!1,s=!1,f=!1,d=!1;function p(){if(!o){if(y()){f=!0;return}a||n.emit("autoplay:play"),function(){let{ownerWindow:t}=n.internalEngine();t.clearTimeout(c),c=t.setTimeout(O,u[n.selectedScrollSnap()]),l=new Date().getTime(),n.emit("autoplay:timerset")}(),a=!0}}function m(){o||(a&&n.emit("autoplay:stop"),function(){let{ownerWindow:t}=n.internalEngine();t.clearTimeout(c),c=0,l=null,n.emit("autoplay:timerstopped")}(),a=!1)}function g(){if(y())return f=a,m();f&&p()}function y(){let{ownerDocument:t}=n.internalEngine();return"hidden"===t.visibilityState}function h(){s||m()}function v(){s||p()}function b(){s=!0,m()}function x(){s=!1,p()}function O(){let{index:t}=n.internalEngine(),r=t.clone().add(1).get(),i=n.scrollSnapList().length-1,o=e.stopOnLastSnap&&r===i;if(n.canScrollNext()?n.scrollNext(d):n.scrollTo(0,d),n.emit("autoplay:select"),o)return m();p()}return{name:"autoplay",options:t,init:function(l,c){n=l;let{mergeOptions:a,optionsAtMedia:s}=c,f=a(r,i.globalOptions);if(e=s(a(f,t)),n.scrollSnapList().length<=1)return;d=e.jump,o=!1,u=function(t,e){let n=t.scrollSnapList();return"number"==typeof e?n.map(()=>e):e(n,t)}(n,e.delay);let{eventStore:y,ownerDocument:O}=n.internalEngine(),S=!!n.internalEngine().options.watchDrag,w=function(t,e){let n=t.rootNode();return e&&e(n)||n}(n,e.rootNode);y.add(O,"visibilitychange",g),S&&n.on("pointerDown",h),S&&!e.stopOnInteraction&&n.on("pointerUp",v),e.stopOnMouseEnter&&y.add(w,"mouseenter",b),e.stopOnMouseEnter&&!e.stopOnInteraction&&y.add(w,"mouseleave",x),e.stopOnFocusIn&&n.on("slideFocusStart",m),e.stopOnFocusIn&&!e.stopOnInteraction&&y.add(n.containerNode(),"focusout",p),e.playOnInit&&p()},destroy:function(){n.off("pointerDown",h).off("pointerUp",v).off("slideFocusStart",m),m(),o=!0,a=!1},play:function(t){void 0!==t&&(d=t),p()},stop:function(){a&&m()},reset:function(){a&&p()},isPlaying:function(){return a},timeUntilNext:function(){return l?u[n.selectedScrollSnap()]-(new Date().getTime()-l):null}}}i.globalOptions=void 0}}]);