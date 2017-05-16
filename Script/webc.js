function $ID(id)
{
	return document.getElementById(id);
}
var skey,sKeyReq;
function init(sesskey)
{
	skey=sesskey;
	document.body.onload=function(){
		try{
			sKeyReq=hex_md5(sKey);
		}catch(e){
			return;
		}
		$ID("sKey").value=sKeyReq;
		$ID("form1").onsubmit=OnSubmit;
		Build();
	} 
}
function getCookie(c_name)
{
if (document.cookie.length>0)
  {
  c_start=document.cookie.indexOf(c_name + "=")
  if (c_start!=-1)
    { 
    c_start=c_start + c_name.length+1 
    c_end=document.cookie.indexOf(";",c_start)
    if (c_end==-1) c_end=document.cookie.length
    return unescape(document.cookie.substring(c_start,c_end))
    } 
  }
return ""
}

function OnSubmit()
{
	setTimeout(function(){$ID("cmd").value="";},300);
}

function Build() {
	pref = "";
	/*if(document.location.port)
	{
		pref=":"+document.location.port;
	}*/
    ws = new WebSocket('ws://'+document.location.host+pref+'/ws');

    ws.onopen = OnOpen;
    ws.onmessage = OnMessage;
}

function OnOpen(event) {
    ws.send(sKeyReq);
}

function OnMessage(event) {
    $ID("stdout").value += event.data;
};