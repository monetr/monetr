"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[6110],{6110:function(e,t,s){s.d(t,{diagram:function(){return V}});var i=s(4725),a=s(8472),r=s(6357),o=s(9523),l=s(1197);s(7693),s(7608),s(1699),s(9500),s(9817);let d="rect",n="rectWithTitle",c="statediagram",p=`${c}-state`,b="transition",g=`${b} note-edge`,h=`${c}-note`,u=`${c}-cluster`,y=`${c}-cluster-alt`,f="parent",w="note",x="----",$=`${x}${w}`,m=`${x}${f}`,T="fill:none",S="fill: #333",k="text",D="normal",A={},v=0;function B(e="",t=0,s="",i=x){let a=null!==s&&s.length>0?`${i}${s}`:"";return`state-${e}${a}-${t}`}let E=(e,t,s,a,r,l)=>{var c;let b=s.id,x=null==(c=a[b])?"":c.classes?c.classes.join(" "):"";if("root"!==b){let t=d;!0===s.start&&(t="start"),!1===s.start&&(t="end"),s.type!==i.D&&(t=s.type),A[b]||(A[b]={id:b,shape:t,description:o.e.sanitizeText(b,(0,o.c)()),classes:`${x} ${p}`});let a=A[b];s.description&&(Array.isArray(a.description)?(a.shape=n,a.description.push(s.description)):a.description.length>0?(a.shape=n,a.description===b?a.description=[s.description]:a.description=[a.description,s.description]):(a.shape=d,a.description=s.description),a.description=o.e.sanitizeTextOrArray(a.description,(0,o.c)())),1===a.description.length&&a.shape===n&&(a.shape=d),!a.type&&s.doc&&(o.l.info("Setting cluster for ",b,C(s)),a.type="group",a.dir=C(s),a.shape=s.type===i.a?"divider":"roundedWithTitle",a.classes=a.classes+" "+u+" "+(l?y:""));let r={labelStyle:"",shape:a.shape,labelText:a.description,classes:a.classes,style:"",id:b,dir:a.dir,domId:B(b,v),type:a.type,padding:15};if(r.centerLabel=!0,s.note){let t={labelStyle:"",shape:"note",labelText:s.note.text,classes:h,style:"",id:b+$+"-"+v,domId:B(b,v,w),type:a.type,padding:15},i={labelStyle:"",shape:"noteGroup",labelText:s.note.text,classes:a.classes,style:"",id:b+m,domId:B(b,v,f),type:"group",padding:0};v++;let o=b+m;e.setNode(o,i),e.setNode(t.id,t),e.setNode(b,r),e.setParent(b,o),e.setParent(t.id,o);let l=b,d=t.id;"left of"===s.note.position&&(l=t.id,d=b),e.setEdge(l,d,{arrowhead:"none",arrowType:"",style:T,labelStyle:"",classes:g,arrowheadStyle:S,labelpos:"c",labelType:k,thickness:D})}else e.setNode(b,r)}t&&"root"!==t.id&&(o.l.trace("Setting node ",b," to be child of its parent ",t.id),e.setParent(b,t.id)),s.doc&&(o.l.trace("Adding nodes children "),N(e,s,s.doc,a,r,!l))},N=(e,t,s,a,r,l)=>{o.l.trace("items",s),s.forEach(s=>{switch(s.stmt){case i.b:case i.D:E(e,t,s,a,r,l);break;case i.S:{E(e,t,s.state1,a,r,l),E(e,t,s.state2,a,r,l);let i={id:"edge"+v,arrowhead:"normal",arrowTypeEnd:"arrow_barb",style:T,labelStyle:"",label:o.e.sanitizeText(s.description,(0,o.c)()),arrowheadStyle:S,labelpos:"c",labelType:k,thickness:D,classes:b};e.setEdge(s.state1.id,s.state2.id,i,v),v++}}})},C=(e,t=i.c)=>{let s=t;if(e.doc)for(let t=0;t<e.doc.length;t++){let i=e.doc[t];"dir"===i.stmt&&(s=i.value)}return s},R=async function(e,t,s,i){let n;o.l.info("Drawing state diagram (v2)",t),A={},i.db.getDirection();let{securityLevel:p,state:b}=(0,o.c)(),g=b.nodeSpacing||50,h=b.rankSpacing||50;o.l.info(i.db.getRootDocV2()),i.db.extract(i.db.getRootDocV2()),o.l.info(i.db.getRootDocV2());let u=i.db.getStates(),y=new a.k({multigraph:!0,compound:!0}).setGraph({rankdir:C(i.db.getRootDocV2()),nodesep:g,ranksep:h,marginx:8,marginy:8}).setDefaultEdgeLabel(function(){return{}});E(y,void 0,i.db.getRootDocV2(),u,i.db,!0),"sandbox"===p&&(n=(0,r.Ys)("#i"+t));let f="sandbox"===p?(0,r.Ys)(n.nodes()[0].contentDocument.body):(0,r.Ys)("body"),w=f.select(`[id="${t}"]`),x=f.select("#"+t+" g");await (0,l.r)(x,y,["barb"],c,t),o.u.insertTitle(w,"statediagramTitleText",b.titleTopMargin,i.db.getDiagramTitle());let $=w.node().getBBox(),m=$.width+16,T=$.height+16;w.attr("class",c);let S=w.node().getBBox();(0,o.i)(w,T,m,b.useMaxWidth);let k=`${S.x-8} ${S.y-8} ${m} ${T}`;for(let e of(o.l.debug(`viewBox ${k}`),w.attr("viewBox",k),document.querySelectorAll('[id="'+t+'"] .edgeLabel .label'))){let t=e.getBBox(),s=document.createElementNS("http://www.w3.org/2000/svg",d);s.setAttribute("rx",0),s.setAttribute("ry",0),s.setAttribute("width",t.width),s.setAttribute("height",t.height),e.insertBefore(s,e.firstChild)}},V={parser:i.p,db:i.d,renderer:{setConf:function(e){for(let t of Object.keys(e))e[t]},getClasses:function(e,t){return t.db.extract(t.db.getRootDocV2()),t.db.getClasses()},draw:R},styles:i.s,init:e=>{e.state||(e.state={}),e.state.arrowMarkerAbsolute=e.arrowMarkerAbsolute,i.d.clear()}}}}]);