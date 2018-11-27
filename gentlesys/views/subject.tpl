<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
   <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <meta name="description" content="{{.Title}}"/>
   <title>{{.Title}} Powered by Gentlesys!</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
   <style>
   .ul-css{
    	margin-top:0px;
    	margin-bottom:0px;
   }
	.body-css {
		padding-top:55px;
	}
	.root-css {
	 background-color:#fcfdf8;
	}
   </style>
</head>

<body class="body-css">
{{str2html .Navigation}}
<div class="container-fluid">
	  <div class="row root-css col-lg-7 col-lg-offset-1">
      <p class="crumbs"><a href="/">首页</a> &gt;<a href="/article{{.Args}}">[我要发帖]</a></p>

       {{if .ExistNotice}}
       <table class="table table-condensed"> 
       <caption><h5><a href="{{.NoticeRef}}"><strong>版区公告</strong></a></h5></caption>
        <tbody>
         {{range .Notice}}
         <tr>
            <td style="vertical-align:middle;width: 45px">
            <img src="/static/img/n1.png">
            </td>
            <td>
            <p><span><a href="/browse?sid={{$.Nid}}&aid={{.Id}}"><strong>{{.Title}}</strong></a></span><span><small>&nbsp;&nbsp;&nbsp;&nbsp;<a href="/user?name={{.UserName}}">{{.UserName}}</a>&nbsp;&nbsp;•&nbsp;{{.Date}}&nbsp;</small></span></p>
            </td>
         </tr>
         {{end}}
         </tbody>
       </table>
       {{end}}
       
      <ul class="pagination pagination-sm pull-right ul-css">
        <li><a class="btn btn-default" href="/article{{.Args}}" role="button">发帖</a></li>
    	<li><a href="{{.PrePage}}">&laquo;</a></li>
    	 {{range .RecordIndexs}}
        	<li class="{{.IsActive}}"><a href="{{.Ref}}">{{.Title}}</a></li>
    	{{end}}
    	<li><a href="{{.NextPage}}">&raquo;</a></li>
      </ul>

      </div>

      <div class="row root-css col-lg-7 col-lg-offset-1">
      <ul class="nav nav-pills">
      <li><a href="{{.TotalRef}}">全部</a></li>
      <li><a href="{{.HotRef}}">热帖</a></li>
      </ul>
  
	  <table class="table table-condensed"> 
        <tbody>
         {{range .Topic}}
         <tr>
            <td style="vertical-align:middle;width: 45px">
            <img src="/static/img/t2.png">
            </td>
            <td>
            <h5><a href="/browse?sid={{$.Sid}}&aid={{.Id}}"><strong>{{.Title}}</strong></a></h5>
            <small><a href="/user?name={{.UserName}}">{{.UserName}}</a>&nbsp;&nbsp;•&nbsp;{{.Date}}&nbsp;发布&nbsp;&nbsp;•&nbsp;[{{.ReadTimes}}/{{.ReplyTimes}}]</small>
            </td>
         </tr>
         {{end}}
         </tbody>
       </table>
        
       {{if .NoMore}}
        <hr />
        <p>没有更多帖子了...</p>
       {{end}}

       <ul class="pagination pagination-sm ul-css">
        <li><a class="btn btn-default" href="/article{{.Args}}" role="button">发帖</a></li>
    	<li><a href="{{.PrePage}}">&laquo;</a></li>
    	 {{range .RecordIndexs}}
        	<li class="{{.IsActive}}"><a href="{{.Ref}}">{{.Title}}</a></li>
    	{{end}}
    	<li><a href="{{.NextPage}}">&raquo;</a></li>
      </ul>
	  </div>

</div>
</body>

</html>

