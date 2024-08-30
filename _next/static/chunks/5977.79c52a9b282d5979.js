"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[5977],{5977:function(t,e,a){a.d(e,{diagram:function(){return x}});var r=a(325),i=a(6357),n=a(9500),d=a(8472),o=a(89);a(7693),a(7608),a(1699);let l=0,s=function(t){let e=t.id;return t.type&&(e+="<"+(0,o.x)(t.type)+">"),e},p=function(t,e,a,r){let{displayText:i,cssStyle:n}=e.getDisplayDetails(),d=t.append("tspan").attr("x",r.padding).text(i);""!==n&&d.attr("style",e.cssStyle),a||d.attr("dy",r.textHeight)},g={drawClass:function(t,e,a,r){let i,n,d,l;o.l.debug("Rendering class ",e,a);let g=e.id,h={id:g,label:e.id,width:0,height:0},c=t.append("g").attr("id",r.db.lookUpDomId(g)).attr("class","classGroup");i=e.link?c.append("svg:a").attr("xlink:href",e.link).attr("target",e.linkTarget).append("text").attr("y",a.textHeight+a.padding).attr("x",0):c.append("text").attr("y",a.textHeight+a.padding).attr("x",0);let f=!0;e.annotations.forEach(function(t){let e=i.append("tspan").text("\xab"+t+"\xbb");f||e.attr("dy",a.textHeight),f=!1});let x=s(e),u=i.append("tspan").text(x).attr("class","title");f||u.attr("dy",a.textHeight);let y=i.node().getBBox().height;if(e.members.length>0){n=c.append("line").attr("x1",0).attr("y1",a.padding+y+a.dividerMargin/2).attr("y2",a.padding+y+a.dividerMargin/2);let t=c.append("text").attr("x",a.padding).attr("y",y+a.dividerMargin+a.textHeight).attr("fill","white").attr("class","classText");f=!0,e.members.forEach(function(e){p(t,e,f,a),f=!1}),d=t.node().getBBox()}if(e.methods.length>0){l=c.append("line").attr("x1",0).attr("y1",a.padding+y+a.dividerMargin+d.height).attr("y2",a.padding+y+a.dividerMargin+d.height);let t=c.append("text").attr("x",a.padding).attr("y",y+2*a.dividerMargin+d.height+a.textHeight).attr("fill","white").attr("class","classText");f=!0,e.methods.forEach(function(e){p(t,e,f,a),f=!1})}let b=c.node().getBBox();var m=" ";e.cssClasses.length>0&&(m+=e.cssClasses.join(" "));let w=c.insert("rect",":first-child").attr("x",0).attr("y",0).attr("width",b.width+2*a.padding).attr("height",b.height+a.padding+.5*a.dividerMargin).attr("class",m).node().getBBox().width;return i.node().childNodes.forEach(function(t){t.setAttribute("x",(w-t.getBBox().width)/2)}),e.tooltip&&i.insert("title").text(e.tooltip),n&&n.attr("x2",w),l&&l.attr("x2",w),h.width=w,h.height=b.height+a.padding+.5*a.dividerMargin,h},drawEdge:function(t,e,a,r,n){let d,s,p,g,h,c;let f=function(t){switch(t){case n.db.relationType.AGGREGATION:return"aggregation";case n.db.relationType.EXTENSION:return"extension";case n.db.relationType.COMPOSITION:return"composition";case n.db.relationType.DEPENDENCY:return"dependency";case n.db.relationType.LOLLIPOP:return"lollipop"}};e.points=e.points.filter(t=>!Number.isNaN(t.y));let x=e.points,u=(0,i.jvg)().x(function(t){return t.x}).y(function(t){return t.y}).curve(i.$0Z),y=t.append("path").attr("d",u(x)).attr("id","edge"+l).attr("class","relation"),b="";r.arrowMarkerAbsolute&&(b=(b=(b=window.location.protocol+"//"+window.location.host+window.location.pathname+window.location.search).replace(/\(/g,"\\(")).replace(/\)/g,"\\)")),1==a.relation.lineType&&y.attr("class","relation dashed-line"),10==a.relation.lineType&&y.attr("class","relation dotted-line"),"none"!==a.relation.type1&&y.attr("marker-start","url("+b+"#"+f(a.relation.type1)+"Start)"),"none"!==a.relation.type2&&y.attr("marker-end","url("+b+"#"+f(a.relation.type2)+"End)");let m=e.points.length,w=o.u.calcLabelPosition(e.points);if(d=w.x,s=w.y,m%2!=0&&m>1){let t=o.u.calcCardinalityPosition("none"!==a.relation.type1,e.points,e.points[0]),r=o.u.calcCardinalityPosition("none"!==a.relation.type2,e.points,e.points[m-1]);o.l.debug("cardinality_1_point "+JSON.stringify(t)),o.l.debug("cardinality_2_point "+JSON.stringify(r)),p=t.x,g=t.y,h=r.x,c=r.y}if(void 0!==a.title){let e=t.append("g").attr("class","classLabel"),i=e.append("text").attr("class","label").attr("x",d).attr("y",s).attr("fill","red").attr("text-anchor","middle").text(a.title);window.label=i;let n=i.node().getBBox();e.insert("rect",":first-child").attr("class","box").attr("x",n.x-r.padding/2).attr("y",n.y-r.padding/2).attr("width",n.width+r.padding).attr("height",n.height+r.padding)}o.l.info("Rendering relation "+JSON.stringify(a)),void 0!==a.relationTitle1&&"none"!==a.relationTitle1&&t.append("g").attr("class","cardinality").append("text").attr("class","type1").attr("x",p).attr("y",g).attr("fill","black").attr("font-size","6").text(a.relationTitle1),void 0!==a.relationTitle2&&"none"!==a.relationTitle2&&t.append("g").attr("class","cardinality").append("text").attr("class","type2").attr("x",h).attr("y",c).attr("fill","black").attr("font-size","6").text(a.relationTitle2),l++},drawNote:function(t,e,a,r){o.l.debug("Rendering note ",e,a);let i=e.id,n={id:i,text:e.text,width:0,height:0},d=t.append("g").attr("id",i).attr("class","classGroup"),l=d.append("text").attr("y",a.textHeight+a.padding).attr("x",0),s=JSON.parse(`"${e.text}"`).split("\n");s.forEach(function(t){o.l.debug(`Adding line: ${t}`),l.append("tspan").text(t).attr("class","title").attr("dy",a.textHeight)});let p=d.node().getBBox(),g=d.insert("rect",":first-child").attr("x",0).attr("y",0).attr("width",p.width+2*a.padding).attr("height",p.height+s.length*a.textHeight+a.padding+.5*a.dividerMargin).node().getBBox().width;return l.node().childNodes.forEach(function(t){t.setAttribute("x",(g-t.getBBox().width)/2)}),n.width=g,n.height=p.height+s.length*a.textHeight+a.padding+.5*a.dividerMargin,n}},h={},c=function(t){let e=Object.entries(h).find(e=>e[1].label===t);if(e)return e[0]},f=function(t){t.append("defs").append("marker").attr("id","extensionStart").attr("class","extension").attr("refX",0).attr("refY",7).attr("markerWidth",190).attr("markerHeight",240).attr("orient","auto").append("path").attr("d","M 1,7 L18,13 V 1 Z"),t.append("defs").append("marker").attr("id","extensionEnd").attr("refX",19).attr("refY",7).attr("markerWidth",20).attr("markerHeight",28).attr("orient","auto").append("path").attr("d","M 1,1 V 13 L18,7 Z"),t.append("defs").append("marker").attr("id","compositionStart").attr("class","extension").attr("refX",0).attr("refY",7).attr("markerWidth",190).attr("markerHeight",240).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id","compositionEnd").attr("refX",19).attr("refY",7).attr("markerWidth",20).attr("markerHeight",28).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id","aggregationStart").attr("class","extension").attr("refX",0).attr("refY",7).attr("markerWidth",190).attr("markerHeight",240).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id","aggregationEnd").attr("refX",19).attr("refY",7).attr("markerWidth",20).attr("markerHeight",28).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id","dependencyStart").attr("class","extension").attr("refX",0).attr("refY",7).attr("markerWidth",190).attr("markerHeight",240).attr("orient","auto").append("path").attr("d","M 5,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id","dependencyEnd").attr("refX",19).attr("refY",7).attr("markerWidth",20).attr("markerHeight",28).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L14,7 L9,1 Z")},x={parser:r.p,db:r.d,renderer:{draw:function(t,e,a,r){let l;let s=(0,o.c)().class;h={},o.l.info("Rendering diagram "+t);let p=(0,o.c)().securityLevel;"sandbox"===p&&(l=(0,i.Ys)("#i"+e));let x="sandbox"===p?(0,i.Ys)(l.nodes()[0].contentDocument.body):(0,i.Ys)("body"),u=x.select(`[id='${e}']`);f(u);let y=new d.k({multigraph:!0});y.setGraph({isMultiGraph:!0}),y.setDefaultEdgeLabel(function(){return{}});let b=r.db.getClasses();for(let t of Object.keys(b)){let e=b[t],a=g.drawClass(u,e,s,r);h[a.id]=a,y.setNode(a.id,a),o.l.info("Org height: "+a.height)}r.db.getRelations().forEach(function(t){o.l.info("tjoho"+c(t.id1)+c(t.id2)+JSON.stringify(t)),y.setEdge(c(t.id1),c(t.id2),{relation:t},t.title||"DEFAULT")}),r.db.getNotes().forEach(function(t){o.l.debug(`Adding note: ${JSON.stringify(t)}`);let e=g.drawNote(u,t,s,r);h[e.id]=e,y.setNode(e.id,e),t.class&&t.class in b&&y.setEdge(t.id,c(t.class),{relation:{id1:t.id,id2:t.class,relation:{type1:"none",type2:"none",lineType:10}}},"DEFAULT")}),(0,n.bK)(y),y.nodes().forEach(function(t){void 0!==t&&void 0!==y.node(t)&&(o.l.debug("Node "+t+": "+JSON.stringify(y.node(t))),x.select("#"+(r.db.lookUpDomId(t)||t)).attr("transform","translate("+(y.node(t).x-y.node(t).width/2)+","+(y.node(t).y-y.node(t).height/2)+" )"))}),y.edges().forEach(function(t){void 0!==t&&void 0!==y.edge(t)&&(o.l.debug("Edge "+t.v+" -> "+t.w+": "+JSON.stringify(y.edge(t))),g.drawEdge(u,y.edge(t),y.edge(t).relation,s,r))});let m=u.node().getBBox(),w=m.width+40,k=m.height+40;(0,o.i)(u,k,w,s.useMaxWidth);let E=`${m.x-20} ${m.y-20} ${w} ${k}`;o.l.debug(`viewBox ${E}`),u.attr("viewBox",E)}},styles:r.s,init:t=>{t.class||(t.class={}),t.class.arrowMarkerAbsolute=t.arrowMarkerAbsolute,r.d.clear()}}}}]);