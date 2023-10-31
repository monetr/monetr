"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[8574],{8574:function(e,t,s){s.d(t,{diagram:function(){return B}});var i=s(7308),o=s(8472),r=s(6357),a=s(6388),l=s(5220);s(7693),s(7608),s(1699),s(9500),s(6576);let d="rect",n="rectWithTitle",c="statediagram",p=`${c}-state`,g="transition",b=`${g} note-edge`,u=`${c}-note`,h=`${c}-cluster`,y=`${c}-cluster-alt`,f="parent",m="note",w="----",D=`${w}${m}`,x=`${w}${f}`,$="fill:none",T="fill: #333",S="text",k="normal",N={},A=0;function getClassesFromDbInfo(e){return null==e?"":e.classes?e.classes.join(" "):""}function stateDomId(e="",t=0,s="",i=w){let o=null!==s&&s.length>0?`${i}${s}`:"";return`state-${e}${o}-${t}`}let setupNode=(e,t,s,o,r,l)=>{let c=s.id,g=getClassesFromDbInfo(o[c]);if("root"!==c){let t=d;!0===s.start&&(t="start"),!1===s.start&&(t="end"),s.type!==i.D&&(t=s.type),N[c]||(N[c]={id:c,shape:t,description:a.e.sanitizeText(c,(0,a.c)()),classes:`${g} ${p}`});let o=N[c];s.description&&(Array.isArray(o.description)?(o.shape=n,o.description.push(s.description)):o.description.length>0?(o.shape=n,o.description===c?o.description=[s.description]:o.description=[o.description,s.description]):(o.shape=d,o.description=s.description),o.description=a.e.sanitizeTextOrArray(o.description,(0,a.c)())),1===o.description.length&&o.shape===n&&(o.shape=d),!o.type&&s.doc&&(a.l.info("Setting cluster for ",c,getDir(s)),o.type="group",o.dir=getDir(s),o.shape=s.type===i.a?"divider":"roundedWithTitle",o.classes=o.classes+" "+h+" "+(l?y:""));let r={labelStyle:"",shape:o.shape,labelText:o.description,classes:o.classes,style:"",id:c,dir:o.dir,domId:stateDomId(c,A),type:o.type,padding:15};if(r.centerLabel=!0,s.note){let t={labelStyle:"",shape:"note",labelText:s.note.text,classes:u,style:"",id:c+D+"-"+A,domId:stateDomId(c,A,m),type:o.type,padding:15},i={labelStyle:"",shape:"noteGroup",labelText:s.note.text,classes:o.classes,style:"",id:c+x,domId:stateDomId(c,A,f),type:"group",padding:0};A++;let a=c+x;e.setNode(a,i),e.setNode(t.id,t),e.setNode(c,r),e.setParent(c,a),e.setParent(t.id,a);let l=c,d=t.id;"left of"===s.note.position&&(l=t.id,d=c),e.setEdge(l,d,{arrowhead:"none",arrowType:"",style:$,labelStyle:"",classes:b,arrowheadStyle:T,labelpos:"c",labelType:S,thickness:k})}else e.setNode(c,r)}t&&"root"!==t.id&&(a.l.trace("Setting node ",c," to be child of its parent ",t.id),e.setParent(c,t.id)),s.doc&&(a.l.trace("Adding nodes children "),setupDoc(e,s,s.doc,o,r,!l))},setupDoc=(e,t,s,o,r,l)=>{a.l.trace("items",s),s.forEach(s=>{switch(s.stmt){case i.b:case i.D:setupNode(e,t,s,o,r,l);break;case i.S:{setupNode(e,t,s.state1,o,r,l),setupNode(e,t,s.state2,o,r,l);let i={id:"edge"+A,arrowhead:"normal",arrowTypeEnd:"arrow_barb",style:$,labelStyle:"",label:a.e.sanitizeText(s.description,(0,a.c)()),arrowheadStyle:T,labelpos:"c",labelType:S,thickness:k,classes:g};e.setEdge(s.state1.id,s.state2.id,i,A),A++}}})},getDir=(e,t=i.c)=>{let s=t;if(e.doc)for(let t=0;t<e.doc.length;t++){let i=e.doc[t];"dir"===i.stmt&&(s=i.value)}return s},draw=async function(e,t,s,i){let n;a.l.info("Drawing state diagram (v2)",t),N={},i.db.getDirection();let{securityLevel:p,state:g}=(0,a.c)(),b=g.nodeSpacing||50,u=g.rankSpacing||50;a.l.info(i.db.getRootDocV2()),i.db.extract(i.db.getRootDocV2()),a.l.info(i.db.getRootDocV2());let h=i.db.getStates(),y=new o.k({multigraph:!0,compound:!0}).setGraph({rankdir:getDir(i.db.getRootDocV2()),nodesep:b,ranksep:u,marginx:8,marginy:8}).setDefaultEdgeLabel(function(){return{}});setupNode(y,void 0,i.db.getRootDocV2(),h,i.db,!0),"sandbox"===p&&(n=(0,r.Ys)("#i"+t));let f="sandbox"===p?(0,r.Ys)(n.nodes()[0].contentDocument.body):(0,r.Ys)("body"),m=f.select(`[id="${t}"]`),w=f.select("#"+t+" g");await (0,l.r)(w,y,["barb"],c,t),a.u.insertTitle(m,"statediagramTitleText",g.titleTopMargin,i.db.getDiagramTitle());let D=m.node().getBBox(),x=D.width+16,$=D.height+16;m.attr("class",c);let T=m.node().getBBox();(0,a.i)(m,$,x,g.useMaxWidth);let S=`${T.x-8} ${T.y-8} ${x} ${$}`;a.l.debug(`viewBox ${S}`),m.attr("viewBox",S);let k=document.querySelectorAll('[id="'+t+'"] .edgeLabel .label');for(let e of k){let t=e.getBBox(),s=document.createElementNS("http://www.w3.org/2000/svg",d);s.setAttribute("rx",0),s.setAttribute("ry",0),s.setAttribute("width",t.width),s.setAttribute("height",t.height),e.insertBefore(s,e.firstChild)}},B={parser:i.p,db:i.d,renderer:{setConf:function(e){let t=Object.keys(e);for(let s of t)e[s]},getClasses:function(e,t){return t.db.extract(t.db.getRootDocV2()),t.db.getClasses()},draw},styles:i.s,init:e=>{e.state||(e.state={}),e.state.arrowMarkerAbsolute=e.arrowMarkerAbsolute,i.d.clear()}}}}]);