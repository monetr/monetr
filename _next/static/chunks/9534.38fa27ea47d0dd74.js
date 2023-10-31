(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[9534],{9580:function(t){t.exports=function(t,e){var i=e.prototype,n=i.format;i.format=function(t){var e=this,i=this.$locale();if(!this.isValid())return n.bind(this)(t);var r=this.$utils(),l=(t||"YYYY-MM-DDTHH:mm:ssZ").replace(/\[([^\]]+)]|Q|wo|ww|w|WW|W|zzz|z|gggg|GGGG|Do|X|x|k{1,2}|S/g,function(t){switch(t){case"Q":return Math.ceil((e.$M+1)/3);case"Do":return i.ordinal(e.$D);case"gggg":return e.weekYear();case"GGGG":return e.isoWeekYear();case"wo":return i.ordinal(e.week(),"W");case"w":case"ww":return r.s(e.week(),"w"===t?1:2,"0");case"W":case"WW":return r.s(e.isoWeek(),"W"===t?1:2,"0");case"k":case"kk":return r.s(String(0===e.$H?24:e.$H),"k"===t?1:2,"0");case"X":return Math.floor(e.$d.getTime()/1e3);case"x":return e.$d.getTime();case"z":return"["+e.offsetName()+"]";case"zzz":return"["+e.offsetName("long")+"]";default:return t}});return n.bind(this)(l)}}},9746:function(t){t.exports=function(){"use strict";var t={LTS:"h:mm:ss A",LT:"h:mm A",L:"MM/DD/YYYY",LL:"MMMM D, YYYY",LLL:"MMMM D, YYYY h:mm A",LLLL:"dddd, MMMM D, YYYY h:mm A"},e=/(\[[^[]*\])|([-_:/.,()\s]+)|(A|a|YYYY|YY?|MM?M?M?|Do|DD?|hh?|HH?|mm?|ss?|S{1,3}|z|ZZ?)/g,i=/\d\d/,n=/\d\d?/,r=/\d*[^-_:/,()\s\d]+/,l={},s=function(t){return(t=+t)+(t>68?1900:2e3)},a=function(t){return function(e){this[t]=+e}},d=[/[+-]\d\d:?(\d\d)?|Z/,function(t){(this.zone||(this.zone={})).offset=function(t){if(!t||"Z"===t)return 0;var e=t.match(/([+-]|\d\d)/g),i=60*e[1]+(+e[2]||0);return 0===i?0:"+"===e[0]?-i:i}(t)}],h=function(t){var e=l[t];return e&&(e.indexOf?e:e.s.concat(e.f))},u=function(t,e){var i,n=l.meridiem;if(n){for(var r=1;r<=24;r+=1)if(t.indexOf(n(r,0,e))>-1){i=r>12;break}}else i=t===(e?"pm":"PM");return i},f={A:[r,function(t){this.afternoon=u(t,!1)}],a:[r,function(t){this.afternoon=u(t,!0)}],S:[/\d/,function(t){this.milliseconds=100*+t}],SS:[i,function(t){this.milliseconds=10*+t}],SSS:[/\d{3}/,function(t){this.milliseconds=+t}],s:[n,a("seconds")],ss:[n,a("seconds")],m:[n,a("minutes")],mm:[n,a("minutes")],H:[n,a("hours")],h:[n,a("hours")],HH:[n,a("hours")],hh:[n,a("hours")],D:[n,a("day")],DD:[i,a("day")],Do:[r,function(t){var e=l.ordinal,i=t.match(/\d+/);if(this.day=i[0],e)for(var n=1;n<=31;n+=1)e(n).replace(/\[|\]/g,"")===t&&(this.day=n)}],M:[n,a("month")],MM:[i,a("month")],MMM:[r,function(t){var e=h("months"),i=(h("monthsShort")||e.map(function(t){return t.slice(0,3)})).indexOf(t)+1;if(i<1)throw Error();this.month=i%12||i}],MMMM:[r,function(t){var e=h("months").indexOf(t)+1;if(e<1)throw Error();this.month=e%12||e}],Y:[/[+-]?\d+/,a("year")],YY:[i,function(t){this.year=s(t)}],YYYY:[/\d{4}/,a("year")],Z:d,ZZ:d};function c(i){var n,r;n=i,r=l&&l.formats;for(var d=(i=n.replace(/(\[[^\]]+])|(LTS?|l{1,4}|L{1,4})/g,function(e,i,n){var l=n&&n.toUpperCase();return i||r[n]||t[n]||r[l].replace(/(\[[^\]]+])|(MMMM|MM|DD|dddd)/g,function(t,e,i){return e||i.slice(1)})})).match(e),y=d.length,m=0;m<y;m+=1){var k=d[m],p=f[k],g=p&&p[0],b=p&&p[1];d[m]=b?{regex:g,parser:b}:k.replace(/^\[|\]$/g,"")}return function(t){for(var e={},i=0,n=0;i<y;i+=1){var r=d[i];if("string"==typeof r)n+=r.length;else{var l=r.regex,f=r.parser,m=t.slice(n),k=l.exec(m)[0];f.call(e,k),t=t.replace(k,"")}}return function(t){var e=t.afternoon;if(void 0!==e){var i=t.hours;e?i<12&&(t.hours+=12):12===i&&(t.hours=0),delete t.afternoon}}(e),e}}return function(t,e,i){i.p.customParseFormat=!0,t&&t.parseTwoDigitYear&&(s=t.parseTwoDigitYear);var n=e.prototype,r=n.parse;n.parse=function(t){var e=t.date,n=t.utc,d=t.args;this.$u=n;var f=d[1];if("string"==typeof f){var y=!0===d[2],m=!0===d[3],k=d[2];m&&(k=d[2]),l=this.$locale(),!y&&k&&(l=i.Ls[k]),this.$d=function(t,e,i){try{if(["x","X"].indexOf(e)>-1)return new Date(("X"===e?1e3:1)*t);var n=c(e)(t),r=n.year,l=n.month,d=n.day,f=n.hours,y=n.minutes,m=n.seconds,k=n.milliseconds,p=n.zone,g=new Date,b=d||(r||l?1:g.getDate()),T=r||g.getFullYear(),x=0;r&&!l||(x=l>0?l-1:g.getMonth());var v=f||0,_=y||0,w=m||0,D=k||0;return p?new Date(Date.UTC(T,x,b,v,_,w,D+60*p.offset*1e3)):i?new Date(Date.UTC(T,x,b,v,_,w,D)):new Date(T,x,b,v,_,w,D)}catch(t){return new Date("")}}(e,f,n),this.init(),k&&!0!==k&&(this.$L=this.locale(k).$L),(y||m)&&e!=this.format(f)&&(this.$d=new Date("")),l={}}else if(f instanceof Array)for(var p=f.length,g=1;g<=p;g+=1){d[1]=f[g-1];var b=i.apply(this,d);if(b.isValid()){this.$d=b.$d,this.$L=b.$L,this.init();break}g===p&&(this.$d=new Date(""))}else r.call(this,t)}}}()},7635:function(t){t.exports=function(t,e,i){var a=function(t){return t.add(4-t.isoWeekday(),"day")},n=e.prototype;n.isoWeekYear=function(){return a(this).year()},n.isoWeek=function(t){if(!this.$utils().u(t))return this.add(7*(t-this.isoWeek()),"day");var e,n,r,l=a(this),d=(e=this.isoWeekYear(),r=4-(n=(this.$u?i.utc:i)().year(e).startOf("year")).isoWeekday(),n.isoWeekday()>4&&(r+=7),n.add(r,"day"));return l.diff(d,"week")+1},n.isoWeekday=function(t){return this.$utils().u(t)?this.day()||7:this.day(this.day()%7?t:t-7)};var r=n.startOf;n.startOf=function(t,e){var i=this.$utils(),n=!!i.u(e)||e;return"isoweek"===i.p(t)?n?this.date(this.date()-(this.isoWeekday()-1)).startOf("day"):this.date(this.date()-1-(this.isoWeekday()-1)+7).endOf("day"):r.bind(this)(t,e)}}},9534:function(t,e,i){"use strict";let n,r,l,d;i.d(e,{diagram:function(){return G}});var f=i(7608),y=i(7693),m=i(7635),k=i(9746),p=i(9580),g=i(6388),b=i(6357);i(1699);var T=function(){var o=function(t,e,i,n){for(i=i||{},n=t.length;n--;i[t[n]]=e);return i},t=[6,8,10,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,30,32,33,35,37],e=[1,25],i=[1,26],n=[1,27],r=[1,28],l=[1,29],d=[1,30],f=[1,31],y=[1,9],m=[1,10],k=[1,11],p=[1,12],g=[1,13],b=[1,14],T=[1,15],x=[1,16],v=[1,18],_=[1,19],w=[1,20],D=[1,21],$=[1,22],C=[1,24],S=[1,32],E={trace:function(){},yy:{},symbols_:{error:2,start:3,gantt:4,document:5,EOF:6,line:7,SPACE:8,statement:9,NL:10,weekday:11,weekday_monday:12,weekday_tuesday:13,weekday_wednesday:14,weekday_thursday:15,weekday_friday:16,weekday_saturday:17,weekday_sunday:18,dateFormat:19,inclusiveEndDates:20,topAxis:21,axisFormat:22,tickInterval:23,excludes:24,includes:25,todayMarker:26,title:27,acc_title:28,acc_title_value:29,acc_descr:30,acc_descr_value:31,acc_descr_multiline_value:32,section:33,clickStatement:34,taskTxt:35,taskData:36,click:37,callbackname:38,callbackargs:39,href:40,clickStatementDebug:41,$accept:0,$end:1},terminals_:{2:"error",4:"gantt",6:"EOF",8:"SPACE",10:"NL",12:"weekday_monday",13:"weekday_tuesday",14:"weekday_wednesday",15:"weekday_thursday",16:"weekday_friday",17:"weekday_saturday",18:"weekday_sunday",19:"dateFormat",20:"inclusiveEndDates",21:"topAxis",22:"axisFormat",23:"tickInterval",24:"excludes",25:"includes",26:"todayMarker",27:"title",28:"acc_title",29:"acc_title_value",30:"acc_descr",31:"acc_descr_value",32:"acc_descr_multiline_value",33:"section",35:"taskTxt",36:"taskData",37:"click",38:"callbackname",39:"callbackargs",40:"href"},productions_:[0,[3,3],[5,0],[5,2],[7,2],[7,1],[7,1],[7,1],[11,1],[11,1],[11,1],[11,1],[11,1],[11,1],[11,1],[9,1],[9,1],[9,1],[9,1],[9,1],[9,1],[9,1],[9,1],[9,1],[9,1],[9,2],[9,2],[9,1],[9,1],[9,1],[9,2],[34,2],[34,3],[34,3],[34,4],[34,3],[34,4],[34,2],[41,2],[41,3],[41,3],[41,4],[41,3],[41,4],[41,2]],performAction:function(t,e,i,n,r,l,d){var f=l.length-1;switch(r){case 1:return l[f-1];case 2:case 6:case 7:this.$=[];break;case 3:l[f-1].push(l[f]),this.$=l[f-1];break;case 4:case 5:this.$=l[f];break;case 8:n.setWeekday("monday");break;case 9:n.setWeekday("tuesday");break;case 10:n.setWeekday("wednesday");break;case 11:n.setWeekday("thursday");break;case 12:n.setWeekday("friday");break;case 13:n.setWeekday("saturday");break;case 14:n.setWeekday("sunday");break;case 15:n.setDateFormat(l[f].substr(11)),this.$=l[f].substr(11);break;case 16:n.enableInclusiveEndDates(),this.$=l[f].substr(18);break;case 17:n.TopAxis(),this.$=l[f].substr(8);break;case 18:n.setAxisFormat(l[f].substr(11)),this.$=l[f].substr(11);break;case 19:n.setTickInterval(l[f].substr(13)),this.$=l[f].substr(13);break;case 20:n.setExcludes(l[f].substr(9)),this.$=l[f].substr(9);break;case 21:n.setIncludes(l[f].substr(9)),this.$=l[f].substr(9);break;case 22:n.setTodayMarker(l[f].substr(12)),this.$=l[f].substr(12);break;case 24:n.setDiagramTitle(l[f].substr(6)),this.$=l[f].substr(6);break;case 25:this.$=l[f].trim(),n.setAccTitle(this.$);break;case 26:case 27:this.$=l[f].trim(),n.setAccDescription(this.$);break;case 28:n.addSection(l[f].substr(8)),this.$=l[f].substr(8);break;case 30:n.addTask(l[f-1],l[f]),this.$="task";break;case 31:this.$=l[f-1],n.setClickEvent(l[f-1],l[f],null);break;case 32:this.$=l[f-2],n.setClickEvent(l[f-2],l[f-1],l[f]);break;case 33:this.$=l[f-2],n.setClickEvent(l[f-2],l[f-1],null),n.setLink(l[f-2],l[f]);break;case 34:this.$=l[f-3],n.setClickEvent(l[f-3],l[f-2],l[f-1]),n.setLink(l[f-3],l[f]);break;case 35:this.$=l[f-2],n.setClickEvent(l[f-2],l[f],null),n.setLink(l[f-2],l[f-1]);break;case 36:this.$=l[f-3],n.setClickEvent(l[f-3],l[f-1],l[f]),n.setLink(l[f-3],l[f-2]);break;case 37:this.$=l[f-1],n.setLink(l[f-1],l[f]);break;case 38:case 44:this.$=l[f-1]+" "+l[f];break;case 39:case 40:case 42:this.$=l[f-2]+" "+l[f-1]+" "+l[f];break;case 41:case 43:this.$=l[f-3]+" "+l[f-2]+" "+l[f-1]+" "+l[f]}},table:[{3:1,4:[1,2]},{1:[3]},o(t,[2,2],{5:3}),{6:[1,4],7:5,8:[1,6],9:7,10:[1,8],11:17,12:e,13:i,14:n,15:r,16:l,17:d,18:f,19:y,20:m,21:k,22:p,23:g,24:b,25:T,26:x,27:v,28:_,30:w,32:D,33:$,34:23,35:C,37:S},o(t,[2,7],{1:[2,1]}),o(t,[2,3]),{9:33,11:17,12:e,13:i,14:n,15:r,16:l,17:d,18:f,19:y,20:m,21:k,22:p,23:g,24:b,25:T,26:x,27:v,28:_,30:w,32:D,33:$,34:23,35:C,37:S},o(t,[2,5]),o(t,[2,6]),o(t,[2,15]),o(t,[2,16]),o(t,[2,17]),o(t,[2,18]),o(t,[2,19]),o(t,[2,20]),o(t,[2,21]),o(t,[2,22]),o(t,[2,23]),o(t,[2,24]),{29:[1,34]},{31:[1,35]},o(t,[2,27]),o(t,[2,28]),o(t,[2,29]),{36:[1,36]},o(t,[2,8]),o(t,[2,9]),o(t,[2,10]),o(t,[2,11]),o(t,[2,12]),o(t,[2,13]),o(t,[2,14]),{38:[1,37],40:[1,38]},o(t,[2,4]),o(t,[2,25]),o(t,[2,26]),o(t,[2,30]),o(t,[2,31],{39:[1,39],40:[1,40]}),o(t,[2,37],{38:[1,41]}),o(t,[2,32],{40:[1,42]}),o(t,[2,33]),o(t,[2,35],{39:[1,43]}),o(t,[2,34]),o(t,[2,36])],defaultActions:{},parseError:function(t,e){if(e.recoverable)this.trace(t);else{var i=Error(t);throw i.hash=e,i}},parse:function(t){var e=this,i=[0],n=[],r=[null],l=[],d=this.table,f="",y=0,m=0,k=l.slice.call(arguments,1),p=Object.create(this.lexer),g={yy:{}};for(var b in this.yy)Object.prototype.hasOwnProperty.call(this.yy,b)&&(g.yy[b]=this.yy[b]);p.setInput(t,g.yy),g.yy.lexer=p,g.yy.parser=this,void 0===p.yylloc&&(p.yylloc={});var T=p.yylloc;l.push(T);var x=p.options&&p.options.ranges;function lex(){var t;return"number"!=typeof(t=n.pop()||p.lex()||1)&&(t instanceof Array&&(t=(n=t).pop()),t=e.symbols_[t]||t),t}"function"==typeof g.yy.parseError?this.parseError=g.yy.parseError:this.parseError=Object.getPrototypeOf(this).parseError;for(var v,_,w,D,$,C,S,E,M={};;){if(_=i[i.length-1],this.defaultActions[_]?w=this.defaultActions[_]:(null==v&&(v=lex()),w=d[_]&&d[_][v]),void 0===w||!w.length||!w[0]){var Y="";for($ in E=[],d[_])this.terminals_[$]&&$>2&&E.push("'"+this.terminals_[$]+"'");Y=p.showPosition?"Parse error on line "+(y+1)+":\n"+p.showPosition()+"\nExpecting "+E.join(", ")+", got '"+(this.terminals_[v]||v)+"'":"Parse error on line "+(y+1)+": Unexpected "+(1==v?"end of input":"'"+(this.terminals_[v]||v)+"'"),this.parseError(Y,{text:p.match,token:this.terminals_[v]||v,line:p.yylineno,loc:T,expected:E})}if(w[0]instanceof Array&&w.length>1)throw Error("Parse Error: multiple actions possible at state: "+_+", token: "+v);switch(w[0]){case 1:i.push(v),r.push(p.yytext),l.push(p.yylloc),i.push(w[1]),v=null,m=p.yyleng,f=p.yytext,y=p.yylineno,T=p.yylloc;break;case 2:if(C=this.productions_[w[1]][1],M.$=r[r.length-C],M._$={first_line:l[l.length-(C||1)].first_line,last_line:l[l.length-1].last_line,first_column:l[l.length-(C||1)].first_column,last_column:l[l.length-1].last_column},x&&(M._$.range=[l[l.length-(C||1)].range[0],l[l.length-1].range[1]]),void 0!==(D=this.performAction.apply(M,[f,m,y,g.yy,w[1],r,l].concat(k))))return D;C&&(i=i.slice(0,-1*C*2),r=r.slice(0,-1*C),l=l.slice(0,-1*C)),i.push(this.productions_[w[1]][0]),r.push(M.$),l.push(M._$),S=d[i[i.length-2]][i[i.length-1]],i.push(S);break;case 3:return!0}}return!0}};function Parser(){this.yy={}}return E.lexer={EOF:1,parseError:function(t,e){if(this.yy.parser)this.yy.parser.parseError(t,e);else throw Error(t)},setInput:function(t,e){return this.yy=e||this.yy||{},this._input=t,this._more=this._backtrack=this.done=!1,this.yylineno=this.yyleng=0,this.yytext=this.matched=this.match="",this.conditionStack=["INITIAL"],this.yylloc={first_line:1,first_column:0,last_line:1,last_column:0},this.options.ranges&&(this.yylloc.range=[0,0]),this.offset=0,this},input:function(){var t=this._input[0];return this.yytext+=t,this.yyleng++,this.offset++,this.match+=t,this.matched+=t,t.match(/(?:\r\n?|\n).*/g)?(this.yylineno++,this.yylloc.last_line++):this.yylloc.last_column++,this.options.ranges&&this.yylloc.range[1]++,this._input=this._input.slice(1),t},unput:function(t){var e=t.length,i=t.split(/(?:\r\n?|\n)/g);this._input=t+this._input,this.yytext=this.yytext.substr(0,this.yytext.length-e),this.offset-=e;var n=this.match.split(/(?:\r\n?|\n)/g);this.match=this.match.substr(0,this.match.length-1),this.matched=this.matched.substr(0,this.matched.length-1),i.length-1&&(this.yylineno-=i.length-1);var r=this.yylloc.range;return this.yylloc={first_line:this.yylloc.first_line,last_line:this.yylineno+1,first_column:this.yylloc.first_column,last_column:i?(i.length===n.length?this.yylloc.first_column:0)+n[n.length-i.length].length-i[0].length:this.yylloc.first_column-e},this.options.ranges&&(this.yylloc.range=[r[0],r[0]+this.yyleng-e]),this.yyleng=this.yytext.length,this},more:function(){return this._more=!0,this},reject:function(){return this.options.backtrack_lexer?(this._backtrack=!0,this):this.parseError("Lexical error on line "+(this.yylineno+1)+". You can only invoke reject() in the lexer when the lexer is of the backtracking persuasion (options.backtrack_lexer = true).\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},less:function(t){this.unput(this.match.slice(t))},pastInput:function(){var t=this.matched.substr(0,this.matched.length-this.match.length);return(t.length>20?"...":"")+t.substr(-20).replace(/\n/g,"")},upcomingInput:function(){var t=this.match;return t.length<20&&(t+=this._input.substr(0,20-t.length)),(t.substr(0,20)+(t.length>20?"...":"")).replace(/\n/g,"")},showPosition:function(){var t=this.pastInput(),e=Array(t.length+1).join("-");return t+this.upcomingInput()+"\n"+e+"^"},test_match:function(t,e){var i,n,r;if(this.options.backtrack_lexer&&(r={yylineno:this.yylineno,yylloc:{first_line:this.yylloc.first_line,last_line:this.last_line,first_column:this.yylloc.first_column,last_column:this.yylloc.last_column},yytext:this.yytext,match:this.match,matches:this.matches,matched:this.matched,yyleng:this.yyleng,offset:this.offset,_more:this._more,_input:this._input,yy:this.yy,conditionStack:this.conditionStack.slice(0),done:this.done},this.options.ranges&&(r.yylloc.range=this.yylloc.range.slice(0))),(n=t[0].match(/(?:\r\n?|\n).*/g))&&(this.yylineno+=n.length),this.yylloc={first_line:this.yylloc.last_line,last_line:this.yylineno+1,first_column:this.yylloc.last_column,last_column:n?n[n.length-1].length-n[n.length-1].match(/\r?\n?/)[0].length:this.yylloc.last_column+t[0].length},this.yytext+=t[0],this.match+=t[0],this.matches=t,this.yyleng=this.yytext.length,this.options.ranges&&(this.yylloc.range=[this.offset,this.offset+=this.yyleng]),this._more=!1,this._backtrack=!1,this._input=this._input.slice(t[0].length),this.matched+=t[0],i=this.performAction.call(this,this.yy,this,e,this.conditionStack[this.conditionStack.length-1]),this.done&&this._input&&(this.done=!1),i)return i;if(this._backtrack)for(var l in r)this[l]=r[l];return!1},next:function(){if(this.done)return this.EOF;this._input||(this.done=!0),this._more||(this.yytext="",this.match="");for(var t,e,i,n,r=this._currentRules(),l=0;l<r.length;l++)if((i=this._input.match(this.rules[r[l]]))&&(!e||i[0].length>e[0].length)){if(e=i,n=l,this.options.backtrack_lexer){if(!1!==(t=this.test_match(i,r[l])))return t;if(!this._backtrack)return!1;e=!1;continue}if(!this.options.flex)break}return e?!1!==(t=this.test_match(e,r[n]))&&t:""===this._input?this.EOF:this.parseError("Lexical error on line "+(this.yylineno+1)+". Unrecognized text.\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},lex:function(){return this.next()||this.lex()},begin:function(t){this.conditionStack.push(t)},popState:function(){return this.conditionStack.length-1>0?this.conditionStack.pop():this.conditionStack[0]},_currentRules:function(){return this.conditionStack.length&&this.conditionStack[this.conditionStack.length-1]?this.conditions[this.conditionStack[this.conditionStack.length-1]].rules:this.conditions.INITIAL.rules},topState:function(t){return(t=this.conditionStack.length-1-Math.abs(t||0))>=0?this.conditionStack[t]:"INITIAL"},pushState:function(t){this.begin(t)},stateStackSize:function(){return this.conditionStack.length},options:{"case-insensitive":!0},performAction:function(t,e,i,n){switch(i){case 0:return this.begin("open_directive"),"open_directive";case 1:return this.begin("acc_title"),28;case 2:return this.popState(),"acc_title_value";case 3:return this.begin("acc_descr"),30;case 4:return this.popState(),"acc_descr_value";case 5:this.begin("acc_descr_multiline");break;case 6:case 16:case 19:case 22:case 25:this.popState();break;case 7:return"acc_descr_multiline_value";case 8:case 9:case 10:case 12:case 13:case 14:break;case 11:return 10;case 15:this.begin("href");break;case 17:return 40;case 18:this.begin("callbackname");break;case 20:this.popState(),this.begin("callbackargs");break;case 21:return 38;case 23:return 39;case 24:this.begin("click");break;case 26:return 37;case 27:return 4;case 28:return 19;case 29:return 20;case 30:return 21;case 31:return 22;case 32:return 23;case 33:return 25;case 34:return 24;case 35:return 26;case 36:return 12;case 37:return 13;case 38:return 14;case 39:return 15;case 40:return 16;case 41:return 17;case 42:return 18;case 43:return"date";case 44:return 27;case 45:return"accDescription";case 46:return 33;case 47:return 35;case 48:return 36;case 49:return":";case 50:return 6;case 51:return"INVALID"}},rules:[/^(?:%%\{)/i,/^(?:accTitle\s*:\s*)/i,/^(?:(?!\n||)*[^\n]*)/i,/^(?:accDescr\s*:\s*)/i,/^(?:(?!\n||)*[^\n]*)/i,/^(?:accDescr\s*\{\s*)/i,/^(?:[\}])/i,/^(?:[^\}]*)/i,/^(?:%%(?!\{)*[^\n]*)/i,/^(?:[^\}]%%*[^\n]*)/i,/^(?:%%*[^\n]*[\n]*)/i,/^(?:[\n]+)/i,/^(?:\s+)/i,/^(?:#[^\n]*)/i,/^(?:%[^\n]*)/i,/^(?:href[\s]+["])/i,/^(?:["])/i,/^(?:[^"]*)/i,/^(?:call[\s]+)/i,/^(?:\([\s]*\))/i,/^(?:\()/i,/^(?:[^(]*)/i,/^(?:\))/i,/^(?:[^)]*)/i,/^(?:click[\s]+)/i,/^(?:[\s\n])/i,/^(?:[^\s\n]*)/i,/^(?:gantt\b)/i,/^(?:dateFormat\s[^#\n;]+)/i,/^(?:inclusiveEndDates\b)/i,/^(?:topAxis\b)/i,/^(?:axisFormat\s[^#\n;]+)/i,/^(?:tickInterval\s[^#\n;]+)/i,/^(?:includes\s[^#\n;]+)/i,/^(?:excludes\s[^#\n;]+)/i,/^(?:todayMarker\s[^\n;]+)/i,/^(?:weekday\s+monday\b)/i,/^(?:weekday\s+tuesday\b)/i,/^(?:weekday\s+wednesday\b)/i,/^(?:weekday\s+thursday\b)/i,/^(?:weekday\s+friday\b)/i,/^(?:weekday\s+saturday\b)/i,/^(?:weekday\s+sunday\b)/i,/^(?:\d\d\d\d-\d\d-\d\d\b)/i,/^(?:title\s[^#\n;]+)/i,/^(?:accDescription\s[^#\n;]+)/i,/^(?:section\s[^#:\n;]+)/i,/^(?:[^#:\n;]+)/i,/^(?::[^#\n;]+)/i,/^(?::)/i,/^(?:$)/i,/^(?:.)/i],conditions:{acc_descr_multiline:{rules:[6,7],inclusive:!1},acc_descr:{rules:[4],inclusive:!1},acc_title:{rules:[2],inclusive:!1},callbackargs:{rules:[22,23],inclusive:!1},callbackname:{rules:[19,20,21],inclusive:!1},href:{rules:[16,17],inclusive:!1},click:{rules:[25,26],inclusive:!1},INITIAL:{rules:[0,1,3,5,8,9,10,11,12,13,14,15,18,24,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51],inclusive:!0}}},Parser.prototype=E,E.Parser=Parser,new Parser}();T.parser=T,y.extend(m),y.extend(k),y.extend(p);let x="",v="",_="",w=[],D=[],$={},C=[],S=[],E="",M="",Y=["active","done","crit","milestone"],A=[],I=!1,L=!1,F="sunday",O=0,isInvalidDate=function(t,e,i,n){return!n.includes(t.format(e.trim()))&&(!!(t.isoWeekday()>=6&&i.includes("weekends")||i.includes(t.format("dddd").toLowerCase()))||i.includes(t.format(e.trim())))},checkTaskDates=function(t,e,i,n){let r,l;if(!i.length||t.manualEndTime)return;r=(r=t.startTime instanceof Date?y(t.startTime):y(t.startTime,e,!0)).add(1,"d"),l=t.endTime instanceof Date?y(t.endTime):y(t.endTime,e,!0);let[d,f]=fixTaskDates(r,l,e,i,n);t.endTime=d.toDate(),t.renderEndTime=f},fixTaskDates=function(t,e,i,n,r){let l=!1,d=null;for(;t<=e;)l||(d=e.toDate()),(l=isInvalidDate(t,i,n,r))&&(e=e.add(1,"d")),t=t.add(1,"d");return[e,d]},getStartDate=function(t,e,i){i=i.trim();let n=/^after\s+([\d\w- ]+)/.exec(i.trim());if(null!==n){let t=null;if(n[1].split(" ").forEach(function(e){let i=findTaskById(e);void 0!==i&&(t?i.endTime>t.endTime&&(t=i):t=i)}),t)return t.endTime;{let t=new Date;return t.setHours(0,0,0,0),t}}let r=y(i,e.trim(),!0);if(r.isValid())return r.toDate();{g.l.debug("Invalid date:"+i),g.l.debug("With date format:"+e.trim());let t=new Date(i);if(void 0===t||isNaN(t.getTime())||-1e4>t.getFullYear()||t.getFullYear()>1e4)throw Error("Invalid date:"+i);return t}},parseDuration=function(t){let e=/^(\d+(?:\.\d+)?)([Mdhmswy]|ms)$/.exec(t.trim());return null!==e?[Number.parseFloat(e[1]),e[2]]:[NaN,"ms"]},getEndDate=function(t,e,i,n=!1){let r=y(i=i.trim(),e.trim(),!0);if(r.isValid())return n&&(r=r.add(1,"d")),r.toDate();let l=y(t),[d,f]=parseDuration(i);if(!Number.isNaN(d)){let t=l.add(d,f);t.isValid()&&(l=t)}return l.toDate()},W=0,parseId=function(t){return void 0===t?"task"+(W+=1):t},compileData=function(t,e){let i;i=":"===e.substr(0,1)?e.substr(1,e.length):e;let n=i.split(","),r={};getTaskTags(n,r,Y);for(let t=0;t<n.length;t++)n[t]=n[t].trim();let l="";switch(n.length){case 1:r.id=parseId(),r.startTime=t.endTime,l=n[0];break;case 2:r.id=parseId(),r.startTime=getStartDate(void 0,x,n[0]),l=n[1];break;case 3:r.id=parseId(n[0]),r.startTime=getStartDate(void 0,x,n[1]),l=n[2]}return l&&(r.endTime=getEndDate(r.startTime,x,l,I),r.manualEndTime=y(l,"YYYY-MM-DD",!0).isValid(),checkTaskDates(r,x,D,w)),r},parseData=function(t,e){let i;i=":"===e.substr(0,1)?e.substr(1,e.length):e;let n=i.split(","),r={};getTaskTags(n,r,Y);for(let t=0;t<n.length;t++)n[t]=n[t].trim();switch(n.length){case 1:r.id=parseId(),r.startTime={type:"prevTaskEnd",id:t},r.endTime={data:n[0]};break;case 2:r.id=parseId(),r.startTime={type:"getStartDate",startData:n[0]},r.endTime={data:n[1]};break;case 3:r.id=parseId(n[0]),r.startTime={type:"getStartDate",startData:n[1]},r.endTime={data:n[2]}}return r},z=[],B={},findTaskById=function(t){let e=B[t];return z[e]},compileTasks=function(){let t=!0;for(let[e,i]of z.entries())!function(t){let e=z[t],i="";switch(z[t].raw.startTime.type){case"prevTaskEnd":{let t=findTaskById(e.prevTaskId);e.startTime=t.endTime;break}case"getStartDate":(i=getStartDate(void 0,x,z[t].raw.startTime.startData))&&(z[t].startTime=i)}z[t].startTime&&(z[t].endTime=getEndDate(z[t].startTime,x,z[t].raw.endTime.data,I),z[t].endTime&&(z[t].processed=!0,z[t].manualEndTime=y(z[t].raw.endTime.data,"YYYY-MM-DD",!0).isValid(),checkTaskDates(z[t],x,D,w))),z[t].processed}(e),t=t&&i.processed;return t},setClass=function(t,e){t.split(",").forEach(function(t){let i=findTaskById(t);void 0!==i&&i.classes.push(e)})},setClickFun=function(t,e,i){if("loose"!==(0,g.c)().securityLevel||void 0===e)return;let n=[];if("string"==typeof i){n=i.split(/,(?=(?:(?:[^"]*"){2})*[^"]*$)/);for(let t=0;t<n.length;t++){let e=n[t].trim();'"'===e.charAt(0)&&'"'===e.charAt(e.length-1)&&(e=e.substr(1,e.length-2)),n[t]=e}}0===n.length&&n.push(t),void 0!==findTaskById(t)&&pushFun(t,()=>{g.u.runFunc(e,...n)})},pushFun=function(t,e){A.push(function(){let i=document.querySelector(`[id="${t}"]`);null!==i&&i.addEventListener("click",function(){e()})},function(){let i=document.querySelector(`[id="${t}-text"]`);null!==i&&i.addEventListener("click",function(){e()})})},P={getConfig:()=>(0,g.c)().gantt,clear:function(){C=[],S=[],E="",A=[],W=0,n=void 0,r=void 0,z=[],x="",v="",M="",d=void 0,_="",w=[],D=[],I=!1,L=!1,O=0,$={},(0,g.t)(),F="sunday"},setDateFormat:function(t){x=t},getDateFormat:function(){return x},enableInclusiveEndDates:function(){I=!0},endDatesAreInclusive:function(){return I},enableTopAxis:function(){L=!0},topAxisEnabled:function(){return L},setAxisFormat:function(t){v=t},getAxisFormat:function(){return v},setTickInterval:function(t){d=t},getTickInterval:function(){return d},setTodayMarker:function(t){_=t},getTodayMarker:function(){return _},setAccTitle:g.s,getAccTitle:g.g,setDiagramTitle:g.q,getDiagramTitle:g.r,setDisplayMode:function(t){M=t},getDisplayMode:function(){return M},setAccDescription:g.b,getAccDescription:g.a,addSection:function(t){E=t,C.push(t)},getSections:function(){return C},getTasks:function(){let t=compileTasks(),e=0;for(;!t&&e<10;)t=compileTasks(),e++;return S=z},addTask:function(t,e){let i={section:E,type:E,processed:!1,manualEndTime:!1,renderEndTime:null,raw:{data:e},task:t,classes:[]},n=parseData(r,e);i.raw.startTime=n.startTime,i.raw.endTime=n.endTime,i.id=n.id,i.prevTaskId=r,i.active=n.active,i.done=n.done,i.crit=n.crit,i.milestone=n.milestone,i.order=O,O++;let l=z.push(i);r=i.id,B[i.id]=l-1},findTaskById,addTaskOrg:function(t,e){let i={section:E,type:E,description:t,task:t,classes:[]},r=compileData(n,e);i.startTime=r.startTime,i.endTime=r.endTime,i.id=r.id,i.active=r.active,i.done=r.done,i.crit=r.crit,i.milestone=r.milestone,n=i,S.push(i)},setIncludes:function(t){w=t.toLowerCase().split(/[\s,]+/)},getIncludes:function(){return w},setExcludes:function(t){D=t.toLowerCase().split(/[\s,]+/)},getExcludes:function(){return D},setClickEvent:function(t,e,i){t.split(",").forEach(function(t){setClickFun(t,e,i)}),setClass(t,"clickable")},setLink:function(t,e){let i=e;"loose"!==(0,g.c)().securityLevel&&(i=(0,f.Nm)(e)),t.split(",").forEach(function(t){void 0!==findTaskById(t)&&(pushFun(t,()=>{window.open(i,"_self")}),$[t]=i)}),setClass(t,"clickable")},getLinks:function(){return $},bindFunctions:function(t){A.forEach(function(e){e(t)})},parseDuration,isInvalidDate,setWeekday:function(t){F=t},getWeekday:function(){return F}};function getTaskTags(t,e,i){let n=!0;for(;n;)n=!1,i.forEach(function(i){let r="^\\s*"+i+"\\s*$",l=new RegExp(r);t[0].match(l)&&(e[i]=!0,t.shift(1),n=!0)})}let N={monday:b.Ox9,tuesday:b.YDX,wednesday:b.EFj,thursday:b.Igq,friday:b.y2j,saturday:b.LqH,sunday:b.Zyz},getMaxIntersections=(t,e)=>{let i=[...t].map(()=>-1/0),n=[...t].sort((t,e)=>t.startTime-e.startTime||t.order-e.order),r=0;for(let t of n)for(let n=0;n<i.length;n++)if(t.startTime>=i[n]){i[n]=t.endTime,t.order=n+e,n>r&&(r=n);break}return r},G={parser:T,db:P,renderer:{setConf:function(){g.l.debug("Something is calling, setConf, remove the call")},draw:function(t,e,i,n){let r;let d=(0,g.c)().gantt,f=(0,g.c)().securityLevel;"sandbox"===f&&(r=(0,b.Ys)("#i"+e));let m="sandbox"===f?(0,b.Ys)(r.nodes()[0].contentDocument.body):(0,b.Ys)("body"),k="sandbox"===f?r.nodes()[0].contentDocument:document,p=k.getElementById(e);void 0===(l=p.parentElement.offsetWidth)&&(l=1200),void 0!==d.useWidth&&(l=d.useWidth);let T=n.db.getTasks(),x=[];for(let t of T)x.push(t.type);x=checkUnique(x);let v={},_=2*d.topPadding;if("compact"===n.db.getDisplayMode()||"compact"===d.displayMode){let t={};for(let e of T)void 0===t[e.section]?t[e.section]=[e]:t[e.section].push(e);let e=0;for(let i of Object.keys(t)){let n=getMaxIntersections(t[i],e)+1;e+=n,_+=n*(d.barHeight+d.barGap),v[i]=n}}else for(let t of(_+=T.length*(d.barHeight+d.barGap),x))v[t]=T.filter(e=>e.type===t).length;p.setAttribute("viewBox","0 0 "+l+" "+_);let w=m.select(`[id="${e}"]`),D=(0,b.Xf)().domain([(0,b.VV$)(T,function(t){return t.startTime}),(0,b.Fp7)(T,function(t){return t.endTime})]).rangeRound([0,l-d.leftPadding-d.rightPadding]);function taskCompare(t,e){let i=t.startTime,n=e.startTime,r=0;return i>n?r=1:i<n&&(r=-1),r}function makeGant(t,e,i){let r=d.barHeight,l=r+d.barGap,f=d.topPadding,y=d.leftPadding,m=(0,b.BYU)().domain([0,x.length]).range(["#00B9FA","#F95002"]).interpolate(b.JHv);drawExcludeDays(l,f,y,e,i,t,n.db.getExcludes(),n.db.getIncludes()),makeGrid(y,f,e,i),drawRects(t,l,f,y,r,m,e),vertLabels(l,f),drawToday(y,f,e,i)}function drawRects(t,i,r,l,f,y,m){let k=[...new Set(t.map(t=>t.order))],p=k.map(e=>t.find(t=>t.order===e));w.append("g").selectAll("rect").data(p).enter().append("rect").attr("x",0).attr("y",function(t,e){return t.order*i+r-2}).attr("width",function(){return m-d.rightPadding/2}).attr("height",i).attr("class",function(t){for(let[e,i]of x.entries())if(t.type===i)return"section section"+e%d.numberSectionStyles;return"section section0"});let T=w.append("g").selectAll("rect").data(t).enter(),v=n.db.getLinks();T.append("rect").attr("id",function(t){return t.id}).attr("rx",3).attr("ry",3).attr("x",function(t){return t.milestone?D(t.startTime)+l+.5*(D(t.endTime)-D(t.startTime))-.5*f:D(t.startTime)+l}).attr("y",function(t,e){return t.order*i+r}).attr("width",function(t){return t.milestone?f:D(t.renderEndTime||t.endTime)-D(t.startTime)}).attr("height",f).attr("transform-origin",function(t,e){return e=t.order,(D(t.startTime)+l+.5*(D(t.endTime)-D(t.startTime))).toString()+"px "+(e*i+r+.5*f).toString()+"px"}).attr("class",function(t){let e="";t.classes.length>0&&(e=t.classes.join(" "));let i=0;for(let[e,n]of x.entries())t.type===n&&(i=e%d.numberSectionStyles);let n="";return t.active?t.crit?n+=" activeCrit":n=" active":t.done?n=t.crit?" doneCrit":" done":t.crit&&(n+=" crit"),0===n.length&&(n=" task"),t.milestone&&(n=" milestone "+n),"task"+(n+=i+" "+e)}),T.append("text").attr("id",function(t){return t.id+"-text"}).text(function(t){return t.task}).attr("font-size",d.fontSize).attr("x",function(t){let e=D(t.startTime),i=D(t.renderEndTime||t.endTime);t.milestone&&(e+=.5*(D(t.endTime)-D(t.startTime))-.5*f),t.milestone&&(i=e+f);let n=this.getBBox().width;return n>i-e?i+n+1.5*d.leftPadding>m?e+l-5:i+l+5:(i-e)/2+e+l}).attr("y",function(t,e){return t.order*i+d.barHeight/2+(d.fontSize/2-2)+r}).attr("text-height",f).attr("class",function(t){let e=D(t.startTime),i=D(t.endTime);t.milestone&&(i=e+f);let n=this.getBBox().width,r="";t.classes.length>0&&(r=t.classes.join(" "));let l=0;for(let[e,i]of x.entries())t.type===i&&(l=e%d.numberSectionStyles);let y="";return(t.active&&(y=t.crit?"activeCritText"+l:"activeText"+l),t.done?y=t.crit?y+" doneCritText"+l:y+" doneText"+l:t.crit&&(y=y+" critText"+l),t.milestone&&(y+=" milestoneText"),n>i-e)?i+n+1.5*d.leftPadding>m?r+" taskTextOutsideLeft taskTextOutside"+l+" "+y:r+" taskTextOutsideRight taskTextOutside"+l+" "+y+" width-"+n:r+" taskText taskText"+l+" "+y+" width-"+n});let _=(0,g.c)().securityLevel;if("sandbox"===_){let t;t=(0,b.Ys)("#i"+e);let i=t.nodes()[0].contentDocument;T.filter(function(t){return void 0!==v[t.id]}).each(function(t){var e=i.querySelector("#"+t.id),n=i.querySelector("#"+t.id+"-text");let r=e.parentNode;var l=i.createElement("a");l.setAttribute("xlink:href",v[t.id]),l.setAttribute("target","_top"),r.appendChild(l),l.appendChild(e),l.appendChild(n)})}}function drawExcludeDays(t,e,i,r,l,f,m,k){let p,b;if(0===m.length&&0===k.length)return;for(let{startTime:t,endTime:e}of f)(void 0===p||t<p)&&(p=t),(void 0===b||e>b)&&(b=e);if(!p||!b)return;if(y(b).diff(y(p),"year")>5){g.l.warn("The difference between the min and max time is more than 5 years. This will cause performance issues. Skipping drawing exclude days.");return}let T=n.db.getDateFormat(),x=[],v=null,_=y(p);for(;_.valueOf()<=b;)n.db.isInvalidDate(_,T,m,k)?v?v.end=_:v={start:_,end:_}:v&&(x.push(v),v=null),_=_.add(1,"d");let $=w.append("g").selectAll("rect").data(x).enter();$.append("rect").attr("id",function(t){return"exclude-"+t.start.format("YYYY-MM-DD")}).attr("x",function(t){return D(t.start)+i}).attr("y",d.gridLineStartPadding).attr("width",function(t){let e=t.end.add(1,"day");return D(e)-D(t.start)}).attr("height",l-e-d.gridLineStartPadding).attr("transform-origin",function(e,n){return(D(e.start)+i+.5*(D(e.end)-D(e.start))).toString()+"px "+(n*t+.5*l).toString()+"px"}).attr("class","exclude-range")}function makeGrid(t,e,i,r){let l=(0,b.LLu)(D).tickSize(-r+e+d.gridLineStartPadding).tickFormat((0,b.i$Z)(n.db.getAxisFormat()||d.axisFormat||"%Y-%m-%d")),f=/^([1-9]\d*)(millisecond|second|minute|hour|day|week|month)$/.exec(n.db.getTickInterval()||d.tickInterval);if(null!==f){let t=f[1],e=f[2],i=n.db.getWeekday()||d.weekday;switch(e){case"millisecond":l.ticks(b.U8T.every(t));break;case"second":l.ticks(b.S1K.every(t));break;case"minute":l.ticks(b.Z_i.every(t));break;case"hour":l.ticks(b.WQD.every(t));break;case"day":l.ticks(b.rr1.every(t));break;case"week":l.ticks(N[i].every(t));break;case"month":l.ticks(b.F0B.every(t))}}if(w.append("g").attr("class","grid").attr("transform","translate("+t+", "+(r-50)+")").call(l).selectAll("text").style("text-anchor","middle").attr("fill","#000").attr("stroke","none").attr("font-size",10).attr("dy","1em"),n.db.topAxisEnabled()||d.topAxis){let i=(0,b.F5q)(D).tickSize(-r+e+d.gridLineStartPadding).tickFormat((0,b.i$Z)(n.db.getAxisFormat()||d.axisFormat||"%Y-%m-%d"));if(null!==f){let t=f[1],e=f[2],r=n.db.getWeekday()||d.weekday;switch(e){case"millisecond":i.ticks(b.U8T.every(t));break;case"second":i.ticks(b.S1K.every(t));break;case"minute":i.ticks(b.Z_i.every(t));break;case"hour":i.ticks(b.WQD.every(t));break;case"day":i.ticks(b.rr1.every(t));break;case"week":i.ticks(N[r].every(t));break;case"month":i.ticks(b.F0B.every(t))}}w.append("g").attr("class","grid").attr("transform","translate("+t+", "+e+")").call(i).selectAll("text").style("text-anchor","middle").attr("fill","#000").attr("stroke","none").attr("font-size",10)}}function vertLabels(t,e){let i=0,n=Object.keys(v).map(t=>[t,v[t]]);w.append("g").selectAll("text").data(n).enter().append(function(t){let e=t[0].split(g.e.lineBreakRegex),i=-(e.length-1)/2,n=k.createElementNS("http://www.w3.org/2000/svg","text");for(let[t,r]of(n.setAttribute("dy",i+"em"),e.entries())){let e=k.createElementNS("http://www.w3.org/2000/svg","tspan");e.setAttribute("alignment-baseline","central"),e.setAttribute("x","10"),t>0&&e.setAttribute("dy","1em"),e.textContent=r,n.appendChild(e)}return n}).attr("x",10).attr("y",function(r,l){if(!(l>0))return r[1]*t/2+e;for(let d=0;d<l;d++)return i+=n[l-1][1],r[1]*t/2+i*t+e}).attr("font-size",d.sectionFontSize).attr("class",function(t){for(let[e,i]of x.entries())if(t[0]===i)return"sectionTitle sectionTitle"+e%d.numberSectionStyles;return"sectionTitle"})}function drawToday(t,e,i,r){let l=n.db.getTodayMarker();if("off"===l)return;let f=w.append("g").attr("class","today"),y=new Date,m=f.append("line");m.attr("x1",D(y)+t).attr("x2",D(y)+t).attr("y1",d.titleTopMargin).attr("y2",r-d.titleTopMargin).attr("class","today"),""!==l&&m.attr("style",l.replace(/,/g,";"))}function checkUnique(t){let e={},i=[];for(let n=0,r=t.length;n<r;++n)Object.prototype.hasOwnProperty.call(e,t[n])||(e[t[n]]=!0,i.push(t[n]));return i}T.sort(taskCompare),makeGant(T,l,_),(0,g.i)(w,_,l,d.useMaxWidth),w.append("text").text(n.db.getDiagramTitle()).attr("x",l/2).attr("y",d.titleTopMargin).attr("class","titleText")}},styles:t=>`
  .mermaid-main-font {
    font-family: "trebuchet ms", verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);
  }
  .exclude-range {
    fill: ${t.excludeBkgColor};
  }

  .section {
    stroke: none;
    opacity: 0.2;
  }

  .section0 {
    fill: ${t.sectionBkgColor};
  }

  .section2 {
    fill: ${t.sectionBkgColor2};
  }

  .section1,
  .section3 {
    fill: ${t.altSectionBkgColor};
    opacity: 0.2;
  }

  .sectionTitle0 {
    fill: ${t.titleColor};
  }

  .sectionTitle1 {
    fill: ${t.titleColor};
  }

  .sectionTitle2 {
    fill: ${t.titleColor};
  }

  .sectionTitle3 {
    fill: ${t.titleColor};
  }

  .sectionTitle {
    text-anchor: start;
    // font-size: ${t.ganttFontSize};
    // text-height: 14px;
    font-family: 'trebuchet ms', verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);

  }


  /* Grid and axis */

  .grid .tick {
    stroke: ${t.gridColor};
    opacity: 0.8;
    shape-rendering: crispEdges;
    text {
      font-family: ${t.fontFamily};
      fill: ${t.textColor};
    }
  }

  .grid path {
    stroke-width: 0;
  }


  /* Today line */

  .today {
    fill: none;
    stroke: ${t.todayLineColor};
    stroke-width: 2px;
  }


  /* Task styling */

  /* Default task */

  .task {
    stroke-width: 2;
  }

  .taskText {
    text-anchor: middle;
    font-family: 'trebuchet ms', verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);
  }

  // .taskText:not([font-size]) {
  //   font-size: ${t.ganttFontSize};
  // }

  .taskTextOutsideRight {
    fill: ${t.taskTextDarkColor};
    text-anchor: start;
    // font-size: ${t.ganttFontSize};
    font-family: 'trebuchet ms', verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);

  }

  .taskTextOutsideLeft {
    fill: ${t.taskTextDarkColor};
    text-anchor: end;
    // font-size: ${t.ganttFontSize};
  }

  /* Special case clickable */
  .task.clickable {
    cursor: pointer;
  }
  .taskText.clickable {
    cursor: pointer;
    fill: ${t.taskTextClickableColor} !important;
    font-weight: bold;
  }

  .taskTextOutsideLeft.clickable {
    cursor: pointer;
    fill: ${t.taskTextClickableColor} !important;
    font-weight: bold;
  }

  .taskTextOutsideRight.clickable {
    cursor: pointer;
    fill: ${t.taskTextClickableColor} !important;
    font-weight: bold;
  }

  /* Specific task settings for the sections*/

  .taskText0,
  .taskText1,
  .taskText2,
  .taskText3 {
    fill: ${t.taskTextColor};
  }

  .task0,
  .task1,
  .task2,
  .task3 {
    fill: ${t.taskBkgColor};
    stroke: ${t.taskBorderColor};
  }

  .taskTextOutside0,
  .taskTextOutside2
  {
    fill: ${t.taskTextOutsideColor};
  }

  .taskTextOutside1,
  .taskTextOutside3 {
    fill: ${t.taskTextOutsideColor};
  }


  /* Active task */

  .active0,
  .active1,
  .active2,
  .active3 {
    fill: ${t.activeTaskBkgColor};
    stroke: ${t.activeTaskBorderColor};
  }

  .activeText0,
  .activeText1,
  .activeText2,
  .activeText3 {
    fill: ${t.taskTextDarkColor} !important;
  }


  /* Completed task */

  .done0,
  .done1,
  .done2,
  .done3 {
    stroke: ${t.doneTaskBorderColor};
    fill: ${t.doneTaskBkgColor};
    stroke-width: 2;
  }

  .doneText0,
  .doneText1,
  .doneText2,
  .doneText3 {
    fill: ${t.taskTextDarkColor} !important;
  }


  /* Tasks on the critical line */

  .crit0,
  .crit1,
  .crit2,
  .crit3 {
    stroke: ${t.critBorderColor};
    fill: ${t.critBkgColor};
    stroke-width: 2;
  }

  .activeCrit0,
  .activeCrit1,
  .activeCrit2,
  .activeCrit3 {
    stroke: ${t.critBorderColor};
    fill: ${t.activeTaskBkgColor};
    stroke-width: 2;
  }

  .doneCrit0,
  .doneCrit1,
  .doneCrit2,
  .doneCrit3 {
    stroke: ${t.critBorderColor};
    fill: ${t.doneTaskBkgColor};
    stroke-width: 2;
    cursor: pointer;
    shape-rendering: crispEdges;
  }

  .milestone {
    transform: rotate(45deg) scale(0.8,0.8);
  }

  .milestoneText {
    font-style: italic;
  }
  .doneCritText0,
  .doneCritText1,
  .doneCritText2,
  .doneCritText3 {
    fill: ${t.taskTextDarkColor} !important;
  }

  .activeCritText0,
  .activeCritText1,
  .activeCritText2,
  .activeCritText3 {
    fill: ${t.taskTextDarkColor} !important;
  }

  .titleText {
    text-anchor: middle;
    font-size: 18px;
    fill: ${t.textColor}    ;
    font-family: 'trebuchet ms', verdana, arial, sans-serif;
    font-family: var(--mermaid-font-family);
  }
`}}}]);