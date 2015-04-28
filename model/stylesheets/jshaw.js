var postCardH = 138;
var cardMargin = 42;
var gRatio = 0.618;
var doubleCheck = 0;

function ResizeNavbar() {
    var y = Math.max(window.pageYOffset,0);
    var navH = Math.max(postCardH-y,cardMargin);
    if (navH == cardMargin)
        document.getElementById("navbar").style.opacity=0.98;
    else
        document.getElementById("navbar").style.opacity=0.9;
    if (navH==doubleCheck) return;//reduce the cpu consumption
    doubleCheck=navH;
    var content=navH.toString()+"px";
    //Change background
    document.getElementById("navbar").style.height = content;
    //Make text vertical-align
    var change = document.getElementsByClassName("nav-item");
    for (i=0; i < change.length; ++i) {
        change[i].style.lineHeight = content;
    }
    //Change icon size
    change = document.getElementById("brand_icon");
    var gheight = navH*gRatio;
    change.style.height = gheight.toString()+"px";
    change.style.width = gheight.toString()+"px";
    //Brand Text Animation
    if (gheight<cardMargin)
        document.getElementById("brand_text").style.opacity=0;
    else
        document.getElementById("brand_text").style.opacity=0.8;
}

function GetPageNumber(url) {
    var pos = url.lastIndexOf("num=");
    if (pos==-1) {
        return 1;
    } else {
        var str = url.substring(pos+"num=".length,url.length);
        var number = parseInt(str);
        if (isNaN(number)) {
            return 1;
        }
        return number;
    }
}

function OnloadPosts() {
    var item = document.getElementById("prev");
    var number = GetPageNumber(window.location.search);
    //if (number == 1)
      //  item.style.visibility="hidden";//the 1st page
    item.setAttribute("href","/?page="+(number-1).toString());
    item = document.getElementById("next");
    if (item != null)
        item.setAttribute("href","/?page="+(number+1).toString());
}
function OnloadArticle() {
    var height = document.getElementById("content").offsetHeight.toString()+"px";
    document.getElementById("art-bkground").style.height=height;
    document.getElementById("art-container").style.top="-"+height;
}

function OnloadPage() {
    var url = location.href;
    var pos = url.lastIndexOf("article?key=");
    if (pos != -1) {
        SendMessage(url.substring(pos+"article?key=".length,url.length));
        return;
    }
    GetPostPage(-1);
}
/*this function has been abanded due to the partical update cannot use backward & forward*/
function GetPostPage(number) {
    if (number<0) number = GetPageNumber(location.href);
    xmlhttp=new XMLHttpRequest();
    xmlhttp.onreadystatechange=function()
    {
        if (xmlhttp.readyState==4 && xmlhttp.status==200)
        {
            setTimeout(function(change,content,number) {
                       change.innerHTML=content;
                       ResizeNavbar();
                       change.style.opacity = 1;
                       var stateObj = { num : number };
                       history.pushState(stateObj, "", "/page?num=" + number.toString());
                       var item = document.getElementById("prev");
                       if (number!=1) {
                            item.style.cursor = "pointer";
                            item.setAttribute("onclick","GetPostPage("+(number-1).toString()+");");
                       } else item.style.visibility="hidden";//the 1st page
                       item = document.getElementById("next");
                       if (item != null) {
                            item.setAttribute("onclick","GetPostPage("+(number+1).toString()+");");
                            item.style.cursor = "pointer";
                       }
                       },200,change,xmlhttp.responseText,number);//200ms should be equal with .css #playground transition duration
        }
    }
    xmlhttp.open("POST","/page?num="+number.toString(),true);
    xmlhttp.send();
    change = document.getElementById("playground");
    change.style.opacity = 0;
}
