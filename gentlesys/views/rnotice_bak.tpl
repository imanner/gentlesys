<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
   <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <meta name="description" content="Gentlesys {{.Title}}"/>
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
   <script src="/static/bootsp/edt/kindeditor-all-min.js"></script>
   <script src="/static/bootsp/edt/lang/zh-CN.js"></script>

   <style>
   	.h5 {
   	color:#009900;
   	}
	h4 {
    color:#4876FF;
    }
	.table-css {
		text-align:center;
	    color:#000033;
	}
	.key-prob {
		color:#0000CC;
		padding-right:2px;
	}
	.title-css {
	float: left;
	}
	.comment-css {
	background-color:#e0ffff;
	}
	.body-css {
	 background-color:#fcfdf8;
	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Navigation}}
<div class="container-fluid">
<div class="row">

<div class="col-md-6 col-md-offset-1">
    <p class="crumbs"><a href="/">首页</a> &gt;  <a href="{{.HrefSub}}">{{.SubName}}</a>&gt;<a href="/article{{.Args}}">[我要发帖]</a></p>
    <h4 class="text-center">{{.Title}}</h4>
    <span class = "key-prob">作者:{{.UserName}}</span>
    <span class = "key-prob">{{.Date}}</span></p>

    <div id="story" class="body-css">
        {{str2html .Story}}
        <h5 class="page-header"></h5>
    </div>
</div>
</div>
</div>
</body>
</html>
